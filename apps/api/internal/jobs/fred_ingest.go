package jobs

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/0510t/intelligence-terminal/apps/api/internal/domain"
	"github.com/0510t/intelligence-terminal/apps/api/internal/storage"
)

// FREDIngestJob fetches and stores FRED macro observations.
type FREDIngestJob struct {
	client       *FREDClient
	macroRepo    *storage.MacroRepo
	sourceRunRepo *storage.SourceRunRepo
}

func NewFREDIngestJob(client *FREDClient, macroRepo *storage.MacroRepo, sourceRunRepo *storage.SourceRunRepo) *FREDIngestJob {
	return &FREDIngestJob{
		client:        client,
		macroRepo:     macroRepo,
		sourceRunRepo: sourceRunRepo,
	}
}

// Run executes the FRED ingestion for all enabled series.
// It fetches new observations since the last known date (or the full lookback window for first run).
func (j *FREDIngestJob) Run(ctx context.Context, lookbackYears int) error {
	runID, err := j.sourceRunRepo.Start(ctx, domain.SourceFRED, "ingest-macro")
	if err != nil {
		return fmt.Errorf("starting run: %w", err)
	}

	totalRecords, jobErr := j.run(ctx, lookbackYears)
	if jobErr != nil {
		if failErr := j.sourceRunRepo.Fail(ctx, runID, jobErr.Error()); failErr != nil {
			slog.Error("failed to mark run as failed", "error", failErr)
		}
		return jobErr
	}

	if err := j.sourceRunRepo.Complete(ctx, runID, totalRecords, nil); err != nil {
		return fmt.Errorf("completing run: %w", err)
	}

	slog.Info("FRED ingestion complete", "total_records", totalRecords)
	return nil
}

func (j *FREDIngestJob) run(ctx context.Context, lookbackYears int) (int, error) {
	series, err := j.macroRepo.ListEnabledSeries(ctx)
	if err != nil {
		return 0, fmt.Errorf("listing enabled series: %w", err)
	}

	if len(series) == 0 {
		slog.Warn("no enabled macro series found")
		return 0, nil
	}

	totalRecords := 0
	now := time.Now().UTC()
	defaultStart := now.AddDate(-lookbackYears, 0, 0)

	for _, s := range series {
		if ctx.Err() != nil {
			return totalRecords, ctx.Err()
		}

		count, err := j.ingestSeries(ctx, s, defaultStart, now)
		if err != nil {
			slog.Error("failed to ingest series", "slug", s.Slug, "error", err)
			continue // partial failure tolerated
		}
		totalRecords += count
		slog.Info("ingested series", "slug", s.Slug, "records", count)

		// Throttle to stay within FRED rate limit (120 req/min)
		time.Sleep(600 * time.Millisecond)
	}

	return totalRecords, nil
}

func (j *FREDIngestJob) ingestSeries(ctx context.Context, series domain.MacroSeries, defaultStart, endDate time.Time) (int, error) {
	// Determine start date: either after the last known observation or the default lookback
	startDate := defaultStart
	latest, err := j.macroRepo.GetLatestObservationDate(ctx, series.ID)
	if err != nil {
		return 0, fmt.Errorf("getting latest date for %s: %w", series.Slug, err)
	}
	if latest != nil {
		// Fetch from one day after the latest known observation
		startDate = latest.AddDate(0, 0, 1)
	}

	if startDate.After(endDate) {
		slog.Info("series already up to date", "slug", series.Slug)
		return 0, nil
	}

	resp, rawBody, err := j.client.FetchObservations(ctx, series.SourceSeriesID, startDate, endDate)
	if err != nil {
		return 0, fmt.Errorf("fetching %s: %w", series.Slug, err)
	}

	count := 0
	for _, obs := range resp.Observations {
		if obs.Value == "." {
			// FRED uses "." for missing values
			continue
		}

		obsDate, err := time.Parse("2006-01-02", obs.Date)
		if err != nil {
			slog.Warn("skipping observation with invalid date", "series", series.Slug, "date", obs.Date)
			continue
		}

		value, err := strconv.ParseFloat(obs.Value, 64)
		if err != nil {
			slog.Warn("skipping observation with invalid value", "series", series.Slug, "value", obs.Value)
			continue
		}

		// Store raw observation
		rawObs := &domain.MacroObservationRaw{
			SeriesID:        series.ID,
			ObservationDate: obsDate,
			RawValue:        obs.Value,
			PayloadJSON:     rawBody,
			IngestedAt:      time.Now().UTC(),
		}

		// Only store payload_json for the first observation per batch to save space
		if count > 0 {
			rawObs.PayloadJSON = nil
		}

		if err := j.macroRepo.InsertRawObservation(ctx, rawObs); err != nil {
			slog.Warn("skipping duplicate raw observation", "series", series.Slug, "date", obs.Date)
			// Continue on duplicate — idempotent behavior
		}

		// Store clean observation
		cleanObs := &domain.MacroObservationClean{
			SeriesID:        series.ID,
			ObservationDate: obsDate,
			Value:           value,
			IsLatest:        true,
			Frequency:       series.Frequency,
		}
		if err := j.macroRepo.UpsertCleanObservation(ctx, cleanObs); err != nil {
			return count, fmt.Errorf("upserting clean observation for %s on %s: %w", series.Slug, obs.Date, err)
		}

		count++
	}

	// Update is_latest flags: only the most recent observation per series should be marked latest
	if count > 0 {
		if err := j.updateIsLatest(ctx, series.ID); err != nil {
			slog.Error("failed to update is_latest flags", "series", series.Slug, "error", err)
		}
	}

	// Log raw payload size for storage monitoring
	slog.Info("raw payload size", "series", series.Slug, "bytes", len(rawBody))

	return count, nil
}

func (j *FREDIngestJob) updateIsLatest(ctx context.Context, seriesID int64) error {
	// Mark all as not latest, then mark only the most recent
	_, err := j.macroRepo.ExecRaw(ctx, `
		UPDATE macro_observations_clean SET is_latest = false WHERE series_id = $1 AND is_latest = true
	`, seriesID)
	if err != nil {
		return err
	}

	_, err = j.macroRepo.ExecRaw(ctx, `
		UPDATE macro_observations_clean SET is_latest = true
		WHERE series_id = $1 AND vintage_date IS NULL
		  AND observation_date = (
		    SELECT MAX(observation_date) FROM macro_observations_clean WHERE series_id = $1 AND vintage_date IS NULL
		  )
	`, seriesID)
	return err
}

// RunSingle ingests a single series by slug. Useful for targeted backfills.
func (j *FREDIngestJob) RunSingle(ctx context.Context, slug string, startDate, endDate time.Time) error {
	series, err := j.macroRepo.ListEnabledSeries(ctx)
	if err != nil {
		return err
	}

	for _, s := range series {
		if s.Slug == slug {
			count, err := j.ingestSeries(ctx, s, startDate, endDate)
			if err != nil {
				return err
			}
			slog.Info("single series ingestion complete", "slug", slug, "records", count)
			return nil
		}
	}

	return fmt.Errorf("series %q not found or not enabled", slug)
}

// marshalObs is a helper to marshal a single FRED observation for raw storage.
func marshalObs(obs FREDObservation) json.RawMessage {
	b, _ := json.Marshal(obs)
	return b
}
