package jobs

import (
	"context"
	"crypto/sha256"
	"fmt"
	"log/slog"
	"time"

	"github.com/0510t/intelligence-terminal/apps/api/internal/domain"
	"github.com/0510t/intelligence-terminal/apps/api/internal/storage"
)

// CFTCIngestJob fetches and stores CFTC Disaggregated COT report data.
type CFTCIngestJob struct {
	client        *CFTCClient
	cotRepo       *storage.COTRepo
	sourceRunRepo *storage.SourceRunRepo
}

func NewCFTCIngestJob(client *CFTCClient, cotRepo *storage.COTRepo, sourceRunRepo *storage.SourceRunRepo) *CFTCIngestJob {
	return &CFTCIngestJob{
		client:        client,
		cotRepo:       cotRepo,
		sourceRunRepo: sourceRunRepo,
	}
}

// Run executes the CFTC ingestion for all active commodities.
// It downloads year-by-year CSVs going back lookbackYears from the current year.
func (j *CFTCIngestJob) Run(ctx context.Context, lookbackYears int) error {
	runID, err := j.sourceRunRepo.Start(ctx, domain.SourceCFTC, "ingest-cot")
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

	slog.Info("CFTC ingestion complete", "total_records", totalRecords)
	return nil
}

func (j *CFTCIngestJob) run(ctx context.Context, lookbackYears int) (int, error) {
	// Load active commodities and build code -> ID lookup
	commodities, err := j.cotRepo.ListActiveCommodities(ctx)
	if err != nil {
		return 0, fmt.Errorf("listing active commodities: %w", err)
	}
	if len(commodities) == 0 {
		slog.Warn("no active commodities found")
		return 0, nil
	}

	codeToID := make(map[string]int64)
	filterCodes := make(map[string]bool)
	for _, c := range commodities {
		codeToID[c.CFTCCommodityCode] = c.ID
		filterCodes[c.CFTCCommodityCode] = true
	}

	totalRecords := 0
	currentYear := time.Now().UTC().Year()
	startYear := currentYear - lookbackYears

	// Process historical years, then current year
	for year := startYear; year <= currentYear; year++ {
		if ctx.Err() != nil {
			return totalRecords, ctx.Err()
		}

		count, err := j.processYear(ctx, year, currentYear, filterCodes, codeToID)
		if err != nil {
			slog.Error("failed to process CFTC year", "year", year, "error", err)
			continue // partial failure tolerated
		}
		totalRecords += count
		slog.Info("processed CFTC year", "year", year, "records", count)
	}

	return totalRecords, nil
}

func (j *CFTCIngestJob) processYear(ctx context.Context, year, currentYear int, filterCodes map[string]bool, codeToID map[string]int64) (int, error) {
	var rawCSV string
	var err error

	if year == currentYear {
		rawCSV, err = j.client.FetchCurrentYear(ctx)
	} else {
		rawCSV, err = j.client.FetchHistoricalYear(ctx, year)
	}
	if err != nil {
		return 0, fmt.Errorf("fetching year %d: %w", year, err)
	}

	// Compute SHA-256 checksum for dedup
	checksum := fmt.Sprintf("%x", sha256.Sum256([]byte(rawCSV)))

	// Parse CSV, filtering to our commodity codes
	rows, err := ParseDisaggregatedReport(rawCSV, filterCodes)
	if err != nil {
		return 0, fmt.Errorf("parsing year %d: %w", year, err)
	}

	if len(rows) == 0 {
		slog.Info("no matching rows for year", "year", year)
		return 0, nil
	}

	// Find the max report date in the file for the raw record
	var maxDate time.Time
	for _, r := range rows {
		if r.ReportDate.After(maxDate) {
			maxDate = r.ReportDate
		}
	}

	// Store raw report (idempotent: skips if checksum or date+type already exists)
	rawReport := &domain.COTReportRaw{
		ReportDate:   maxDate,
		ReportType:   "disaggregated",
		FileChecksum: checksum,
		PayloadText:  &rawCSV,
	}
	inserted, err := j.cotRepo.InsertRawReport(ctx, rawReport)
	if err != nil {
		return 0, fmt.Errorf("inserting raw report for year %d: %w", year, err)
	}
	if !inserted {
		slog.Info("skipping already-processed file", "year", year, "checksum", checksum[:12])
		return 0, nil
	}

	// Upsert clean positions
	count := 0
	for _, row := range rows {
		commodityID, ok := codeToID[row.CFTCContractMarketCode]
		if !ok {
			continue
		}

		pos := &domain.COTPositionClean{
			CommodityID:           commodityID,
			ReportDate:            row.ReportDate,
			OpenInterest:          row.OpenInterest,
			ProducerMerchantLong:  row.ProdMercLong,
			ProducerMerchantShort: row.ProdMercShort,
			SwapDealerLong:        row.SwapDealerLong,
			SwapDealerShort:       row.SwapDealerShort,
			ManagedMoneyLong:      row.ManagedMoneyLong,
			ManagedMoneyShort:     row.ManagedMoneyShort,
			OtherReportableLong:   row.OtherReportableLong,
			OtherReportableShort:  row.OtherReportableShort,
			NonreportableLong:     row.NonreportableLong,
			NonreportableShort:    row.NonreportableShort,
		}

		if err := j.cotRepo.UpsertCleanPosition(ctx, pos); err != nil {
			slog.Warn("failed to upsert position", "commodity_id", commodityID, "date", row.ReportDate, "error", err)
			continue
		}
		count++
	}

	slog.Info("raw payload size", "year", year, "bytes", len(rawCSV))
	return count, nil
}
