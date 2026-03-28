package analytics

import (
	"math"
	"testing"
	"time"

	"github.com/0510t/intelligence-terminal/apps/api/internal/domain"
)

// makeObs creates a slice of monthly observations with the given values,
// starting from a base date and incrementing by 1 month each.
func makeObs(values []float64) []domain.MacroObservationClean {
	base := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	obs := make([]domain.MacroObservationClean, len(values))
	for i, v := range values {
		obs[i] = domain.MacroObservationClean{
			ObservationDate: base.AddDate(0, i, 0),
			Value:           v,
		}
	}
	return obs
}

// makeDailyObs creates a slice of daily observations.
func makeDailyObs(values []float64) []domain.MacroObservationClean {
	base := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	obs := make([]domain.MacroObservationClean, len(values))
	for i, v := range values {
		obs[i] = domain.MacroObservationClean{
			ObservationDate: base.AddDate(0, 0, i),
			Value:           v,
		}
	}
	return obs
}

func TestComputeInflationScore_Rising(t *testing.T) {
	// CPI rising steadily: each month 0.5% higher than previous
	cpi := make([]float64, 24)
	cpi[0] = 300.0
	for i := 1; i < 24; i++ {
		cpi[i] = cpi[i-1] * 1.005
	}

	series := SeriesMap{
		"cpi":      makeObs(cpi),
		"core-cpi": makeObs(cpi),
	}

	score := ComputeInflationScore(series)
	if score <= 0 {
		t.Errorf("expected positive inflation score for rising CPI, got %.3f", score)
	}
}

func TestComputeInflationScore_Falling(t *testing.T) {
	// CPI falling: each month 0.3% lower
	cpi := make([]float64, 24)
	cpi[0] = 300.0
	for i := 1; i < 24; i++ {
		cpi[i] = cpi[i-1] * 0.997
	}

	series := SeriesMap{
		"cpi":      makeObs(cpi),
		"core-cpi": makeObs(cpi),
	}

	score := ComputeInflationScore(series)
	if score >= 0 {
		t.Errorf("expected negative inflation score for falling CPI, got %.3f", score)
	}
}

func TestComputeInflationScore_InsufficientData(t *testing.T) {
	series := SeriesMap{
		"cpi":      makeObs([]float64{100, 101, 102}),
		"core-cpi": makeObs([]float64{100, 101, 102}),
	}

	score := ComputeInflationScore(series)
	if score != 0 {
		t.Errorf("expected 0 for insufficient data, got %.3f", score)
	}
}

func TestComputeGrowthScore_Expanding(t *testing.T) {
	// All growth indicators rising steadily
	vals := make([]float64, 24)
	vals[0] = 100.0
	for i := 1; i < 24; i++ {
		vals[i] = vals[i-1] * 1.003
	}

	series := SeriesMap{
		"industrial-prod":   makeObs(vals),
		"retail-sales":      makeObs(vals),
		"nonfarm-payrolls":  makeObs(vals),
	}

	score := ComputeGrowthScore(series)
	if score <= 0 {
		t.Errorf("expected positive growth score for expanding economy, got %.3f", score)
	}
}

func TestComputeGrowthScore_Contracting(t *testing.T) {
	vals := make([]float64, 24)
	vals[0] = 100.0
	for i := 1; i < 24; i++ {
		vals[i] = vals[i-1] * 0.995
	}

	series := SeriesMap{
		"industrial-prod":   makeObs(vals),
		"retail-sales":      makeObs(vals),
		"nonfarm-payrolls":  makeObs(vals),
	}

	score := ComputeGrowthScore(series)
	if score >= 0 {
		t.Errorf("expected negative growth score for contracting economy, got %.3f", score)
	}
}

func TestComputeGrowthScore_NoData(t *testing.T) {
	series := SeriesMap{}
	score := ComputeGrowthScore(series)
	if score != 0 {
		t.Errorf("expected 0 for no data, got %.3f", score)
	}
}

func TestComputeLaborScore_TightMarket(t *testing.T) {
	// Unemployment falling
	unrate := make([]float64, 24)
	unrate[0] = 5.0
	for i := 1; i < 24; i++ {
		unrate[i] = unrate[i-1] - 0.05
	}

	// Payrolls rising
	payrolls := make([]float64, 24)
	payrolls[0] = 150000
	for i := 1; i < 24; i++ {
		payrolls[i] = payrolls[i-1] * 1.002
	}

	series := SeriesMap{
		"unemployment":      makeObs(unrate),
		"nonfarm-payrolls":  makeObs(payrolls),
	}

	score := ComputeLaborScore(series)
	if score <= 0 {
		t.Errorf("expected positive labor score for tight market, got %.3f", score)
	}
}

func TestComputeStressScore_InvertedCurve(t *testing.T) {
	// Yield curve inverted (negative spread)
	yc := makeDailyObs([]float64{-0.5})

	series := SeriesMap{
		"yield-curve-10y2y": yc,
	}

	score := ComputeStressScore(series)
	if score <= 0 {
		t.Errorf("expected positive stress score for inverted yield curve, got %.3f", score)
	}
}

func TestComputeStressScore_NormalCurve(t *testing.T) {
	// Normal yield curve (positive spread)
	yc := makeDailyObs([]float64{1.5})

	series := SeriesMap{
		"yield-curve-10y2y": yc,
	}

	score := ComputeStressScore(series)
	if score >= 0 {
		t.Errorf("expected negative stress score for normal yield curve, got %.3f", score)
	}
}

func TestClampScore(t *testing.T) {
	tests := []struct {
		input, want float64
	}{
		{0.5, 0.5},
		{1.5, 1.0},
		{-1.5, -1.0},
		{0, 0},
		{1.0, 1.0},
		{-1.0, -1.0},
	}
	for _, tt := range tests {
		got := clampScore(tt.input)
		if got != tt.want {
			t.Errorf("clampScore(%.1f) = %.1f, want %.1f", tt.input, got, tt.want)
		}
	}
}

func TestMeanStd(t *testing.T) {
	vals := []float64{2, 4, 4, 4, 5, 5, 7, 9}
	mean, std := meanStd(vals)

	if math.Abs(mean-5.0) > 0.01 {
		t.Errorf("expected mean 5.0, got %.3f", mean)
	}
	if math.Abs(std-2.0) > 0.01 {
		t.Errorf("expected std 2.0, got %.3f", std)
	}
}

func TestMeanStd_Empty(t *testing.T) {
	mean, std := meanStd(nil)
	if mean != 0 || std != 0 {
		t.Errorf("expected (0,0) for empty, got (%.3f, %.3f)", mean, std)
	}
}

func TestSortObservations(t *testing.T) {
	obs := []domain.MacroObservationClean{
		{ObservationDate: time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC), Value: 3},
		{ObservationDate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), Value: 1},
		{ObservationDate: time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC), Value: 2},
	}
	SortObservations(obs)

	if obs[0].Value != 1 || obs[1].Value != 2 || obs[2].Value != 3 {
		t.Errorf("observations not sorted correctly: %v", obs)
	}
}

func TestFilterBefore(t *testing.T) {
	obs := makeObs([]float64{1, 2, 3, 4, 5})
	cutoff := time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC)
	filtered := FilterBefore(obs, cutoff)

	if len(filtered) != 3 {
		t.Errorf("expected 3 observations on or before cutoff, got %d", len(filtered))
	}
}
