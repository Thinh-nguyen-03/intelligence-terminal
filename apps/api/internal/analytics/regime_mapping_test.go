package analytics

import "testing"

func TestGetRegimeImpact_CommodityOverride(t *testing.T) {
	// Gold in Credit Stress should be strongly bullish (override, not group default)
	impact := GetRegimeImpact(RegimeCreditStressDefensive, "gold", "metals")
	if impact != ImpactStronglyBullish {
		t.Errorf("expected strongly_bullish for gold in credit stress, got %s", impact)
	}

	// Copper in Credit Stress should be strongly bearish
	impact = GetRegimeImpact(RegimeCreditStressDefensive, "copper", "metals")
	if impact != ImpactStronglyBearish {
		t.Errorf("expected strongly_bearish for copper in credit stress, got %s", impact)
	}
}

func TestGetRegimeImpact_GroupFallback(t *testing.T) {
	// Silver in inflationary growth has no override → falls back to metals group (mixed)
	impact := GetRegimeImpact(RegimeInflationaryGrowth, "silver", "metals")
	// Silver has an override: bullish
	if impact != ImpactBullish {
		t.Errorf("expected bullish for silver in inflationary growth, got %s", impact)
	}

	// Natural gas in recovery → falls back to energy group (bullish)
	impact = GetRegimeImpact(RegimeRecoveryRiskOn, "natural-gas", "energy")
	if impact != ImpactBullish {
		t.Errorf("expected bullish for natural gas in recovery, got %s", impact)
	}
}

func TestGetRegimeImpact_TransitionStripped(t *testing.T) {
	// Transitioning label should be stripped
	impact := GetRegimeImpact("Inflationary Growth (transitioning)", "gold", "metals")
	if impact != ImpactMixed {
		t.Errorf("expected mixed for gold in inflationary growth, got %s", impact)
	}
}

func TestGetRegimeImpact_UnknownRegime(t *testing.T) {
	impact := GetRegimeImpact("Unknown Regime", "gold", "metals")
	if impact != ImpactMixed {
		t.Errorf("expected mixed for unknown regime, got %s", impact)
	}
}

func TestComputeRegimeMismatch(t *testing.T) {
	tests := []struct {
		name   string
		netMM  int64
		impact RegimeImpact
		expect float64
	}{
		{"long against bearish", 5000, ImpactBearish, 0.8},
		{"short against bullish", -5000, ImpactBullish, 0.8},
		{"long with bullish", 5000, ImpactBullish, 0.0},
		{"short with bearish", -5000, ImpactBearish, 0.0},
		{"any in mixed", 5000, ImpactMixed, 0.1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ComputeRegimeMismatch(tt.netMM, tt.impact)
			if got != tt.expect {
				t.Errorf("expected %.1f, got %.1f", tt.expect, got)
			}
		})
	}
}

func TestComputeContinuationSupport(t *testing.T) {
	// Long position with strongly bullish regime at 80% confidence
	score := ComputeContinuationSupport(10000, ImpactStronglyBullish, 80)
	if score < 0.5 {
		t.Errorf("expected high support for long+strongly_bullish, got %.2f", score)
	}

	// Short position with bullish regime → no support
	score = ComputeContinuationSupport(-10000, ImpactBullish, 80)
	if score != 0.0 {
		t.Errorf("expected 0 support for short against bullish, got %.2f", score)
	}

	// Mixed regime → low support
	score = ComputeContinuationSupport(10000, ImpactMixed, 80)
	if score > 0.3 {
		t.Errorf("expected low support for mixed regime, got %.2f", score)
	}
}

func TestStripTransition(t *testing.T) {
	tests := []struct {
		input, want string
	}{
		{"Inflationary Growth (transitioning)", "Inflationary Growth"},
		{"Recovery / Risk-On", "Recovery / Risk-On"},
		{"Credit Stress / Defensive", "Credit Stress / Defensive"},
	}
	for _, tt := range tests {
		got := stripTransition(tt.input)
		if got != tt.want {
			t.Errorf("stripTransition(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
