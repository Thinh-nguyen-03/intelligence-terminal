-- Seed: default model configuration
-- These weights and thresholds are configurable without redeployment

INSERT INTO model_config (config_key, config_value, description) VALUES
    -- Combined alert model weights (Section 12.1)
    ('positioning_extreme_weight',     '0.40', 'Weight for positioning extreme score in final alert'),
    ('acceleration_weight',            '0.25', 'Weight for position acceleration score in final alert'),
    ('macro_mismatch_weight',          '0.20', 'Weight for macro regime mismatch score in final alert'),
    ('continuation_support_weight',    '0.15', 'Weight for regime continuation support in final alert'),

    -- Alert severity thresholds (Section 12.2)
    ('alert_threshold_critical',       '0.80', 'Minimum final_alert_score for critical severity'),
    ('alert_threshold_warning',        '0.55', 'Minimum final_alert_score for warning severity'),
    ('alert_threshold_info',           '0.30', 'Minimum final_alert_score for info severity'),
    ('alert_critical_confidence_min',  '0.70', 'Minimum regime confidence for critical alerts'),

    -- Regime factor score ranges (Section 11.1)
    ('factor_strong_positive',         '0.30',  'Threshold for strong positive factor reading'),
    ('factor_strong_negative',         '-0.30', 'Threshold for strong negative factor reading'),
    ('factor_overlap_upper',           '0.15',  'Upper bound of neutral/overlap zone'),
    ('factor_overlap_lower',           '-0.15', 'Lower bound of neutral/overlap zone'),
    ('stress_override_threshold',      '0.50',  'Stress score above this overrides regime to Credit Stress'),

    -- Positioning thresholds (Section 11.2)
    ('crowded_long_zscore',            '1.50',  'Z-score threshold for crowded long label'),
    ('crowded_short_zscore',           '-1.50', 'Z-score threshold for crowded short label'),
    ('crowded_long_percentile',        '90',    'Percentile threshold for crowded long label'),
    ('crowded_short_percentile',       '10',    'Percentile threshold for crowded short label'),

    -- Model versioning
    ('current_model_version',          'v1.0',  'Current analytics model version tag')
ON CONFLICT (config_key) DO UPDATE SET
    config_value = EXCLUDED.config_value,
    updated_at = now();
