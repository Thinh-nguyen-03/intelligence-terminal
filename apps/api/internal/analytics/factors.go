package analytics

import (
	"math"
	"sort"
	"time"

	"github.com/0510t/intelligence-terminal/apps/api/internal/domain"
)

// SeriesMap maps series slug to its ordered observations (ascending by date).
type SeriesMap map[string][]domain.MacroObservationClean

// ComputeInflationScore computes the inflation factor from CPI and Core CPI series.
//
// Method: 3-month annualized CPI momentum vs 12-month rate, confirmed by Core CPI direction.
// Output: [-1.0, +1.0] where positive = rising inflation.
func ComputeInflationScore(series SeriesMap) float64 {
	cpi := series["cpi"]
	coreCPI := series["core-cpi"]

	if len(cpi) < 13 {
		return 0
	}

	// CPI 3-month annualized rate
	mom3 := momentum(cpi, 3)
	annualized3 := mom3 * 4 // annualize monthly rate

	// CPI 12-month rate
	mom12 := momentum(cpi, 12)

	// Acceleration: how much the 3m rate exceeds the 12m rate
	acceleration := annualized3 - mom12

	// Core CPI confirmation
	coreConfirmation := 0.0
	if len(coreCPI) >= 4 {
		coreMom3 := momentum(coreCPI, 3)
		if coreMom3 > 0 {
			coreConfirmation = 0.15
		} else if coreMom3 < 0 {
			coreConfirmation = -0.15
		}
	}

	// Combine: weight acceleration heavily, add core confirmation
	raw := acceleration*8.0 + coreConfirmation
	return clampScore(raw)
}

// ComputeGrowthScore computes the growth factor from industrial production, retail sales, and payrolls.
//
// Method: average of 3-month momentum z-scores across growth indicators.
// Output: [-1.0, +1.0] where positive = economic expansion.
func ComputeGrowthScore(series SeriesMap) float64 {
	indpro := series["industrial-prod"]
	retail := series["retail-sales"]
	payrolls := series["nonfarm-payrolls"]

	scores := make([]float64, 0, 3)

	if len(indpro) >= 13 {
		mom := momentum(indpro, 3)
		zscore := zscoreFromMomentums(indpro, 3, 12)
		scores = append(scores, normalizeGrowth(mom, zscore))
	}
	if len(retail) >= 13 {
		mom := momentum(retail, 3)
		zscore := zscoreFromMomentums(retail, 3, 12)
		scores = append(scores, normalizeGrowth(mom, zscore))
	}
	if len(payrolls) >= 13 {
		mom := momentum(payrolls, 3)
		zscore := zscoreFromMomentums(payrolls, 3, 12)
		scores = append(scores, normalizeGrowth(mom, zscore))
	}

	if len(scores) == 0 {
		return 0
	}

	avg := 0.0
	for _, s := range scores {
		avg += s
	}
	return clampScore(avg / float64(len(scores)))
}

// ComputeLaborScore computes the labor market factor from unemployment and payrolls.
//
// Method: unemployment trend (inverted — falling unemployment = positive) + payroll trend.
// Output: [-1.0, +1.0] where positive = tight labor market.
func ComputeLaborScore(series SeriesMap) float64 {
	unrate := series["unemployment"]
	payrolls := series["nonfarm-payrolls"]

	scores := make([]float64, 0, 2)

	if len(unrate) >= 7 {
		// Inverted: falling unemployment = positive signal
		mom3 := momentum(unrate, 3)
		mom6 := momentum(unrate, 6)
		// Negative momentum in unemployment = good, so negate
		raw := -(mom3*0.6 + mom6*0.4) * 10.0
		scores = append(scores, clampScore(raw))
	}

	if len(payrolls) >= 7 {
		mom3 := momentum(payrolls, 3)
		zscore := 0.0
		if len(payrolls) >= 13 {
			zscore = zscoreFromMomentums(payrolls, 3, 12)
		}
		scores = append(scores, normalizeGrowth(mom3, zscore))
	}

	if len(scores) == 0 {
		return 0
	}

	avg := 0.0
	for _, s := range scores {
		avg += s
	}
	return clampScore(avg / float64(len(scores)))
}

