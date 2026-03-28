package analytics

import (
	"fmt"
	"math"
	"strings"

	"github.com/0510t/intelligence-terminal/apps/api/internal/storage"
)

// Regime labels
const (
	RegimeInflationaryGrowth    = "Inflationary Growth"
	RegimeInflationarySlowdown  = "Inflationary Slowdown"
	RegimeDisinflationSlowdown  = "Disinflationary Slowdown"
	RegimeRecoveryRiskOn        = "Recovery / Risk-On"
	RegimeCreditStressDefensive = "Credit Stress / Defensive"
)

// RegimeResult holds the output of regime classification.
type RegimeResult struct {
	Label            string
	Confidence       float64 // 0-100
	IsTransitioning  bool
	TransitionDetail string
}

// ClassifyRegime determines the macro regime from factor scores using
// overlapping score ranges and transition state detection.
func ClassifyRegime(growth, inflation, stress float64, params *storage.ModelParams) RegimeResult {
	// Stress override: if stress is very high, override everything
	if stress > params.StressOverride {
		conf := 70.0 + (stress-params.StressOverride)*60.0
		if conf > 95 {
			conf = 95
		}
		return RegimeResult{
			Label:      RegimeCreditStressDefensive,
			Confidence: conf,
		}
	}

	// Classify based on growth/inflation quadrant
	label := classifyQuadrant(growth, inflation)

	// Detect transition state: if either score is in the overlap zone
	growthInOverlap := growth >= params.FactorOverlapLower && growth <= params.FactorOverlapUpper
	inflationInOverlap := inflation >= params.FactorOverlapLower && inflation <= params.FactorOverlapUpper

	isTransitioning := growthInOverlap || inflationInOverlap
	transitionDetail := ""

	if isTransitioning {
		details := make([]string, 0, 2)
		if growthInOverlap {
			details = append(details, fmt.Sprintf("growth near boundary (%.2f)", growth))
		}
		if inflationInOverlap {
			details = append(details, fmt.Sprintf("inflation near boundary (%.2f)", inflation))
		}
		transitionDetail = strings.Join(details, "; ")
	}

	// Compute confidence based on distance from boundaries
	confidence := computeConfidence(growth, inflation, params)
	if isTransitioning {
		// Reduce confidence when transitioning
		confidence *= 0.75
	}

	// Clamp confidence
	if confidence < 20 {
		confidence = 20
	}
	if confidence > 95 {
		confidence = 95
	}

	result := RegimeResult{
		Label:            label,
		Confidence:       math.Round(confidence),
		IsTransitioning:  isTransitioning,
		TransitionDetail: transitionDetail,
	}

	if isTransitioning {
		result.Label = label + " (transitioning)"
	}

	return result
}

// classifyQuadrant maps growth/inflation signs to a regime label.
func classifyQuadrant(growth, inflation float64) string {
	growthPositive := growth > 0
	inflationPositive := inflation > 0

	switch {
	case growthPositive && inflationPositive:
		return RegimeInflationaryGrowth
	case !growthPositive && inflationPositive:
		return RegimeInflationarySlowdown
	case !growthPositive && !inflationPositive:
		return RegimeDisinflationSlowdown
	default: // growthPositive && !inflationPositive
		return RegimeRecoveryRiskOn
	}
}

// computeConfidence returns a confidence score (0-100) based on how far
// factor scores are from the zero boundary.
func computeConfidence(growth, inflation float64, params *storage.ModelParams) float64 {
	// Distance of each factor from zero, normalized by the strong threshold
	growthDist := math.Abs(growth) / math.Abs(params.FactorStrongPositive)
	inflationDist := math.Abs(inflation) / math.Abs(params.FactorStrongPositive)

	// Average distance, capped at 1.0
	avgDist := (growthDist + inflationDist) / 2.0
	if avgDist > 1.0 {
		avgDist = 1.0
	}

	// Map to 40-95 range: very clear signals → high confidence
	return 40.0 + avgDist*55.0
}
