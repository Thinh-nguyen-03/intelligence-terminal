package jobs

import (
	"testing"
	"time"
)

const testCSV = `Market_and_Exchange_Names,As_of_Date_In_Form_YYMMDD,Report_Date_as_YYYY-MM-DD,CFTC_Contract_Market_Code,CFTC_Market_Code in Initials,CFTC_Region_Code,CFTC_Commodity_Code,Open_Interest_All,Prod_Merc_Positions_Long_All,Prod_Merc_Positions_Short_All,Swap_Positions_Long_All,Swap__Positions_Short_All,M_Money_Positions_Long_All,M_Money_Positions_Short_All,Other_Rept_Positions_Long_All,Other_Rept_Positions_Short_All,NonRept_Positions_Long_All,NonRept_Positions_Short_All
"GOLD - COMMODITY EXCHANGE INC.",260113,2026-01-13,088691,CEI,0,088,550000,80000,120000,90000,70000,200000,50000,30000,40000,150000,270000
"SILVER - COMMODITY EXCHANGE INC.",260113,2026-01-13,084691,CEI,0,084,180000,30000,50000,25000,20000,70000,15000,10000,12000,45000,83000
"WHEAT-SRW - CHICAGO BOARD OF TRADE",260113,2026-01-13,001602,CBT,0,001,400000,100000,150000,60000,40000,120000,80000,50000,60000,70000,70000
"GOLD - COMMODITY EXCHANGE INC.",260120,2026-01-20,088691,CEI,0,088,560000,82000,118000,92000,72000,205000,48000,32000,42000,149000,280000
`

func TestParseDisaggregatedReport(t *testing.T) {
	filterCodes := map[string]bool{
		"088691": true, // gold
		"084691": true, // silver
	}

	rows, err := ParseDisaggregatedReport(testCSV, filterCodes)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(rows) != 3 {
		t.Fatalf("expected 3 rows (2 gold + 1 silver), got %d", len(rows))
	}

	// First row: gold 2026-01-13
	gold1 := rows[0]
	if gold1.CFTCContractMarketCode != "088691" {
		t.Errorf("expected code 088691, got %s", gold1.CFTCContractMarketCode)
	}
	if gold1.ReportDate != time.Date(2026, 1, 13, 0, 0, 0, 0, time.UTC) {
		t.Errorf("expected 2026-01-13, got %s", gold1.ReportDate)
	}
	if gold1.OpenInterest != 550000 {
		t.Errorf("expected OI 550000, got %d", gold1.OpenInterest)
	}
	if gold1.ManagedMoneyLong != 200000 {
		t.Errorf("expected MM long 200000, got %d", gold1.ManagedMoneyLong)
	}
	if gold1.ManagedMoneyShort != 50000 {
		t.Errorf("expected MM short 50000, got %d", gold1.ManagedMoneyShort)
	}
	if gold1.ProdMercLong != 80000 {
		t.Errorf("expected prod long 80000, got %d", gold1.ProdMercLong)
	}
	if gold1.SwapDealerLong != 90000 {
		t.Errorf("expected swap long 90000, got %d", gold1.SwapDealerLong)
	}
	if gold1.SwapDealerShort != 70000 {
		t.Errorf("expected swap short 70000, got %d", gold1.SwapDealerShort)
	}
	if gold1.MarketName != "GOLD - COMMODITY EXCHANGE INC." {
		t.Errorf("expected market name 'GOLD - COMMODITY EXCHANGE INC.', got %q", gold1.MarketName)
	}

	// Second row: silver
	silver := rows[1]
	if silver.CFTCContractMarketCode != "084691" {
		t.Errorf("expected code 084691, got %s", silver.CFTCContractMarketCode)
	}
	if silver.OpenInterest != 180000 {
		t.Errorf("expected OI 180000, got %d", silver.OpenInterest)
	}

	// Third row: gold 2026-01-20
	gold2 := rows[2]
	if gold2.ReportDate != time.Date(2026, 1, 20, 0, 0, 0, 0, time.UTC) {
		t.Errorf("expected 2026-01-20, got %s", gold2.ReportDate)
	}
	if gold2.OpenInterest != 560000 {
		t.Errorf("expected OI 560000, got %d", gold2.OpenInterest)
	}
}

