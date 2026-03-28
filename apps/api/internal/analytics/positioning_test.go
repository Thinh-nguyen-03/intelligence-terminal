package analytics

import (
	"math"
	"testing"
	"time"

	"github.com/0510t/intelligence-terminal/apps/api/internal/domain"
	"github.com/0510t/intelligence-terminal/apps/api/internal/storage"
)

func posParams() *storage.ModelParams {
	return &storage.ModelParams{
		CrowdedLongZScore:      1.50,
		CrowdedShortZScore:     -1.50,
		CrowdedLongPercentile:  90,
		CrowdedShortPercentile: 10,
	}
}

// makeCOTPositions creates a slice of weekly COT positions with given net managed money values.
func makeCOTPositions(netMMs []int64, oi int64) []domain.COTPositionClean {
	base := time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC) // a Tuesday
	positions := make([]domain.COTPositionClean, len(netMMs))
	for i, net := range netMMs {
		long := int64(0)
		short := int64(0)
		if net >= 0 {
			long = net
		} else {
			short = -net
		}
		positions[i] = domain.COTPositionClean{
			CommodityID:      1,
			ReportDate:       base.AddDate(0, 0, i*7),
			OpenInterest:     oi,
			ManagedMoneyLong: long,
			ManagedMoneyShort: short,
		}
	}
	return positions
}

func TestComputePositionSignal_BasicMetrics(t *testing.T) {
	positions := makeCOTPositions([]int64{5000, 6000, 7000}, 100000)

	sig := ComputePositionSignal(positions, posParams())
	if sig == nil {
		t.Fatal("expected non-nil signal")
	}

	if sig.NetManagedMoney != 7000 {
		t.Errorf("expected net MM 7000, got %d", sig.NetManagedMoney)
	}
	if math.Abs(sig.NetMMPctOI-7.0) > 0.01 {
		t.Errorf("expected pct OI ~7.0, got %.4f", sig.NetMMPctOI)
	}
}

func TestComputePositionSignal_WeeklyChange(t *testing.T) {
	positions := makeCOTPositions([]int64{5000, 8000}, 100000)

	sig := ComputePositionSignal(positions, posParams())
	if sig == nil {
		t.Fatal("expected non-nil signal")
	}

	if sig.WeeklyChangeNetMM == nil {
		t.Fatal("expected weekly change to be set")
	}
	if *sig.WeeklyChangeNetMM != 3000 {
		t.Errorf("expected weekly change 3000, got %d", *sig.WeeklyChangeNetMM)
	}
}

func TestComputePositionSignal_ZScoresRequire26Weeks(t *testing.T) {
	// Only 10 weeks of data — z-scores should be nil
	positions := makeCOTPositions(make([]int64, 10), 100000)
	sig := ComputePositionSignal(positions, posParams())

	if sig.ZScore26W != nil {
		t.Error("expected 26w z-score to be nil with < 26 weeks of data")
	}
	if sig.ZScore52W != nil {
		t.Error("expected 52w z-score to be nil with < 52 weeks of data")
	}
}

func TestComputePositionSignal_ZScoresWith52Weeks(t *testing.T) {
	// 52 weeks of steadily increasing net MM
	netMMs := make([]int64, 52)
	for i := range netMMs {
		netMMs[i] = int64(1000 + i*100)
	}
	positions := makeCOTPositions(netMMs, 200000)

	sig := ComputePositionSignal(positions, posParams())
	if sig.ZScore26W == nil {
		t.Error("expected 26w z-score to be set")
	}
	if sig.ZScore52W == nil {
		t.Error("expected 52w z-score to be set")
	}
	// Latest value is the highest, so z-score should be positive
	if sig.ZScore52W != nil && *sig.ZScore52W <= 0 {
		t.Errorf("expected positive 52w z-score for increasing series, got %.3f", *sig.ZScore52W)
	}
	if sig.Percentile52W == nil {
		t.Error("expected percentile to be set")
	}
	// Latest is highest, so percentile should be near 100
	if sig.Percentile52W != nil && *sig.Percentile52W < 90 {
		t.Errorf("expected high percentile, got %.1f", *sig.Percentile52W)
	}
}

func TestComputePositionSignal_CrowdingHigh(t *testing.T) {
	// Create data where latest is extreme (high z-score)
	netMMs := make([]int64, 52)
	for i := 0; i < 50; i++ {
		netMMs[i] = 5000
	}
	netMMs[50] = 15000
	netMMs[51] = 20000 // extreme spike
	positions := makeCOTPositions(netMMs, 200000)

	sig := ComputePositionSignal(positions, posParams())
	if sig.CrowdingScore < 0.5 {
		t.Errorf("expected high crowding score for extreme position, got %.2f", sig.CrowdingScore)
	}
}

func TestComputePositionSignal_SqueezeRisk(t *testing.T) {
	// Extreme short position with short covering (positive weekly change)
	netMMs := make([]int64, 52)
	for i := 0; i < 50; i++ {
		netMMs[i] = -5000
	}
	netMMs[50] = -20000 // extreme short
	netMMs[51] = -18000 // short covering (less negative = positive change)
	positions := makeCOTPositions(netMMs, 200000)

	sig := ComputePositionSignal(positions, posParams())
	if sig.SqueezeRiskScore < 0.3 {
		t.Errorf("expected elevated squeeze risk, got %.2f", sig.SqueezeRiskScore)
	}
}

func TestComputePositionSignal_NoSqueezeWhenLong(t *testing.T) {
	// Net long position — no squeeze risk
	netMMs := make([]int64, 52)
	for i := range netMMs {
		netMMs[i] = 10000
	}
	positions := makeCOTPositions(netMMs, 200000)

	sig := ComputePositionSignal(positions, posParams())
	if sig.SqueezeRiskScore != 0 {
		t.Errorf("expected 0 squeeze risk for long position, got %.2f", sig.SqueezeRiskScore)
	}
}

func TestComputePositionSignal_EmptyPositions(t *testing.T) {
	sig := ComputePositionSignal(nil, posParams())
	if sig != nil {
		t.Error("expected nil signal for empty positions")
	}
}

func TestZscore(t *testing.T) {
	// Values: 1,2,3,...,10. Latest=10. Mean=5.5, std≈2.87. Z=(10-5.5)/2.87≈1.57
	vals := []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	z := zscore(vals, 10)
	if z < 1.0 || z > 2.0 {
		t.Errorf("expected z-score around 1.57, got %.3f", z)
	}
}

func TestPercentile(t *testing.T) {
	// Values 1-10, latest=10. 9 out of 10 values are below → 90th percentile
	vals := []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	p := percentile(vals, 10)
	if p != 90.0 {
		t.Errorf("expected percentile 90.0, got %.1f", p)
	}
}

func TestPercentile_Lowest(t *testing.T) {
	vals := []float64{10, 9, 8, 7, 6, 5, 4, 3, 2, 1}
	p := percentile(vals, 10)
	if p != 0.0 {
		t.Errorf("expected percentile 0.0 for lowest value, got %.1f", p)
	}
}
