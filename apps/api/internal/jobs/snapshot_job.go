package jobs

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/0510t/intelligence-terminal/apps/api/internal/analytics"
	"github.com/0510t/intelligence-terminal/apps/api/internal/domain"
	"github.com/0510t/intelligence-terminal/apps/api/internal/storage"
)

// SnapshotJob computes macro factor snapshots from ingested observations.
type SnapshotJob struct {
	macroRepo    *storage.MacroRepo
	snapshotRepo *storage.SnapshotRepo
	configRepo   *storage.ConfigRepo
	sourceRunRepo *storage.SourceRunRepo
}

func NewSnapshotJob(
	macroRepo *storage.MacroRepo,
	snapshotRepo *storage.SnapshotRepo,
	configRepo *storage.ConfigRepo,
	sourceRunRepo *storage.SourceRunRepo,
) *SnapshotJob {
	return &SnapshotJob{
		macroRepo:     macroRepo,
		snapshotRepo:  snapshotRepo,
		configRepo:    configRepo,
		sourceRunRepo: sourceRunRepo,
	}
}

// Run computes a factor snapshot for the current date (or a specific date if provided).
func (j *SnapshotJob) Run(ctx context.Context) error {
	return j.RunForDate(ctx, time.Now().UTC())
}

// RunForDate computes a factor snapshot as of a given date.
func (j *SnapshotJob) RunForDate(ctx context.Context, asOf time.Time) error {
	runID, err := j.sourceRunRepo.Start(ctx, domain.SourceAnalytics, "rebuild-snapshots")
	if err != nil {
		return fmt.Errorf("starting run: %w", err)
	}

	if err := j.computeSnapshot(ctx, asOf); err != nil {
		if failErr := j.sourceRunRepo.Fail(ctx, runID, err.Error()); failErr != nil {
			slog.Error("failed to mark run as failed", "error", failErr)
		}
		return err
	}

	if err := j.sourceRunRepo.Complete(ctx, runID, 1, nil); err != nil {
		return fmt.Errorf("completing run: %w", err)
	}

	slog.Info("snapshot computation complete", "as_of", asOf.Format("2006-01-02"))
	return nil
}

func (j *SnapshotJob) computeSnapshot(ctx context.Context, asOf time.Time) error {
	// Load model parameters
	params, err := j.configRepo.LoadModelParams(ctx)
	if err != nil {
		return fmt.Errorf("loading model params: %w", err)
	}

	// Load all enabled series
	allSeries, err := j.macroRepo.ListEnabledSeries(ctx)
	if err != nil {
		return fmt.Errorf("listing series: %w", err)
	}

	// Build slug -> series map
	slugToSeries := make(map[string]domain.MacroSeries)
	for _, s := range allSeries {
		slugToSeries[s.Slug] = s
	}

	// Load observations for each series (2 years of data for z-score calculation)
	lookbackStart := asOf.AddDate(-2, 0, 0)
	seriesData := make(analytics.SeriesMap)

	for _, s := range allSeries {
		obs, err := j.macroRepo.ListCleanObservations(ctx, s.ID, lookbackStart, asOf)
		if err != nil {
			slog.Warn("failed to load observations", "slug", s.Slug, "error", err)
			continue
		}
		analytics.SortObservations(obs)
		seriesData[s.Slug] = obs
		slog.Debug("loaded series data", "slug", s.Slug, "count", len(obs))
	}

	// Compute factor scores
	inflationScore := analytics.ComputeInflationScore(seriesData)
	growthScore := analytics.ComputeGrowthScore(seriesData)
	laborScore := analytics.ComputeLaborScore(seriesData)
	stressScore := analytics.ComputeStressScore(seriesData)

	slog.Info("computed factor scores",
		"growth", fmt.Sprintf("%.3f", growthScore),
		"inflation", fmt.Sprintf("%.3f", inflationScore),
		"labor", fmt.Sprintf("%.3f", laborScore),
		"stress", fmt.Sprintf("%.3f", stressScore))

	// Classify regime
	regime := analytics.ClassifyRegime(growthScore, inflationScore, stressScore, params)

	// Get prior regime label for tracking transitions
	var priorLabel *string
	prior, err := j.snapshotRepo.GetLatestSnapshot(ctx, params.ModelVersion)
	if err != nil {
		slog.Warn("failed to get prior snapshot", "error", err)
	}
	if prior != nil {
		priorLabel = &prior.RegimeLabel
	}

	// Build and store snapshot
	snapshot := &domain.MacroFactorSnapshot{
		AsOfDate:         asOf,
		GrowthScore:      growthScore,
		InflationScore:   inflationScore,
		LaborScore:       laborScore,
		StressScore:      stressScore,
		RegimeLabel:      regime.Label,
		RegimePriorLabel: priorLabel,
		Confidence:       regime.Confidence,
		IsTransitioning:  regime.IsTransitioning,
		ModelVersion:     params.ModelVersion,
	}

	if regime.TransitionDetail != "" {
		snapshot.TransitionDetail = &regime.TransitionDetail
	}

	if err := j.snapshotRepo.UpsertFactorSnapshot(ctx, snapshot); err != nil {
		return fmt.Errorf("storing snapshot: %w", err)
	}

	slog.Info("regime classified",
		"label", regime.Label,
		"confidence", regime.Confidence,
		"transitioning", regime.IsTransitioning)

	return nil
}
