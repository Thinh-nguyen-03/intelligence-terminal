package analytics

// RegimeImpact represents how a regime affects a specific commodity group.
type RegimeImpact string

const (
	ImpactStronglyBullish RegimeImpact = "strongly_bullish"
	ImpactBullish         RegimeImpact = "bullish"
	ImpactMixed           RegimeImpact = "mixed"
	ImpactBearish         RegimeImpact = "bearish"
	ImpactStronglyBearish RegimeImpact = "strongly_bearish"
)

// CommodityImpact holds the regime impact for a specific commodity.
type CommodityImpact struct {
	Impact          RegimeImpact
	Description     string
	MismatchScore   float64 // 0-1: how much positioning conflicts with regime direction
	SupportScore    float64 // 0-1: how much regime supports current positioning
}

// regimeGroupMatrix maps (regime, group) to impact.
// Key: regime label (without transition modifier) + "|" + group_name.
var regimeGroupMatrix = map[string]map[string]RegimeImpact{
	RegimeInflationaryGrowth: {
		"metals": ImpactMixed,   // gold mixed, silver/copper bullish → overall mixed
		"energy": ImpactBullish,
	},
	RegimeInflationarySlowdown: {
		"metals": ImpactBullish, // gold bullish, silver mixed, copper bearish → net bullish
		"energy": ImpactMixed,
	},
	RegimeDisinflationSlowdown: {
		"metals": ImpactMixed,   // gold bullish, copper bearish → mixed
		"energy": ImpactBearish,
	},
	RegimeRecoveryRiskOn: {
		"metals": ImpactMixed,   // gold bearish, copper bullish → mixed
		"energy": ImpactBullish,
	},
	RegimeCreditStressDefensive: {
		"metals": ImpactBullish, // gold strongly bullish, copper strongly bearish → net bullish (safe haven)
		"energy": ImpactBearish,
	},
}

// Per-commodity overrides where the commodity diverges from its group.
var commodityOverrides = map[string]map[string]RegimeImpact{
	RegimeInflationaryGrowth: {
		"gold":   ImpactMixed,
		"silver": ImpactBullish,
		"copper": ImpactBullish,
	},
	RegimeInflationarySlowdown: {
		"gold":   ImpactBullish,
		"silver": ImpactMixed,
		"copper": ImpactBearish,
	},
	RegimeDisinflationSlowdown: {
		"gold":   ImpactBullish,
		"copper": ImpactBearish,
	},
	RegimeRecoveryRiskOn: {
		"gold":   ImpactBearish,
		"copper": ImpactBullish,
	},
	RegimeCreditStressDefensive: {
		"gold":   ImpactStronglyBullish,
		"copper": ImpactStronglyBearish,
	},
}

// GetRegimeImpact returns the impact of the current regime on a specific commodity.
// Uses per-commodity overrides when available, otherwise falls back to group-level impact.
func GetRegimeImpact(regimeLabel, commoditySlug, groupName string) RegimeImpact {
	// Strip transition modifier
	baseRegime := stripTransition(regimeLabel)

	// Check per-commodity override first
	if overrides, ok := commodityOverrides[baseRegime]; ok {
		if impact, ok := overrides[commoditySlug]; ok {
			return impact
		}
	}

	// Fall back to group-level impact
	if group, ok := regimeGroupMatrix[baseRegime]; ok {
		if impact, ok := group[groupName]; ok {
			return impact
		}
	}

	return ImpactMixed // default when no mapping exists
}

// ComputeRegimeMismatch computes how much the current positioning
// conflicts with the regime-implied direction for this commodity.
// Returns 0-1 where 1 = maximum mismatch.
func ComputeRegimeMismatch(netMM int64, impact RegimeImpact) float64 {
	isLong := netMM > 0
	isShort := netMM < 0

	switch impact {
	case ImpactStronglyBullish, ImpactBullish:
		if isShort {
			return 0.8 // short against bullish regime
		}
		return 0.0
	case ImpactStronglyBearish, ImpactBearish:
		if isLong {
			return 0.8 // long against bearish regime
		}
		return 0.0
	case ImpactMixed:
		return 0.1 // mild uncertainty in mixed environments
	}
	return 0.0
}

// ComputeContinuationSupport computes how much the regime supports
// the current positioning direction.
// Returns 0-1 where 1 = strong regime support for continuation.
func ComputeContinuationSupport(netMM int64, impact RegimeImpact, regimeConfidence float64) float64 {
	isLong := netMM > 0
	isShort := netMM < 0
	confFactor := regimeConfidence / 100.0

	switch impact {
	case ImpactStronglyBullish:
		if isLong {
			return 0.9 * confFactor
		}
		return 0.0
	case ImpactBullish:
		if isLong {
			return 0.7 * confFactor
		}
		return 0.0
	case ImpactStronglyBearish:
		if isShort {
			return 0.9 * confFactor
		}
		return 0.0
	case ImpactBearish:
		if isShort {
			return 0.7 * confFactor
		}
		return 0.0
	case ImpactMixed:
		return 0.2 * confFactor
	}
	return 0.0
}

// stripTransition removes the "(transitioning)" suffix from a regime label.
func stripTransition(label string) string {
	const suffix = " (transitioning)"
	if len(label) > len(suffix) && label[len(label)-len(suffix):] == suffix {
		return label[:len(label)-len(suffix)]
	}
	return label
}

// ImpactDirectionString returns a human-readable direction for the impact.
func ImpactDirectionString(impact RegimeImpact) string {
	switch impact {
	case ImpactStronglyBullish:
		return "strongly bullish"
	case ImpactBullish:
		return "bullish"
	case ImpactBearish:
		return "bearish"
	case ImpactStronglyBearish:
		return "strongly bearish"
	default:
		return "mixed"
	}
}
