package analytics

import (
	"encoding/json"
	"fmt"
	"math"

	"github.com/0510t/intelligence-terminal/apps/api/internal/domain"
	"github.com/0510t/intelligence-terminal/apps/api/internal/storage"
)

// AlertInput bundles everything needed to evaluate alerts for a commodity.
type AlertInput struct {
	Commodity       domain.Commodity
	Signal          *PositionSignal
	RegimeLabel     string
	RegimeConfidence float64
	IsTransitioning bool
}

// AlertOutput is a generated alert ready for storage.
type AlertOutput struct {
	CommodityID      int64
	Severity         domain.Severity
	AlertType        string
	Headline         string
	Summary          string
	ExplanationJSON  json.RawMessage
	RegimeLabel      string
	RegimeConfidence float64
	FinalAlertScore  float64
}

// ExplanationPayload is the structured explanation stored as JSON.
type ExplanationPayload struct {
	Factors           []ExplanationFactor `json:"factors"`
	RegimeContext     RegimeContext        `json:"regime_context"`
	PositioningContext PositioningContext  `json:"positioning_context"`
}

type ExplanationFactor struct {
	Name         string  `json:"name"`
	Value        float64 `json:"value"`
	Weight       float64 `json:"weight"`
	Contribution float64 `json:"contribution"`
	Detail       string  `json:"detail"`
}

type RegimeContext struct {
	Label            string `json:"label"`
	Confidence       float64 `json:"confidence"`
	GroupImpact      string  `json:"commodity_group_impact"`
	IsTransitioning  bool    `json:"is_transitioning"`
}

type PositioningContext struct {
	NetMMPctOI    float64  `json:"net_mm_pct_oi"`
	ZScore52W     *float64 `json:"zscore_52w"`
	Percentile52W *float64 `json:"percentile_52w"`
	WeeklyChange  *int64   `json:"weekly_change"`
}

// GenerateAlerts evaluates a commodity's positioning against the regime
// and returns any alerts that exceed the configured thresholds.
func GenerateAlerts(input AlertInput, params *storage.ModelParams) []AlertOutput {
	sig := input.Signal
	if sig == nil {
		return nil
	}

	impact := GetRegimeImpact(input.RegimeLabel, input.Commodity.Slug, input.Commodity.GroupName)

	// Compute the four score components
	positioningExtreme := sig.CrowdingScore
	acceleration := computeAcceleration(sig)
	mismatch := ComputeRegimeMismatch(sig.NetManagedMoney, impact)
	continuation := ComputeContinuationSupport(sig.NetManagedMoney, impact, input.RegimeConfidence)

	// Combined score (Section 12.1)
	finalScore := params.PositioningExtremeWeight*positioningExtreme +
		params.AccelerationWeight*acceleration +
		params.MacroMismatchWeight*mismatch +
		params.ContinuationSupportWeight*continuation

	// Determine severity
	severity := classifySeverity(finalScore, input.RegimeConfidence, input.IsTransitioning, params)
	if severity == "" {
		return nil // below info threshold
	}

	// Determine alert type and generate explanation
	alertType := classifyAlertType(sig, impact, input.IsTransitioning)
	headline := generateHeadline(input.Commodity.Name, alertType, sig)
	summary := generateSummary(input, sig, impact, alertType)

	explanation := ExplanationPayload{
		Factors: []ExplanationFactor{
			{
				Name:         "positioning_extreme",
				Value:        positioningExtreme,
				Weight:       params.PositioningExtremeWeight,
				Contribution: params.PositioningExtremeWeight * positioningExtreme,
				Detail:       fmt.Sprintf("crowding=%.2f", sig.CrowdingScore),
			},
			{
				Name:         "acceleration",
				Value:        acceleration,
				Weight:       params.AccelerationWeight,
				Contribution: params.AccelerationWeight * acceleration,
				Detail:       accelerationDetail(sig),
			},
			{
				Name:         "macro_regime_mismatch",
				Value:        mismatch,
				Weight:       params.MacroMismatchWeight,
				Contribution: params.MacroMismatchWeight * mismatch,
				Detail:       fmt.Sprintf("impact=%s, net_mm=%d", ImpactDirectionString(impact), sig.NetManagedMoney),
			},
			{
				Name:         "continuation_support",
				Value:        continuation,
				Weight:       params.ContinuationSupportWeight,
				Contribution: params.ContinuationSupportWeight * continuation,
				Detail:       fmt.Sprintf("regime_confidence=%.0f%%", input.RegimeConfidence),
			},
		},
		RegimeContext: RegimeContext{
			Label:           input.RegimeLabel,
			Confidence:      input.RegimeConfidence,
			GroupImpact:     ImpactDirectionString(impact),
			IsTransitioning: input.IsTransitioning,
		},
		PositioningContext: PositioningContext{
			NetMMPctOI:    sig.NetMMPctOI,
			ZScore52W:     sig.ZScore52W,
			Percentile52W: sig.Percentile52W,
			WeeklyChange:  sig.WeeklyChangeNetMM,
		},
	}

	explanationJSON, _ := json.Marshal(explanation)

	return []AlertOutput{{
		CommodityID:      input.Commodity.ID,
		Severity:         severity,
		AlertType:        alertType,
		Headline:         headline,
		Summary:          summary,
		ExplanationJSON:  explanationJSON,
		RegimeLabel:      input.RegimeLabel,
		RegimeConfidence: input.RegimeConfidence,
		FinalAlertScore:  finalScore,
	}}
}

