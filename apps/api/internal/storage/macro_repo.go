package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/0510t/intelligence-terminal/apps/api/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// MacroRepo handles persistence for macro series and observations.
type MacroRepo struct {
	pool *pgxpool.Pool
}

func NewMacroRepo(pool *pgxpool.Pool) *MacroRepo {
	return &MacroRepo{pool: pool}
}

// ListEnabledSeries returns all macro series that are enabled for ingestion.
func (r *MacroRepo) ListEnabledSeries(ctx context.Context) ([]domain.MacroSeries, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, source, source_series_id, slug, name, frequency, units, enabled, created_at
		FROM macro_series
		WHERE enabled = true
		ORDER BY id
	`)
	if err != nil {
		return nil, fmt.Errorf("querying enabled series: %w", err)
	}
	defer rows.Close()

	var series []domain.MacroSeries
	for rows.Next() {
		var s domain.MacroSeries
		if err := rows.Scan(&s.ID, &s.Source, &s.SourceSeriesID, &s.Slug, &s.Name, &s.Frequency, &s.Units, &s.Enabled, &s.CreatedAt); err != nil {
			return nil, fmt.Errorf("scanning series: %w", err)
		}
		series = append(series, s)
	}
	return series, rows.Err()
}

// GetLatestObservationDate returns the most recent observation date for a series, or nil if none exist.
func (r *MacroRepo) GetLatestObservationDate(ctx context.Context, seriesID int64) (*time.Time, error) {
	var d time.Time
	err := r.pool.QueryRow(ctx, `
		SELECT MAX(observation_date)
		FROM macro_observations_clean
		WHERE series_id = $1
	`, seriesID).Scan(&d)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("querying latest observation: %w", err)
	}
	if d.IsZero() {
		return nil, nil
	}
	return &d, nil
}

// InsertRawObservation inserts a raw FRED/ALFRED observation.
func (r *MacroRepo) InsertRawObservation(ctx context.Context, obs *domain.MacroObservationRaw) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO macro_observations_raw (series_id, observation_date, realtime_start, realtime_end, raw_value, payload_json, ingested_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, obs.SeriesID, obs.ObservationDate, obs.RealtimeStart, obs.RealtimeEnd, obs.RawValue, obs.PayloadJSON, obs.IngestedAt)
	if err != nil {
		return fmt.Errorf("inserting raw observation: %w", err)
	}
	return nil
}

// UpsertCleanObservation inserts or updates a clean observation.
// Uses series_id + observation_date as the natural key for non-vintage data.
func (r *MacroRepo) UpsertCleanObservation(ctx context.Context, obs *domain.MacroObservationClean) error {
	if obs.VintageDate != nil {
		_, err := r.pool.Exec(ctx, `
			INSERT INTO macro_observations_clean (series_id, observation_date, value, vintage_date, is_latest, frequency)
			VALUES ($1, $2, $3, $4, $5, $6)
			ON CONFLICT (series_id, observation_date, vintage_date) WHERE vintage_date IS NOT NULL
			DO UPDATE SET value = EXCLUDED.value, is_latest = EXCLUDED.is_latest
		`, obs.SeriesID, obs.ObservationDate, obs.Value, obs.VintageDate, obs.IsLatest, obs.Frequency)
		if err != nil {
			return fmt.Errorf("upserting vintage observation: %w", err)
		}
	} else {
		_, err := r.pool.Exec(ctx, `
			INSERT INTO macro_observations_clean (series_id, observation_date, value, vintage_date, is_latest, frequency)
			VALUES ($1, $2, $3, NULL, $4, $5)
			ON CONFLICT (series_id, observation_date) WHERE vintage_date IS NULL
			DO UPDATE SET value = EXCLUDED.value, is_latest = EXCLUDED.is_latest
		`, obs.SeriesID, obs.ObservationDate, obs.Value, obs.IsLatest, obs.Frequency)
		if err != nil {
			return fmt.Errorf("upserting observation: %w", err)
		}
	}
	return nil
}

// ListCleanObservations returns clean observations for a series within a date range, ordered ascending.
func (r *MacroRepo) ListCleanObservations(ctx context.Context, seriesID int64, from, to time.Time) ([]domain.MacroObservationClean, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, series_id, observation_date, value, vintage_date, is_latest, frequency
		FROM macro_observations_clean
		WHERE series_id = $1 AND observation_date >= $2 AND observation_date <= $3
		  AND vintage_date IS NULL
		ORDER BY observation_date ASC
	`, seriesID, from, to)
	if err != nil {
		return nil, fmt.Errorf("querying observations: %w", err)
	}
	defer rows.Close()

	var obs []domain.MacroObservationClean
	for rows.Next() {
		var o domain.MacroObservationClean
		if err := rows.Scan(&o.ID, &o.SeriesID, &o.ObservationDate, &o.Value, &o.VintageDate, &o.IsLatest, &o.Frequency); err != nil {
			return nil, fmt.Errorf("scanning observation: %w", err)
		}
		obs = append(obs, o)
	}
	return obs, rows.Err()
}

// ExecRaw executes a raw SQL statement. Used for bulk updates like is_latest flag management.
func (r *MacroRepo) ExecRaw(ctx context.Context, sql string, args ...any) (int64, error) {
	tag, err := r.pool.Exec(ctx, sql, args...)
	if err != nil {
		return 0, fmt.Errorf("exec raw: %w", err)
	}
	return tag.RowsAffected(), nil
}
