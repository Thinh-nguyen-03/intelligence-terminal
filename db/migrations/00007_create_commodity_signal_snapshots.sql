-- +goose Up
CREATE TABLE commodity_signal_snapshots (
    id BIGSERIAL PRIMARY KEY,
    commodity_id BIGINT NOT NULL REFERENCES commodities(id),
    as_of_date DATE NOT NULL,
    net_managed_money BIGINT NOT NULL,
    net_mm_pct_oi DOUBLE PRECISION NOT NULL,
    position_zscore_26w DOUBLE PRECISION,
    position_zscore_52w DOUBLE PRECISION,
    position_percentile_52w DOUBLE PRECISION,
    weekly_change_net_mm BIGINT,
    crowding_score DOUBLE PRECISION NOT NULL,
    squeeze_risk_score DOUBLE PRECISION NOT NULL,
    reversal_risk_score DOUBLE PRECISION NOT NULL,
    trend_support_score DOUBLE PRECISION NOT NULL,
    model_version TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX idx_commodity_signal_commodity_date ON commodity_signal_snapshots (commodity_id, as_of_date, model_version);
CREATE INDEX idx_commodity_signal_date ON commodity_signal_snapshots (as_of_date DESC);

-- +goose Down
DROP TABLE IF EXISTS commodity_signal_snapshots;
