package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/0510t/intelligence-terminal/apps/api/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// AlertRepo handles persistence for generated alerts.
type AlertRepo struct {
	pool *pgxpool.Pool
}

func NewAlertRepo(pool *pgxpool.Pool) *AlertRepo {
	return &AlertRepo{pool: pool}
}

// InsertAlert stores a new alert.
func (r *AlertRepo) InsertAlert(ctx context.Context, a *domain.Alert) (int64, error) {
	var id int64
	err := r.pool.QueryRow(ctx, `
		INSERT INTO alerts (
			commodity_id, as_of_date, severity, alert_type, headline, summary,
			explanation_json, regime_label, regime_confidence, final_alert_score, is_active
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id
	`, a.CommodityID, a.AsOfDate, a.Severity, a.AlertType, a.Headline, a.Summary,
		a.ExplanationJSON, a.RegimeLabel, a.RegimeConfidence, a.FinalAlertScore, a.IsActive).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("inserting alert: %w", err)
	}
	return id, nil
}

// DeactivateOlderAlerts marks all alerts before the given date as inactive.
func (r *AlertRepo) DeactivateOlderAlerts(ctx context.Context, beforeDate time.Time) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE alerts SET is_active = false
		WHERE as_of_date < $1 AND is_active = true
	`, beforeDate)
	if err != nil {
		return fmt.Errorf("deactivating old alerts: %w", err)
	}
	return nil
}

// ListActiveAlerts returns active alerts, optionally filtered by severity and/or commodity slug.
func (r *AlertRepo) ListActiveAlerts(ctx context.Context, severity *string, commoditySlug *string) ([]domain.Alert, error) {
	query := `
		SELECT a.id, a.commodity_id, a.as_of_date, a.severity, a.alert_type, a.headline, a.summary,
			   a.explanation_json, a.regime_label, a.regime_confidence, a.final_alert_score, a.is_active, a.created_at
		FROM alerts a
	`
	args := []any{}
	where := "WHERE a.is_active = true"
	argIdx := 1

	if commoditySlug != nil {
		query += " JOIN commodities c ON c.id = a.commodity_id"
		where += fmt.Sprintf(" AND c.slug = $%d", argIdx)
		args = append(args, *commoditySlug)
		argIdx++
	}

	if severity != nil {
		where += fmt.Sprintf(" AND a.severity = $%d", argIdx)
		args = append(args, *severity)
		argIdx++
	}

	query += " " + where + " ORDER BY a.final_alert_score DESC, a.as_of_date DESC"

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("querying active alerts: %w", err)
	}
	defer rows.Close()

	return scanAlerts(rows)
}

// GetAlertByID returns a single alert by ID.
func (r *AlertRepo) GetAlertByID(ctx context.Context, id int64) (*domain.Alert, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, commodity_id, as_of_date, severity, alert_type, headline, summary,
			   explanation_json, regime_label, regime_confidence, final_alert_score, is_active, created_at
		FROM alerts WHERE id = $1
	`, id)

	var a domain.Alert
	err := row.Scan(&a.ID, &a.CommodityID, &a.AsOfDate, &a.Severity, &a.AlertType, &a.Headline, &a.Summary,
		&a.ExplanationJSON, &a.RegimeLabel, &a.RegimeConfidence, &a.FinalAlertScore, &a.IsActive, &a.CreatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("querying alert: %w", err)
	}
	return &a, nil
}

func scanAlerts(rows pgx.Rows) ([]domain.Alert, error) {
	var alerts []domain.Alert
	for rows.Next() {
		var a domain.Alert
		if err := rows.Scan(&a.ID, &a.CommodityID, &a.AsOfDate, &a.Severity, &a.AlertType, &a.Headline, &a.Summary,
			&a.ExplanationJSON, &a.RegimeLabel, &a.RegimeConfidence, &a.FinalAlertScore, &a.IsActive, &a.CreatedAt); err != nil {
			return nil, fmt.Errorf("scanning alert: %w", err)
		}
		alerts = append(alerts, a)
	}
	return alerts, rows.Err()
}
