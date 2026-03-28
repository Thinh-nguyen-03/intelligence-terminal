package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/0510t/intelligence-terminal/apps/api/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

// COTRepo handles persistence for COT reports and positions.
type COTRepo struct {
	pool *pgxpool.Pool
}

func NewCOTRepo(pool *pgxpool.Pool) *COTRepo {
	return &COTRepo{pool: pool}
}

// ListActiveCommodities returns all active commodities.
func (r *COTRepo) ListActiveCommodities(ctx context.Context) ([]domain.Commodity, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, slug, name, cftc_commodity_code, group_name, active, created_at
		FROM commodities
		WHERE active = true
		ORDER BY id
	`)
	if err != nil {
		return nil, fmt.Errorf("querying active commodities: %w", err)
	}
	defer rows.Close()

	var commodities []domain.Commodity
	for rows.Next() {
		var c domain.Commodity
		if err := rows.Scan(&c.ID, &c.Slug, &c.Name, &c.CFTCCommodityCode, &c.GroupName, &c.Active, &c.CreatedAt); err != nil {
			return nil, fmt.Errorf("scanning commodity: %w", err)
		}
		commodities = append(commodities, c)
	}
	return commodities, rows.Err()
}

// InsertRawReport stores a raw COT report file. Skips silently if the checksum
// or (report_date, report_type) already exists. Returns true if a new row was inserted.
func (r *COTRepo) InsertRawReport(ctx context.Context, report *domain.COTReportRaw) (bool, error) {
	tag, err := r.pool.Exec(ctx, `
		INSERT INTO cot_reports_raw (report_date, report_type, file_checksum, payload_text)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT DO NOTHING
	`, report.ReportDate, report.ReportType, report.FileChecksum, report.PayloadText)
	if err != nil {
		return false, fmt.Errorf("inserting raw report: %w", err)
	}
	return tag.RowsAffected() > 0, nil
}

// UpsertCleanPosition inserts or updates a clean COT position row.
func (r *COTRepo) UpsertCleanPosition(ctx context.Context, pos *domain.COTPositionClean) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO cot_positions_clean (
			commodity_id, report_date, open_interest,
			producer_merchant_long, producer_merchant_short,
			swap_dealer_long, swap_dealer_short,
			managed_money_long, managed_money_short,
			other_reportable_long, other_reportable_short,
			nonreportable_long, nonreportable_short
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		ON CONFLICT (commodity_id, report_date)
		DO UPDATE SET
			open_interest = EXCLUDED.open_interest,
			producer_merchant_long = EXCLUDED.producer_merchant_long,
			producer_merchant_short = EXCLUDED.producer_merchant_short,
			swap_dealer_long = EXCLUDED.swap_dealer_long,
			swap_dealer_short = EXCLUDED.swap_dealer_short,
			managed_money_long = EXCLUDED.managed_money_long,
			managed_money_short = EXCLUDED.managed_money_short,
			other_reportable_long = EXCLUDED.other_reportable_long,
			other_reportable_short = EXCLUDED.other_reportable_short,
			nonreportable_long = EXCLUDED.nonreportable_long,
			nonreportable_short = EXCLUDED.nonreportable_short
	`, pos.CommodityID, pos.ReportDate, pos.OpenInterest,
		pos.ProducerMerchantLong, pos.ProducerMerchantShort,
		pos.SwapDealerLong, pos.SwapDealerShort,
		pos.ManagedMoneyLong, pos.ManagedMoneyShort,
		pos.OtherReportableLong, pos.OtherReportableShort,
		pos.NonreportableLong, pos.NonreportableShort)
	if err != nil {
		return fmt.Errorf("upserting clean position: %w", err)
	}
	return nil
}

// GetLatestReportDate returns the most recent report date across all commodities, or nil if empty.
func (r *COTRepo) GetLatestReportDate(ctx context.Context) (*time.Time, error) {
	var d time.Time
	err := r.pool.QueryRow(ctx, `SELECT MAX(report_date) FROM cot_positions_clean`).Scan(&d)
	if err != nil {
		return nil, nil
	}
	if d.IsZero() {
		return nil, nil
	}
	return &d, nil
}
