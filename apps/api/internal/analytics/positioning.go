package analytics

import (
	"math"
	"sort"

	"github.com/0510t/intelligence-terminal/apps/api/internal/domain"
	"github.com/0510t/intelligence-terminal/apps/api/internal/storage"
)

// PositionSignal holds computed positioning metrics for a single commodity.
type PositionSignal struct {
	NetManagedMoney    int64
	NetMMPctOI         float64
	ZScore26W          *float64
	ZScore52W          *float64
	Percentile52W      *float64
	WeeklyChangeNetMM  *int64
	CrowdingScore      float64
	SqueezeRiskScore   float64
	ReversalRiskScore  float64
	TrendSupportScore  float64
}

// ComputePositionSignal computes positioning metrics from a history of clean COT positions.
// Positions must be sorted by report_date ascending. The last element is the current week.
func ComputePositionSignal(positions []domain.COTPositionClean, params *storage.ModelParams) *PositionSignal {
	if len(positions) == 0 {
		return nil
	}

	current := positions[len(positions)-1]
	netMM := current.ManagedMoneyLong - current.ManagedMoneyShort
	oi := current.OpenInterest

	pctOI := 0.0
	if oi > 0 {
		pctOI = float64(netMM) / float64(oi) * 100.0
	}

	sig := &PositionSignal{
		NetManagedMoney: netMM,
		NetMMPctOI:      pctOI,
	}

	// Build historical net managed money series
	netMMHistory := make([]float64, len(positions))
	for i, p := range positions {
		netMMHistory[i] = float64(p.ManagedMoneyLong - p.ManagedMoneyShort)
	}

	// 26-week z-score
	if len(netMMHistory) >= 26 {
		z := zscore(netMMHistory, 26)
		sig.ZScore26W = &z
	}

	// 52-week z-score
	if len(netMMHistory) >= 52 {
		z := zscore(netMMHistory, 52)
		sig.ZScore52W = &z
	}

	// 52-week percentile
	if len(netMMHistory) >= 52 {
		p := percentile(netMMHistory, 52)
		sig.Percentile52W = &p
	}

	// Weekly change
	if len(positions) >= 2 {
		prevNet := positions[len(positions)-2].ManagedMoneyLong - positions[len(positions)-2].ManagedMoneyShort
		change := netMM - prevNet
		sig.WeeklyChangeNetMM = &change
	}

	// Crowding score: based on z-score and percentile extremes
	sig.CrowdingScore = computeCrowding(sig, params)

	// Squeeze risk: extreme short + accelerating short covering
	sig.SqueezeRiskScore = computeSqueezeRisk(sig, params)

	// Reversal risk: extreme position + decelerating momentum
	sig.ReversalRiskScore = computeReversalRisk(sig, params)

	// Trend support: moderate positioning with consistent direction
	sig.TrendSupportScore = computeTrendSupport(sig, netMMHistory)

	return sig
}

// zscore computes the z-score of the latest value vs the last `window` values.
func zscore(vals []float64, window int) float64 {
	if len(vals) < window {
		return 0
	}
	windowVals := vals[len(vals)-window:]
	latest := vals[len(vals)-1]
	mean, std := meanStd(windowVals)
	if std < 1e-10 {
		return 0
	}
	return (latest - mean) / std
}

// percentile computes the percentile rank (0-100) of the latest value within the last `window` values.
func percentile(vals []float64, window int) float64 {
	if len(vals) < window {
		return 50
	}
	windowVals := make([]float64, window)
	copy(windowVals, vals[len(vals)-window:])
	latest := vals[len(vals)-1]

	sort.Float64s(windowVals)

	// Count values below latest
	below := 0
	for _, v := range windowVals {
		if v < latest {
			below++
		}
	}
	return float64(below) / float64(window) * 100.0
}

