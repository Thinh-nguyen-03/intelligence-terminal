package jobs

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	cftcCurrentYearURL   = "https://www.cftc.gov/dea/newcot/f_disagg.txt"
	cftcHistoricalURLFmt = "https://www.cftc.gov/files/dea/history/fut_disagg_txt_%d.zip"
)

// CFTCClient fetches Disaggregated COT report data from the CFTC website.
type CFTCClient struct {
	httpClient *http.Client
}

func NewCFTCClient() *CFTCClient {
	return &CFTCClient{
		httpClient: &http.Client{
			Timeout: 120 * time.Second,
		},
	}
}

// DisaggregatedRow is one parsed row from the CFTC Disaggregated Futures-Only report.
type DisaggregatedRow struct {
	MarketName             string
	CFTCContractMarketCode string
	ReportDate             time.Time
	OpenInterest           int64
	ProdMercLong           int64
	ProdMercShort          int64
	SwapDealerLong         int64
	SwapDealerShort        int64
	ManagedMoneyLong       int64
	ManagedMoneyShort      int64
	OtherReportableLong    int64
	OtherReportableShort   int64
	NonreportableLong      int64
	NonreportableShort     int64
}

// FetchCurrentYear downloads the current year's Disaggregated Futures-Only report.
func (c *CFTCClient) FetchCurrentYear(ctx context.Context) (string, error) {
	return c.fetchURL(ctx, cftcCurrentYearURL)
}

// FetchHistoricalYear downloads a historical year's report from a zip archive.
func (c *CFTCClient) FetchHistoricalYear(ctx context.Context, year int) (string, error) {
	url := fmt.Sprintf(cftcHistoricalURLFmt, year)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("creating request: %w", err)
	}

	slog.Info("fetching CFTC historical data", "year", year)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("CFTC returned status %d for year %d", resp.StatusCode, year)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("reading response: %w", err)
	}

	zipReader, err := zip.NewReader(bytes.NewReader(body), int64(len(body)))
	if err != nil {
		return "", fmt.Errorf("opening zip: %w", err)
	}

	for _, f := range zipReader.File {
		rc, err := f.Open()
		if err != nil {
			return "", fmt.Errorf("opening file in zip: %w", err)
		}
		content, err := io.ReadAll(rc)
		rc.Close()
		if err != nil {
			return "", fmt.Errorf("reading file in zip: %w", err)
		}
		return string(content), nil
	}

	return "", fmt.Errorf("zip archive is empty for year %d", year)
}

func (c *CFTCClient) fetchURL(ctx context.Context, url string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("creating request: %w", err)
	}

	slog.Info("fetching CFTC data", "url", url)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("CFTC returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("reading response: %w", err)
	}

	return string(body), nil
}

// ParseDisaggregatedReport parses the CSV and returns rows matching the given CFTC contract market codes.
func ParseDisaggregatedReport(rawCSV string, filterCodes map[string]bool) ([]DisaggregatedRow, error) {
	reader := csv.NewReader(strings.NewReader(rawCSV))
	reader.LazyQuotes = true

	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("parsing CSV: %w", err)
	}

	if len(records) < 2 {
		return nil, fmt.Errorf("CSV has no data rows")
	}

	// Build column index map from headers
	colMap := make(map[string]int)
	for i, h := range records[0] {
		colMap[strings.TrimSpace(h)] = i
	}

	// Resolve required columns (handles naming inconsistencies across years)
	colContractCode, err := findCol(colMap, "CFTC_Contract_Market_Code")
	if err != nil {
		return nil, err
	}
	colReportDate, err := findCol(colMap, "Report_Date_as_YYYY-MM-DD")
	if err != nil {
		return nil, err
	}
	colOI, err := findCol(colMap, "Open_Interest_All")
	if err != nil {
		return nil, err
	}
	colProdLong, err := findCol(colMap, "Prod_Merc_Positions_Long_All")
	if err != nil {
		return nil, err
	}
	colProdShort, err := findCol(colMap, "Prod_Merc_Positions_Short_All")
	if err != nil {
		return nil, err
	}
	colSwapLong, err := findCol(colMap, "Swap_Positions_Long_All", "Swap__Positions_Long_All")
	if err != nil {
		return nil, err
	}
	colSwapShort, err := findCol(colMap, "Swap__Positions_Short_All", "Swap_Positions_Short_All")
	if err != nil {
		return nil, err
	}
	colMMLong, err := findCol(colMap, "M_Money_Positions_Long_All")
	if err != nil {
		return nil, err
	}
	colMMShort, err := findCol(colMap, "M_Money_Positions_Short_All")
	if err != nil {
		return nil, err
	}
	colOtherLong, err := findCol(colMap, "Other_Rept_Positions_Long_All")
	if err != nil {
		return nil, err
	}
	colOtherShort, err := findCol(colMap, "Other_Rept_Positions_Short_All")
	if err != nil {
		return nil, err
	}
	colNonreptLong, err := findCol(colMap, "NonRept_Positions_Long_All")
	if err != nil {
		return nil, err
	}
	colNonreptShort, err := findCol(colMap, "NonRept_Positions_Short_All")
	if err != nil {
		return nil, err
	}
	colMarketName, _ := findCol(colMap, "Market_and_Exchange_Names")

	var rows []DisaggregatedRow
	for i := 1; i < len(records); i++ {
		rec := records[i]
		if len(rec) <= colContractCode {
			continue
		}

		code := strings.TrimSpace(rec[colContractCode])
		if !filterCodes[code] {
			continue
		}

		reportDate, err := time.Parse("2006-01-02", strings.TrimSpace(rec[colReportDate]))
		if err != nil {
			slog.Warn("skipping row with invalid date", "row", i, "date", rec[colReportDate])
			continue
		}

		row := DisaggregatedRow{
			CFTCContractMarketCode: code,
			ReportDate:             reportDate,
			OpenInterest:           parseCFTCInt(rec[colOI]),
			ProdMercLong:           parseCFTCInt(rec[colProdLong]),
			ProdMercShort:          parseCFTCInt(rec[colProdShort]),
			SwapDealerLong:         parseCFTCInt(rec[colSwapLong]),
			SwapDealerShort:        parseCFTCInt(rec[colSwapShort]),
			ManagedMoneyLong:       parseCFTCInt(rec[colMMLong]),
			ManagedMoneyShort:      parseCFTCInt(rec[colMMShort]),
			OtherReportableLong:    parseCFTCInt(rec[colOtherLong]),
			OtherReportableShort:   parseCFTCInt(rec[colOtherShort]),
			NonreportableLong:      parseCFTCInt(rec[colNonreptLong]),
			NonreportableShort:     parseCFTCInt(rec[colNonreptShort]),
		}

		if colMarketName >= 0 && colMarketName < len(rec) {
			row.MarketName = strings.TrimSpace(rec[colMarketName])
		}

		rows = append(rows, row)
	}

	return rows, nil
}

// findCol looks up a column index by trying multiple header names (cross-year compatibility).
func findCol(colMap map[string]int, names ...string) (int, error) {
	for _, name := range names {
		if idx, ok := colMap[name]; ok {
			return idx, nil
		}
	}
	return -1, fmt.Errorf("required column not found, tried: %v", names)
}

func parseCFTCInt(s string) int64 {
	v, _ := strconv.ParseInt(strings.TrimSpace(s), 10, 64)
	return v
}