func TestParseDisaggregatedReport_FiltersCorrectly(t *testing.T) {
	filterCodes := map[string]bool{
		"084691": true, // silver only
	}

	rows, err := ParseDisaggregatedReport(testCSV, filterCodes)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(rows) != 1 {
		t.Fatalf("expected 1 row (silver only), got %d", len(rows))
	}
	if rows[0].CFTCContractMarketCode != "084691" {
		t.Errorf("expected silver code, got %s", rows[0].CFTCContractMarketCode)
	}
}

func TestParseDisaggregatedReport_NoMatchingCodes(t *testing.T) {
	filterCodes := map[string]bool{
		"999999": true,
	}

	rows, err := ParseDisaggregatedReport(testCSV, filterCodes)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rows) != 0 {
		t.Errorf("expected 0 rows, got %d", len(rows))
	}
}

func TestParseDisaggregatedReport_EmptyCSV(t *testing.T) {
	_, err := ParseDisaggregatedReport("", map[string]bool{"088691": true})
	if err == nil {
		t.Error("expected error for empty CSV")
	}
}

func TestParseDisaggregatedReport_HeadersOnly(t *testing.T) {
	csv := "Market_and_Exchange_Names,CFTC_Contract_Market_Code,Report_Date_as_YYYY-MM-DD\n"
	_, err := ParseDisaggregatedReport(csv, map[string]bool{"088691": true})
	if err == nil {
		t.Error("expected error for missing required columns")
	}
}

// Tests the alternate swap column naming that some historical years use.
func TestParseDisaggregatedReport_AlternateSwapColumns(t *testing.T) {
	// Uses Swap__Positions_Long_All (double underscore) and Swap_Positions_Short_All (single underscore)
	csv := `Market_and_Exchange_Names,Report_Date_as_YYYY-MM-DD,CFTC_Contract_Market_Code,Open_Interest_All,Prod_Merc_Positions_Long_All,Prod_Merc_Positions_Short_All,Swap__Positions_Long_All,Swap_Positions_Short_All,M_Money_Positions_Long_All,M_Money_Positions_Short_All,Other_Rept_Positions_Long_All,Other_Rept_Positions_Short_All,NonRept_Positions_Long_All,NonRept_Positions_Short_All
"GOLD - COMMODITY EXCHANGE INC.",2026-01-13,088691,550000,80000,120000,90000,70000,200000,50000,30000,40000,150000,270000
`
	rows, err := ParseDisaggregatedReport(csv, map[string]bool{"088691": true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}
	if rows[0].SwapDealerLong != 90000 {
		t.Errorf("expected swap long 90000, got %d", rows[0].SwapDealerLong)
	}
	if rows[0].SwapDealerShort != 70000 {
		t.Errorf("expected swap short 70000, got %d", rows[0].SwapDealerShort)
	}
}

func TestParseCFTCInt(t *testing.T) {
	tests := []struct {
		input string
		want  int64
	}{
		{"12345", 12345},
		{" 12345 ", 12345},
		{"0", 0},
		{"-5000", -5000},
		{"", 0},
		{"abc", 0},
	}
	for _, tt := range tests {
		got := parseCFTCInt(tt.input)
		if got != tt.want {
			t.Errorf("parseCFTCInt(%q) = %d, want %d", tt.input, got, tt.want)
		}
	}
}

func TestFindCol(t *testing.T) {
	colMap := map[string]int{
		"Alpha": 0,
		"Beta":  1,
		"Gamma": 2,
	}

	idx, err := findCol(colMap, "Beta")
	if err != nil || idx != 1 {
		t.Errorf("expected 1, got %d (err: %v)", idx, err)
	}

	// Fallback name
	idx, err = findCol(colMap, "Missing", "Gamma")
	if err != nil || idx != 2 {
		t.Errorf("expected 2 via fallback, got %d (err: %v)", idx, err)
	}

	// All missing
	_, err = findCol(colMap, "X", "Y", "Z")
	if err == nil {
		t.Error("expected error when all names missing")
	}
}
