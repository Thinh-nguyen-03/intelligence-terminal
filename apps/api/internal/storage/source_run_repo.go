package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/0510t/intelligence-terminal/apps/api/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

// SourceRunRepo handles persistence for job run tracking.
type SourceRunRepo struct {
	pool *pgxpool.Pool
}

func NewSourceRunRepo(pool *pgxpool.Pool) *SourceRunRepo {
	return &SourceRunRepo{pool: pool}
}

// Start creates a new source run record in 'running' status and returns its ID.
func (r *SourceRunRepo) Start(ctx context.Context, source domain.Source, jobName string) (int64, error) {
	var id int64
	err := r.pool.QueryRow(ctx, `
		INSERT INTO source_runs (source, job_name, status, started_at)
		VALUES ($1, $2, 'running', $3)
		RETURNING id
	`, source, jobName, time.Now().UTC()).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("starting source run: %w", err)
	}
	return id, nil
}

// Complete marks a run as successful.
func (r *SourceRunRepo) Complete(ctx context.Context, runID int64, recordsProcessed int, checksum *string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE source_runs
		SET status = 'success', finished_at = $2, records_processed = $3, checksum = $4
		WHERE id = $1
	`, runID, time.Now().UTC(), recordsProcessed, checksum)
	if err != nil {
		return fmt.Errorf("completing source run: %w", err)
	}
	return nil
}

// Fail marks a run as failed with an error message.
func (r *SourceRunRepo) Fail(ctx context.Context, runID int64, errMsg string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE source_runs
		SET status = 'failed', finished_at = $2, error_message = $3
		WHERE id = $1
	`, runID, time.Now().UTC(), errMsg)
	if err != nil {
		return fmt.Errorf("failing source run: %w", err)
	}
	return nil
}

// GetLatestSuccessful returns the most recent successful run for a given source and job.
func (r *SourceRunRepo) GetLatestSuccessful(ctx context.Context, source domain.Source, jobName string) (*domain.SourceRun, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, source, job_name, status, started_at, finished_at, records_processed, error_message, checksum
		FROM source_runs
		WHERE source = $1 AND job_name = $2 AND status = 'success'
		ORDER BY finished_at DESC
		LIMIT 1
	`, source, jobName)

	var sr domain.SourceRun
	err := row.Scan(&sr.ID, &sr.Source, &sr.JobName, &sr.Status, &sr.StartedAt, &sr.FinishedAt, &sr.RecordsProcessed, &sr.ErrorMessage, &sr.Checksum)
	if err != nil {
		return nil, nil // no previous successful run
	}
	return &sr, nil
}
