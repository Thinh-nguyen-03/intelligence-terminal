package analytics

import (
	"strings"
	"testing"

	"github.com/0510t/intelligence-terminal/apps/api/internal/storage"
)

func defaultParams() *storage.ModelParams {
	return &storage.ModelParams{
		FactorStrongPositive: 0.30,
		FactorStrongNegative: -0.30,
		FactorOverlapUpper:   0.15,
		FactorOverlapLower:   -0.15,
		StressOverride:       0.50,
		ModelVersion:         "v1.0-test",
	}
}

func TestClassifyRegime_InflationaryGrowth(t *testing.T) {
	r := ClassifyRegime(0.5, 0.5, 0.1, defaultParams())
	if r.Label != RegimeInflationaryGrowth {
		t.Errorf("expected %s, got %s", RegimeInflationaryGrowth, r.Label)
	}
	if r.IsTransitioning {
		t.Error("should not be transitioning with strong signals")
	}
	if r.Confidence < 60 {
		t.Errorf("expected high confidence, got %.0f", r.Confidence)
	}
}

func TestClassifyRegime_InflationarySlowdown(t *testing.T) {
	r := ClassifyRegime(-0.4, 0.6, 0.2, defaultParams())
	if r.Label != RegimeInflationarySlowdown {
		t.Errorf("expected %s, got %s", RegimeInflationarySlowdown, r.Label)
	}
}

func TestClassifyRegime_DisinflationarySlowdown(t *testing.T) {
	r := ClassifyRegime(-0.5, -0.5, 0.1, defaultParams())
	if r.Label != RegimeDisinflationSlowdown {
		t.Errorf("expected %s, got %s", RegimeDisinflationSlowdown, r.Label)
	}
}

func TestClassifyRegime_RecoveryRiskOn(t *testing.T) {
	r := ClassifyRegime(0.6, -0.4, 0.1, defaultParams())
	if r.Label != RegimeRecoveryRiskOn {
		t.Errorf("expected %s, got %s", RegimeRecoveryRiskOn, r.Label)
	}
}

func TestClassifyRegime_StressOverride(t *testing.T) {
	// Even with positive growth/inflation, high stress overrides
	r := ClassifyRegime(0.5, 0.5, 0.7, defaultParams())
	if r.Label != RegimeCreditStressDefensive {
		t.Errorf("expected %s, got %s", RegimeCreditStressDefensive, r.Label)
	}
	if r.IsTransitioning {
		t.Error("stress override should not be transitioning")
	}
}

func TestClassifyRegime_TransitionGrowth(t *testing.T) {
	// Growth in overlap zone [-0.15, 0.15]
	r := ClassifyRegime(0.05, 0.5, 0.1, defaultParams())
	if !r.IsTransitioning {
		t.Error("expected transitioning when growth is in overlap zone")
	}
	if !strings.Contains(r.Label, "(transitioning)") {
		t.Errorf("expected transitioning label, got %s", r.Label)
	}
	if !strings.Contains(r.TransitionDetail, "growth near boundary") {
		t.Errorf("expected growth transition detail, got %q", r.TransitionDetail)
	}
}

func TestClassifyRegime_TransitionInflation(t *testing.T) {
	// Inflation in overlap zone
	r := ClassifyRegime(0.5, 0.10, 0.1, defaultParams())
	if !r.IsTransitioning {
		t.Error("expected transitioning when inflation is in overlap zone")
	}
	if !strings.Contains(r.TransitionDetail, "inflation near boundary") {
		t.Errorf("expected inflation transition detail, got %q", r.TransitionDetail)
	}
}

func TestClassifyRegime_TransitionBoth(t *testing.T) {
	// Both in overlap zone
	r := ClassifyRegime(0.05, -0.05, 0.1, defaultParams())
	if !r.IsTransitioning {
		t.Error("expected transitioning")
	}
	if !strings.Contains(r.TransitionDetail, "growth") || !strings.Contains(r.TransitionDetail, "inflation") {
		t.Errorf("expected both factors in transition detail, got %q", r.TransitionDetail)
	}
}

func TestClassifyRegime_TransitionReducesConfidence(t *testing.T) {
	params := defaultParams()

	// Same absolute scores, one transitioning, one not
	clearResult := ClassifyRegime(0.5, 0.5, 0.1, params)
	transResult := ClassifyRegime(0.05, 0.5, 0.1, params)

	if transResult.Confidence >= clearResult.Confidence {
		t.Errorf("transitioning confidence (%.0f) should be lower than clear (%.0f)",
			transResult.Confidence, clearResult.Confidence)
	}
}

func TestClassifyRegime_ConfidenceBounds(t *testing.T) {
	params := defaultParams()

	// Very weak signals
	r := ClassifyRegime(0.01, 0.01, 0.0, params)
	if r.Confidence < 20 || r.Confidence > 95 {
		t.Errorf("confidence %.0f out of [20, 95] range", r.Confidence)
	}

	// Very strong signals
	r = ClassifyRegime(1.0, 1.0, 0.0, params)
	if r.Confidence < 20 || r.Confidence > 95 {
		t.Errorf("confidence %.0f out of [20, 95] range", r.Confidence)
	}
}
