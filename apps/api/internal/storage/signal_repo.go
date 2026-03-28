package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/0510t/intelligence-terminal/apps/api/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// SignalRepo handles persistence for commodity signal snapshots.
type SignalRepo struct {
	pool *pgxpool.Pool
}

func NewSignalRepo(pool *pgxpool.Pool) *SignalRepo {
	return &SignalRepo{pool: pool}
}

// UpsertSignalSnapshot inserts or updates a commodity signal snapshot.
func (r *SignalRepo) UpsertSignalSnapshot(ctx context.Context, s *domain.CommoditySignalSnapshot) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO commodity_signal_snapshots (
			commodity_id, as_of_date, net_managed_money, net_mm_pct_oi,
			position_zscore_26w, position_zscore_52w, position_percentile_52w,
			weekly_change_net_mm,
			crowding_score, squeeze_risk_score, reversal_risk_score, trend_support_score,
			model_version
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		ON CONFLICT (commodity_id, as_of_date, model_version)
		DO UPDATE SET
			net_managed_money = EXCLUDED.net_managed_money,
			net_mm_pct_oi = EXCLUDED.net_mm_pct_oi,
			position_zscore_26w = EXCLUDED.position_zscore_26w,
			position_zscore_52w = EXCLUDED.position_zscore_52w,
			position_percentile_52w = EXCLUDED.position_percentile_52w,
			weekly_change_net_mm = EXCLUDED.weekly_change_net_mm,
			crowding_score = EXCLUDED.crowding_score,
			squeeze_risk_score = EXCLUDED.squeeze_risk_score,
			reversal_risk_score = EXCLUDED.reversal_risk_score,
			trend_support_score = EXCLUDED.trend_support_score,
			created_at = now()
	`, s.CommodityID, s.AsOfDate, s.NetManagedMoney, s.NetMMPctOI,
		s.PositionZScore26W, s.PositionZScore52W, s.PositionPercentile,
		s.WeeklyChangeNetMM,
		s.CrowdingScore, s.SqueezeRiskScore, s.ReversalRiskScore, s.TrendSupportScore,
		s.ModelVersion)
	if err != nil {
		return fmt.Errorf("upserting signal snapshot: %w", err)
	}
	return nil
}

// GetLatestSignal returns the most recent signal snapshot for a commodity.
func (r *SignalRepo) GetLatestSignal(ctx context.Context, commodityID int64, modelVersion string) (*domain.CommoditySignalSnapshot, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, commodity_id, as_of_date, net_managed_money, net_mm_pct_oi,
			   position_zscore_26w, position_zscore_52w, position_percentile_52w,
			   weekly_change_net_mm,
			   crowding_score, squeeze_risk_score, reversal_risk_score, trend_support_score,
			   model_version, created_at
		FROM commodity_signal_snapshots
		WHERE commodity_id = $1 AND model_version = $2
		ORDER BY as_of_date DESC
		LIMIT 1
	`, commodityID, modelVersion)

	return scanSignal(row)
}

// ListSignals returns signal snapshots for a commodity within a date range, ordered descending.
func (r *SignalRepo) ListSignals(ctx context.Context, commodityID int64, from, to time.Time, modelVersion string) ([]domain.CommoditySignalSnapshot, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, commodity_id, as_of_date, net_managed_money, net_mm_pct_oi,
			   position_zscore_26w, position_zscore_52w, position_percentile_52w,
			   weekly_change_net_mm,
			   crowding_score, squeeze_risk_score, reversal_risk_score, trend_support_score,
			   model_version, created_at
		FROM commodity_signal_snapshots
		WHERE commodity_id = $1 AND as_of_date >= $2 AND as_of_date <= $3 AND model_version = $4
		ORDER BY as_of_date DESC
	`, commodityID, from, to, modelVersion)
	if err != nil {
		return nil, fmt.Errorf("querying signals: %w", err)
	}
	defer rows.Close()

	var signals []domain.CommoditySignalSnapshot
	for rows.Next() {
		var s domain.CommoditySignalSnapshot
		if err := rows.Scan(&s.ID, &s.CommodityID, &s.AsOfDate, &s.NetManagedMoney, &s.NetMMPctOI,
			&s.PositionZScore26W, &s.PositionZScore52W, &s.PositionPercentile,
			&s.WeeklyChangeNetMM,
			&s.CrowdingScore, &s.SqueezeRiskScore, &s.ReversalRiskScore, &s.TrendSupportScore,
			&s.ModelVersion, &s.CreatedAt); err != nil {
			return nil, fmt.Errorf("scanning signal: %w", err)
		}
		signals = append(signals, s)
	}
	return signals, rows.Err()
}

// ListCOTPositions returns clean COT positions for a commodity, ordered ascending by date.
func (r *SignalRepo) ListCOTPositions(ctx context.Context, commodityID int64, from, to time.Time) ([]domain.COTPositionClean, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, commodity_id, report_date, open_interest,
			   producer_merchant_long, producer_merchant_short,
			   swap_dealer_long, swap_dealer_short,
			   managed_money_long, managed_money_short,
			   other_reportable_long, other_reportable_short,
			   nonreportable_long, nonreportable_short
		FROM cot_positions_clean
		WHERE commodity_id = $1 AND report_date >= $2 AND report_date <= $3
		ORDER BY report_date ASC
	`, commodityID, from, to)
	if err != nil {
		return nil, fmt.Errorf("querying COT positions: %w", err)
	}
	defer rows.Close()

	var positions []domain.COTPositionClean
	for rows.Next() {
		var p domain.COTPositionClean
		if err := rows.Scan(&p.ID, &p.CommodityID, &p.ReportDate, &p.OpenInterest,
			&p.ProducerMerchantLong, &p.ProducerMerchantShort,
			&p.SwapDealerLong, &p.SwapDealerShort,
			&p.ManagedMoneyLong, &p.ManagedMoneyShort,
			&p.OtherReportableLong, &p.OtherReportableShort,
			&p.NonreportableLong, &p.NonreportableShort); err != nil {
			return nil, fmt.Errorf("scanning position: %w", err)
		}
		positions = append(positions, p)
	}
	return positions, rows.Err()
}

func scanSignal(row pgx.Row) (*domain.CommoditySignalSnapshot, error) {
	var s domain.CommoditySignalSnapshot
	err := row.Scan(&s.ID, &s.CommodityID, &s.AsOfDate, &s.NetManagedMoney, &s.NetMMPctOI,
		&s.PositionZScore26W, &s.PositionZScore52W, &s.PositionPercentile,
		&s.WeeklyChangeNetMM,
		&s.CrowdingScore, &s.SqueezeRiskScore, &s.ReversalRiskScore, &s.TrendSupportScore,
		&s.ModelVersion, &s.CreatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("scanning signal: %w", err)
	}
	return &s, nil
}