func computeAcceleration(sig *PositionSignal) float64 {
	if sig.WeeklyChangeNetMM == nil {
		return 0
	}
	change := math.Abs(float64(*sig.WeeklyChangeNetMM))
	netAbs := math.Abs(float64(sig.NetManagedMoney))
	if netAbs == 0 {
		return 0
	}
	// Acceleration = weekly change as fraction of total position, normalized to 0-1
	ratio := change / netAbs
	return clampScore01(ratio * 2.0) // 50% weekly change → 1.0
}

func classifySeverity(score, regimeConfidence float64, isTransitioning bool, params *storage.ModelParams) domain.Severity {
	switch {
	case score > params.AlertThresholdCritical && regimeConfidence > params.AlertCriticalConfidenceMin*100:
		return domain.SeverityCritical
	case score > params.AlertThresholdWarning:
		return domain.SeverityWarning
	case score > params.AlertThresholdInfo && isTransitioning:
		// Lower bar during transitions (spec: score > 0.45 AND transitioning counts as warning)
		return domain.SeverityWarning
	case score > params.AlertThresholdInfo:
		return domain.SeverityInfo
	default:
		return "" // no alert
	}
}

func classifyAlertType(sig *PositionSignal, impact RegimeImpact, isTransitioning bool) string {
	if isTransitioning {
		return "transition_warning"
	}
	if sig.SqueezeRiskScore > 0.5 {
		return "squeeze_candidate"
	}
	if sig.ReversalRiskScore > 0.5 {
		return "reversal_risk"
	}
	if sig.CrowdingScore > 0.6 {
		isLong := sig.NetManagedMoney > 0
		isBullish := impact == ImpactBullish || impact == ImpactStronglyBullish
		isBearish := impact == ImpactBearish || impact == ImpactStronglyBearish
		if (isLong && isBearish) || (!isLong && isBullish) {
			return "crowded_regime_mismatch"
		}
		return "crowded_position"
	}
	if sig.TrendSupportScore > 0.6 {
		return "macro_supported_trend"
	}
	return "positioning_alert"
}

func generateHeadline(commodityName, alertType string, sig *PositionSignal) string {
	direction := "long"
	if sig.NetManagedMoney < 0 {
		direction = "short"
	}

	switch alertType {
	case "squeeze_candidate":
		return fmt.Sprintf("%s: Short squeeze risk elevated", commodityName)
	case "reversal_risk":
		return fmt.Sprintf("%s: %s reversal risk elevated", commodityName, direction)
	case "crowded_regime_mismatch":
		return fmt.Sprintf("%s: Crowded %s against regime direction", commodityName, direction)
	case "crowded_position":
		return fmt.Sprintf("%s: Crowded %s positioning", commodityName, direction)
	case "macro_supported_trend":
		return fmt.Sprintf("%s: Regime-supported %s trend", commodityName, direction)
	case "transition_warning":
		return fmt.Sprintf("%s: Regime transition — signals may shift", commodityName)
	default:
		return fmt.Sprintf("%s: Positioning alert", commodityName)
	}
}

func generateSummary(input AlertInput, sig *PositionSignal, impact RegimeImpact, alertType string) string {
	direction := "long"
	if sig.NetManagedMoney < 0 {
		direction = "short"
	}

	pctStr := ""
	if sig.Percentile52W != nil {
		pctStr = fmt.Sprintf("%.0fth percentile", *sig.Percentile52W)
	}
	zStr := ""
	if sig.ZScore52W != nil {
		zStr = fmt.Sprintf("%.1f z-score", *sig.ZScore52W)
	}

	impactStr := ImpactDirectionString(impact)
	group := input.Commodity.GroupName

	switch alertType {
	case "squeeze_candidate":
		s := fmt.Sprintf("%s managed money shorts are at extreme levels", input.Commodity.Name)
		if pctStr != "" {
			s += fmt.Sprintf(" (%s", pctStr)
			if zStr != "" {
				s += fmt.Sprintf(", %s", zStr)
			}
			s += ")"
		}
		if sig.WeeklyChangeNetMM != nil && *sig.WeeklyChangeNetMM > 0 {
			s += ". Short covering is accelerating"
		}
		s += ". Squeeze risk is high."
		return s

	case "crowded_regime_mismatch":
		s := fmt.Sprintf("%s is crowded %s", input.Commodity.Name, direction)
		if pctStr != "" {
			s += fmt.Sprintf(" (%s)", pctStr)
		}
		s += fmt.Sprintf(" while the macro regime (%s) is %s for %s. Reversal risk is elevated.",
			input.RegimeLabel, impactStr, group)
		return s

	case "macro_supported_trend":
		s := fmt.Sprintf("%s positioning is extended %s", input.Commodity.Name, direction)
		s += fmt.Sprintf(", but the %s regime supports %s exposure.", input.RegimeLabel, group)
		s += fmt.Sprintf(" Confidence: %.0f%%.", input.RegimeConfidence)
		return s

	case "transition_warning":
		s := fmt.Sprintf("Macro regime is transitioning. %s signals may shift — current %s positioning",
			input.Commodity.Name, direction)
		if pctStr != "" {
			s += fmt.Sprintf(" at %s", pctStr)
		}
		s += "."
		return s

	default:
		s := fmt.Sprintf("%s managed money is %s", input.Commodity.Name, direction)
		if pctStr != "" {
			s += fmt.Sprintf(" (%s)", pctStr)
		}
		s += fmt.Sprintf(". Regime: %s (%s for %s).", input.RegimeLabel, impactStr, group)
		return s
	}
}

func accelerationDetail(sig *PositionSignal) string {
	if sig.WeeklyChangeNetMM == nil {
		return "no weekly change data"
	}
	return fmt.Sprintf("weekly_change=%d", *sig.WeeklyChangeNetMM)
}
