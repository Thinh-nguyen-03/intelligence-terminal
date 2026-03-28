package analytics

import (
	"encoding/json"
	"testing"

	"github.com/0510t/intelligence-terminal/apps/api/internal/domain"
	"github.com/0510t/intelligence-terminal/apps/api/internal/storage"
)

func alertParams() *storage.ModelParams {
	return &storage.ModelParams{
		PositioningExtremeWeight:   0.40,
		AccelerationWeight:         0.25,
		MacroMismatchWeight:        0.20,
		ContinuationSupportWeight:  0.15,
		AlertThresholdCritical:     0.80,
		AlertThresholdWarning:      0.55,
		AlertThresholdInfo:         0.30,
		AlertCriticalConfidenceMin: 0.70,
		CrowdedLongZScore:         1.50,
		CrowdedShortZScore:        -1.50,
		CrowdedLongPercentile:     90,
		CrowdedShortPercentile:    10,
	}
}

func TestGenerateAlerts_CrowdedPosition(t *testing.T) {
	z52 := 2.0
	pct := 95.0
	change := int64(5000)

	sig := &PositionSignal{
		NetManagedMoney:   20000,
		NetMMPctOI:        15.0,
		ZScore52W:         &z52,
		Percentile52W:     &pct,
		WeeklyChangeNetMM: &change,
		CrowdingScore:     0.9,
		SqueezeRiskScore:  0.0,
		ReversalRiskScore: 0.3,
		TrendSupportScore: 0.4,
	}

	input := AlertInput{
		Commodity: domain.Commodity{
			ID:        1,
			Slug:      "gold",
			Name:      "Gold",
			GroupName: "metals",
		},
		Signal:           sig,
		RegimeLabel:      RegimeInflationarySlowdown,
		RegimeConfidence: 75,
		IsTransitioning:  false,
	}

	alerts := GenerateAlerts(input, alertParams())
	if len(alerts) == 0 {
		t.Fatal("expected at least one alert for crowded position")
	}

	alert := alerts[0]
	if alert.Severity == "" {
		t.Error("expected non-empty severity")
	}
	if alert.FinalAlertScore <= 0 {
		t.Errorf("expected positive final score, got %.3f", alert.FinalAlertScore)
	}
	if alert.Headline == "" {
		t.Error("expected non-empty headline")
	}
	if alert.Summary == "" {
		t.Error("expected non-empty summary")
	}

	// Verify explanation JSON is valid
	var explanation ExplanationPayload
	if err := json.Unmarshal(alert.ExplanationJSON, &explanation); err != nil {
		t.Fatalf("invalid explanation JSON: %v", err)
	}
	if len(explanation.Factors) != 4 {
		t.Errorf("expected 4 factors, got %d", len(explanation.Factors))
	}
	if explanation.RegimeContext.Label != RegimeInflationarySlowdown {
		t.Errorf("expected regime label in explanation, got %s", explanation.RegimeContext.Label)
	}
}

func TestGenerateAlerts_BelowThreshold(t *testing.T) {
	z52 := 0.3
	pct := 55.0
	change := int64(100)

	sig := &PositionSignal{
		NetManagedMoney:   1000,
		NetMMPctOI:        1.0,
		ZScore52W:         &z52,
		Percentile52W:     &pct,
		WeeklyChangeNetMM: &change,
		CrowdingScore:     0.1,
		SqueezeRiskScore:  0.0,
		ReversalRiskScore: 0.0,
		TrendSupportScore: 0.2,
	}

	input := AlertInput{
		Commodity: domain.Commodity{
			ID:        1,
			Slug:      "gold",
			Name:      "Gold",
			GroupName: "metals",
		},
		Signal:           sig,
		RegimeLabel:      RegimeRecoveryRiskOn,
		RegimeConfidence: 60,
	}

	alerts := GenerateAlerts(input, alertParams())
	if len(alerts) != 0 {
		t.Errorf("expected no alerts for mild positioning, got %d", len(alerts))
	}
}

func TestGenerateAlerts_NilSignal(t *testing.T) {
	input := AlertInput{
		Commodity: domain.Commodity{ID: 1, Slug: "gold", Name: "Gold", GroupName: "metals"},
		Signal:    nil,
	}
	alerts := GenerateAlerts(input, alertParams())
	if len(alerts) != 0 {
		t.Error("expected no alerts for nil signal")
	}
}

func TestGenerateAlerts_SqueezeCandidate(t *testing.T) {
	z52 := -2.0
	pct := 5.0
	change := int64(3000) // positive = short covering

	sig := &PositionSignal{
		NetManagedMoney:   -15000,
		NetMMPctOI:        -10.0,
		ZScore52W:         &z52,
		Percentile52W:     &pct,
		WeeklyChangeNetMM: &change,
		CrowdingScore:     0.85,
		SqueezeRiskScore:  0.9,
		ReversalRiskScore: 0.2,
		TrendSupportScore: 0.1,
	}

	input := AlertInput{
		Commodity: domain.Commodity{
			ID:        3,
			Slug:      "wti-crude",
			Name:      "WTI Crude Oil",
			GroupName: "energy",
		},
		Signal:           sig,
		RegimeLabel:      RegimeRecoveryRiskOn,
		RegimeConfidence: 80,
		IsTransitioning:  false,
	}

	alerts := GenerateAlerts(input, alertParams())
	if len(alerts) == 0 {
		t.Fatal("expected alert for squeeze candidate")
	}
	if alerts[0].AlertType != "squeeze_candidate" {
		t.Errorf("expected squeeze_candidate alert type, got %s", alerts[0].AlertType)
	}
}

func TestGenerateAlerts_TransitionWarning(t *testing.T) {
	z52 := 1.8
	pct := 85.0
	change := int64(5000)

	sig := &PositionSignal{
		NetManagedMoney:   10000,
		NetMMPctOI:        8.0,
		ZScore52W:         &z52,
		Percentile52W:     &pct,
		WeeklyChangeNetMM: &change,
		CrowdingScore:     0.7,
		SqueezeRiskScore:  0.0,
		ReversalRiskScore: 0.2,
		TrendSupportScore: 0.5,
	}

	input := AlertInput{
		Commodity: domain.Commodity{
			ID:        1,
			Slug:      "gold",
			Name:      "Gold",
			GroupName: "metals",
		},
		Signal:           sig,
		RegimeLabel:      "Inflationary Growth (transitioning)",
		RegimeConfidence: 50,
		IsTransitioning:  true,
	}

	alerts := GenerateAlerts(input, alertParams())
	if len(alerts) == 0 {
		t.Fatal("expected alert during regime transition")
	}
	if alerts[0].AlertType != "transition_warning" {
		t.Errorf("expected transition_warning type, got %s", alerts[0].AlertType)
	}
}

func TestClassifySeverity(t *testing.T) {
	params := alertParams()

	tests := []struct {
		name       string
		score      float64
		confidence float64
		transition bool
		want       domain.Severity
	}{
		{"critical", 0.85, 80, false, domain.SeverityCritical},
		{"critical needs confidence", 0.85, 50, false, domain.SeverityWarning},
		{"warning", 0.60, 60, false, domain.SeverityWarning},
		{"info", 0.35, 60, false, domain.SeverityInfo},
		{"transition bumps to warning", 0.35, 60, true, domain.SeverityWarning},
		{"below threshold", 0.20, 60, false, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := classifySeverity(tt.score, tt.confidence, tt.transition, params)
			if got != tt.want {
				t.Errorf("expected %q, got %q", tt.want, got)
			}
		})
	}
}