// computeCrowding scores how crowded the position is (0-1).
// High when z-score or percentile is at extremes.
func computeCrowding(sig *PositionSignal, params *storage.ModelParams) float64 {
	score := 0.0

	if sig.ZScore52W != nil {
		z := math.Abs(*sig.ZScore52W)
		threshold := math.Abs(params.CrowdedLongZScore)
		if z >= threshold {
			score = math.Max(score, 0.8+0.2*math.Min((z-threshold)/threshold, 1.0))
		} else if z >= threshold*0.5 {
			score = math.Max(score, 0.3+(z-threshold*0.5)/(threshold*0.5)*0.5)
		}
	}

	if sig.Percentile52W != nil {
		p := *sig.Percentile52W
		longThresh := params.CrowdedLongPercentile
		shortThresh := params.CrowdedShortPercentile

		if p >= longThresh {
			extreme := (p - longThresh) / (100 - longThresh)
			score = math.Max(score, 0.7+0.3*extreme)
		} else if p <= shortThresh {
			extreme := (shortThresh - p) / shortThresh
			score = math.Max(score, 0.7+0.3*extreme)
		}
	}

	return clampScore01(score)
}

// computeSqueezeRisk scores the probability of a short squeeze (0-1).
// High when: extreme short positioning + accelerating short covering (positive weekly change).
func computeSqueezeRisk(sig *PositionSignal, params *storage.ModelParams) float64 {
	// Only applies when net short
	if sig.NetManagedMoney >= 0 {
		return 0
	}

	score := 0.0

	// Extreme short positioning
	if sig.ZScore52W != nil && *sig.ZScore52W < params.CrowdedShortZScore {
		score += 0.4
	}
	if sig.Percentile52W != nil && *sig.Percentile52W < params.CrowdedShortPercentile {
		score += 0.3
	}

	// Accelerating short covering (positive weekly change when net short)
	if sig.WeeklyChangeNetMM != nil && *sig.WeeklyChangeNetMM > 0 {
		score += 0.3
	}

	return clampScore01(score)
}

// computeReversalRisk scores the probability of a position reversal (0-1).
// High when: extreme position (either direction) + decelerating momentum.
func computeReversalRisk(sig *PositionSignal, params *storage.ModelParams) float64 {
	score := 0.0

	// Check for extreme positioning
	isExtreme := false
	if sig.ZScore52W != nil {
		z := math.Abs(*sig.ZScore52W)
		if z >= math.Abs(params.CrowdedLongZScore) {
			isExtreme = true
			score += 0.4
		}
	}
	if sig.Percentile52W != nil {
		p := *sig.Percentile52W
		if p >= params.CrowdedLongPercentile || p <= params.CrowdedShortPercentile {
			isExtreme = true
			score += 0.2
		}
	}

	if !isExtreme {
		return 0
	}

	// Decelerating momentum: weekly change going against the position direction
	if sig.WeeklyChangeNetMM != nil {
		if sig.NetManagedMoney > 0 && *sig.WeeklyChangeNetMM < 0 {
			// Long position losing momentum
			score += 0.4
		} else if sig.NetManagedMoney < 0 && *sig.WeeklyChangeNetMM > 0 {
			// Short position losing momentum
			score += 0.4
		}
	}

	return clampScore01(score)
}

// computeTrendSupport scores how well the positioning trend is supported (0-1).
// High when: moderate position with consistent direction over recent weeks.
func computeTrendSupport(sig *PositionSignal, netMMHistory []float64) float64 {
	if len(netMMHistory) < 8 {
		return 0
	}

	// Check directional consistency of the last 8 weeks
	recent := netMMHistory[len(netMMHistory)-8:]
	latest := recent[len(recent)-1]

	sameDirection := 0
	for _, v := range recent {
		if (v > 0 && latest > 0) || (v < 0 && latest < 0) {
			sameDirection++
		}
	}

	consistency := float64(sameDirection) / float64(len(recent))

	// Moderate positioning (not extreme) + consistent direction = good trend support
	score := consistency * 0.7

	// Bonus if z-score is moderate (0.5-1.5 range)
	if sig.ZScore52W != nil {
		z := math.Abs(*sig.ZScore52W)
		if z >= 0.5 && z <= 1.5 {
			score += 0.3
		}
	}

	return clampScore01(score)
}

func clampScore01(v float64) float64 {
	if v > 1.0 {
		return 1.0
	}
	if v < 0.0 {
		return 0.0
	}
	return v
}