// ComputeStressScore computes the financial stress factor from yield curve and dollar index.
//
// Method: yield curve inversion (10Y-2Y, inverted = high stress) + dollar strength as tightening proxy.
// Output: [-1.0, +1.0] where positive = elevated stress.
func ComputeStressScore(series SeriesMap) float64 {
	yieldCurve := series["yield-curve-10y2y"]
	dollar := series["dollar-index"]

	scores := make([]float64, 0, 2)

	if len(yieldCurve) >= 1 {
		// Use the most recent value directly
		// Normal spread ~1.0-2.0, inverted = negative values
		latest := lastValue(yieldCurve)
		// Invert: negative spread = high stress (positive score)
		// Normal: 1.5% spread → low stress, -0.5% spread → high stress
		raw := -latest * 0.5
		scores = append(scores, clampScore(raw))
	}

	if len(dollar) >= 60 {
		// Dollar strength: compute 3-month momentum z-score against 12-month window
		mom := momentumDaily(dollar, 63) // ~3 months of trading days
		zscore := zscoreFromMomDaily(dollar, 63, 252) // ~12 months
		// Strong dollar = tightening conditions = stress
		raw := (mom*5.0 + zscore*0.3) * 0.5
		scores = append(scores, clampScore(raw))
	} else if len(dollar) >= 5 {
		// Shorter window fallback
		mom := momentumDaily(dollar, len(dollar)-1)
		raw := mom * 5.0
		scores = append(scores, clampScore(raw))
	}

	if len(scores) == 0 {
		return 0
	}

	avg := 0.0
	for _, s := range scores {
		avg += s
	}
	return clampScore(avg / float64(len(scores)))
}

// --- helpers ---

// momentum computes the percentage change over `periods` (for monthly data).
// Returns the fractional change (e.g., 0.02 = 2%).
func momentum(obs []domain.MacroObservationClean, periods int) float64 {
	if len(obs) <= periods {
		return 0
	}
	current := obs[len(obs)-1].Value
	prior := obs[len(obs)-1-periods].Value
	if prior == 0 {
		return 0
	}
	return (current - prior) / math.Abs(prior)
}

// momentumDaily computes percentage change over `days` for daily series.
func momentumDaily(obs []domain.MacroObservationClean, days int) float64 {
	if len(obs) <= days {
		return 0
	}
	current := obs[len(obs)-1].Value
	prior := obs[len(obs)-1-days].Value
	if prior == 0 {
		return 0
	}
	return (current - prior) / math.Abs(prior)
}

// zscoreFromMomentums computes rolling momentum values over the lookback window,
// then returns the z-score of the most recent momentum vs the historical distribution.
func zscoreFromMomentums(obs []domain.MacroObservationClean, momPeriod, lookbackMonths int) float64 {
	if len(obs) < momPeriod+lookbackMonths {
		return 0
	}

	// Compute rolling momentums for the lookback window
	moms := make([]float64, 0, lookbackMonths)
	for i := len(obs) - lookbackMonths; i < len(obs); i++ {
		if i-momPeriod < 0 {
			continue
		}
		cur := obs[i].Value
		prev := obs[i-momPeriod].Value
		if prev != 0 {
			moms = append(moms, (cur-prev)/math.Abs(prev))
		}
	}

	if len(moms) < 3 {
		return 0
	}

	latest := moms[len(moms)-1]
	mean, std := meanStd(moms)
	if std < 1e-10 {
		return 0
	}
	return (latest - mean) / std
}

// zscoreFromMomDaily is like zscoreFromMomentums but for daily series.
func zscoreFromMomDaily(obs []domain.MacroObservationClean, momDays, lookbackDays int) float64 {
	if len(obs) < momDays+lookbackDays {
		return 0
	}

	moms := make([]float64, 0, lookbackDays)
	for i := len(obs) - lookbackDays; i < len(obs); i++ {
		if i-momDays < 0 {
			continue
		}
		cur := obs[i].Value
		prev := obs[i-momDays].Value
		if prev != 0 {
			moms = append(moms, (cur-prev)/math.Abs(prev))
		}
	}

	if len(moms) < 3 {
		return 0
	}

	latest := moms[len(moms)-1]
	mean, std := meanStd(moms)
	if std < 1e-10 {
		return 0
	}
	return (latest - mean) / std
}

func lastValue(obs []domain.MacroObservationClean) float64 {
	if len(obs) == 0 {
		return 0
	}
	return obs[len(obs)-1].Value
}

func normalizeGrowth(momentum, zscore float64) float64 {
	// Blend recent momentum and z-score for a balanced signal
	raw := momentum*5.0 + zscore*0.3
	return clampScore(raw)
}

func clampScore(v float64) float64 {
	if v > 1.0 {
		return 1.0
	}
	if v < -1.0 {
		return -1.0
	}
	return v
}

func meanStd(vals []float64) (float64, float64) {
	n := float64(len(vals))
	if n == 0 {
		return 0, 0
	}
	sum := 0.0
	for _, v := range vals {
		sum += v
	}
	mean := sum / n

	sumSq := 0.0
	for _, v := range vals {
		d := v - mean
		sumSq += d * d
	}
	std := math.Sqrt(sumSq / n)
	return mean, std
}

// SortObservations sorts observations by date ascending (in place).
func SortObservations(obs []domain.MacroObservationClean) {
	sort.Slice(obs, func(i, j int) bool {
		return obs[i].ObservationDate.Before(obs[j].ObservationDate)
	})
}

// FilterBefore returns observations on or before the given date.
func FilterBefore(obs []domain.MacroObservationClean, asOf time.Time) []domain.MacroObservationClean {
	var result []domain.MacroObservationClean
	for _, o := range obs {
		if !o.ObservationDate.After(asOf) {
			result = append(result, o)
		}
	}
	return result
}
