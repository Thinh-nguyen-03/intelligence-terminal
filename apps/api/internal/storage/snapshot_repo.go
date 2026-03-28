package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/0510t/intelligence-terminal/apps/api/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// SnapshotRepo handles persistence for macro factor snapshots.
type SnapshotRepo struct {
	pool *pgxpool.Pool
}

func NewSnapshotRepo(pool *pgxpool.Pool) *SnapshotRepo {
	return &SnapshotRepo{pool: pool}
}

// UpsertFactorSnapshot inserts or updates a macro factor snapshot.
func (r *SnapshotRepo) UpsertFactorSnapshot(ctx context.Context, s *domain.MacroFactorSnapshot) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO macro_factor_snapshots (
			as_of_date, growth_score, inflation_score, labor_score, stress_score,
			regime_label, regime_prior_label, confidence,
			is_transitioning, transition_detail, model_version
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		ON CONFLICT (as_of_date, model_version)
		DO UPDATE SET
			growth_score = EXCLUDED.growth_score,
			inflation_score = EXCLUDED.inflation_score,
			labor_score = EXCLUDED.labor_score,
			stress_score = EXCLUDED.stress_score,
			regime_label = EXCLUDED.regime_label,
			regime_prior_label = EXCLUDED.regime_prior_label,
			confidence = EXCLUDED.confidence,
			is_transitioning = EXCLUDED.is_transitioning,
			transition_detail = EXCLUDED.transition_detail,
			created_at = now()
	`, s.AsOfDate, s.GrowthScore, s.InflationScore, s.LaborScore, s.StressScore,
		s.RegimeLabel, s.RegimePriorLabel, s.Confidence,
		s.IsTransitioning, s.TransitionDetail, s.ModelVersion)
	if err != nil {
		return fmt.Errorf("upserting factor snapshot: %w", err)
	}
	return nil
}

// GetLatestSnapshot returns the most recent factor snapshot, or nil if none exist.
func (r *SnapshotRepo) GetLatestSnapshot(ctx context.Context, modelVersion string) (*domain.MacroFactorSnapshot, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, as_of_date, growth_score, inflation_score, labor_score, stress_score,
			   regime_label, regime_prior_label, confidence,
			   is_transitioning, transition_detail, model_version, created_at
		FROM macro_factor_snapshots
		WHERE model_version = $1
		ORDER BY as_of_date DESC
		LIMIT 1
	`, modelVersion)

	var s domain.MacroFactorSnapshot
	err := row.Scan(&s.ID, &s.AsOfDate, &s.GrowthScore, &s.InflationScore, &s.LaborScore, &s.StressScore,
		&s.RegimeLabel, &s.RegimePriorLabel, &s.Confidence,
		&s.IsTransitioning, &s.TransitionDetail, &s.ModelVersion, &s.CreatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("querying latest snapshot: %w", err)
	}
	return &s, nil
}

// GetSnapshotByDate returns the snapshot for a specific date, or nil if not found.
func (r *SnapshotRepo) GetSnapshotByDate(ctx context.Context, asOfDate time.Time, modelVersion string) (*domain.MacroFactorSnapshot, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, as_of_date, growth_score, inflation_score, labor_score, stress_score,
			   regime_label, regime_prior_label, confidence,
			   is_transitioning, transition_detail, model_version, created_at
		FROM macro_factor_snapshots
		WHERE as_of_date = $1 AND model_version = $2
	`, asOfDate, modelVersion)

	var s domain.MacroFactorSnapshot
	err := row.Scan(&s.ID, &s.AsOfDate, &s.GrowthScore, &s.InflationScore, &s.LaborScore, &s.StressScore,
		&s.RegimeLabel, &s.RegimePriorLabel, &s.Confidence,
		&s.IsTransitioning, &s.TransitionDetail, &s.ModelVersion, &s.CreatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("querying snapshot by date: %w", err)
	}
	return &s, nil
}

// ListSnapshots returns snapshots within a date range, ordered descending.
func (r *SnapshotRepo) ListSnapshots(ctx context.Context, from, to time.Time, modelVersion string) ([]domain.MacroFactorSnapshot, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, as_of_date, growth_score, inflation_score, labor_score, stress_score,
			   regime_label, regime_prior_label, confidence,
			   is_transitioning, transition_detail, model_version, created_at
		FROM macro_factor_snapshots
		WHERE as_of_date >= $1 AND as_of_date <= $2 AND model_version = $3
		ORDER BY as_of_date DESC
	`, from, to, modelVersion)
	if err != nil {
		return nil, fmt.Errorf("querying snapshots: %w", err)
	}
	defer rows.Close()

	var snapshots []domain.MacroFactorSnapshot
	for rows.Next() {
		var s domain.MacroFactorSnapshot
		if err := rows.Scan(&s.ID, &s.AsOfDate, &s.GrowthScore, &s.InflationScore, &s.LaborScore, &s.StressScore,
			&s.RegimeLabel, &s.RegimePriorLabel, &s.Confidence,
			&s.IsTransitioning, &s.TransitionDetail, &s.ModelVersion, &s.CreatedAt); err != nil {
			return nil, fmt.Errorf("scanning snapshot: %w", err)
		}
		snapshots = append(snapshots, s)
	}
	return snapshots, rows.Err()
}
