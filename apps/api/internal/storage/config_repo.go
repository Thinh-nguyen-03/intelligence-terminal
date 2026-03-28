package storage

import (
	"context"
	"fmt"
	"strconv"

	"github.com/jackc/pgx/v5/pgxpool"
)

// ConfigRepo reads model_config values for the analytics engine.
type ConfigRepo struct {
	pool *pgxpool.Pool
}

func NewConfigRepo(pool *pgxpool.Pool) *ConfigRepo {
	return &ConfigRepo{pool: pool}
}

// LoadAll returns all config key-value pairs as a map.
func (r *ConfigRepo) LoadAll(ctx context.Context) (map[string]string, error) {
	rows, err := r.pool.Query(ctx, `SELECT config_key, config_value FROM model_config`)
	if err != nil {
		return nil, fmt.Errorf("querying model_config: %w", err)
	}
	defer rows.Close()

	cfg := make(map[string]string)
	for rows.Next() {
		var k, v string
		if err := rows.Scan(&k, &v); err != nil {
			return nil, fmt.Errorf("scanning config row: %w", err)
		}
		cfg[k] = v
	}
	return cfg, rows.Err()
}

// ModelParams holds parsed model configuration for the analytics engine.
type ModelParams struct {
	// Regime factor thresholds
	FactorStrongPositive float64
	FactorStrongNegative float64
	FactorOverlapUpper   float64
	FactorOverlapLower   float64
	StressOverride       float64

	// Positioning thresholds
	CrowdedLongZScore      float64
	CrowdedShortZScore     float64
	CrowdedLongPercentile  float64
	CrowdedShortPercentile float64

	// Alert weights
	PositioningExtremeWeight  float64
	AccelerationWeight        float64
	MacroMismatchWeight       float64
	ContinuationSupportWeight float64

	// Alert severity thresholds
	AlertThresholdCritical      float64
	AlertThresholdWarning       float64
	AlertThresholdInfo          float64
	AlertCriticalConfidenceMin  float64

	// Version
	ModelVersion string
}

// LoadModelParams loads and parses all model parameters from the database.
func (r *ConfigRepo) LoadModelParams(ctx context.Context) (*ModelParams, error) {
	raw, err := r.LoadAll(ctx)
	if err != nil {
		return nil, err
	}

	p := &ModelParams{
		FactorStrongPositive:       parseFloat(raw, "factor_strong_positive", 0.30),
		FactorStrongNegative:       parseFloat(raw, "factor_strong_negative", -0.30),
		FactorOverlapUpper:         parseFloat(raw, "factor_overlap_upper", 0.15),
		FactorOverlapLower:         parseFloat(raw, "factor_overlap_lower", -0.15),
		StressOverride:             parseFloat(raw, "stress_override_threshold", 0.50),
		CrowdedLongZScore:          parseFloat(raw, "crowded_long_zscore", 1.50),
		CrowdedShortZScore:         parseFloat(raw, "crowded_short_zscore", -1.50),
		CrowdedLongPercentile:      parseFloat(raw, "crowded_long_percentile", 90),
		CrowdedShortPercentile:     parseFloat(raw, "crowded_short_percentile", 10),
		PositioningExtremeWeight:   parseFloat(raw, "positioning_extreme_weight", 0.40),
		AccelerationWeight:         parseFloat(raw, "acceleration_weight", 0.25),
		MacroMismatchWeight:        parseFloat(raw, "macro_mismatch_weight", 0.20),
		ContinuationSupportWeight:  parseFloat(raw, "continuation_support_weight", 0.15),
		AlertThresholdCritical:     parseFloat(raw, "alert_threshold_critical", 0.80),
		AlertThresholdWarning:      parseFloat(raw, "alert_threshold_warning", 0.55),
		AlertThresholdInfo:         parseFloat(raw, "alert_threshold_info", 0.30),
		AlertCriticalConfidenceMin: parseFloat(raw, "alert_critical_confidence_min", 0.70),
		ModelVersion:               raw["current_model_version"],
	}
	if p.ModelVersion == "" {
		p.ModelVersion = "v1.0"
	}

	return p, nil
}

func parseFloat(m map[string]string, key string, fallback float64) float64 {
	s, ok := m[key]
	if !ok {
		return fallback
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return fallback
	}
	return v
}
