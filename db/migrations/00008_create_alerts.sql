-- +goose Up
CREATE TYPE severity_enum AS ENUM ('critical', 'warning', 'info');

CREATE TABLE alerts (
    id BIGSERIAL PRIMARY KEY,
    commodity_id BIGINT NOT NULL REFERENCES commodities(id),
    as_of_date DATE NOT NULL,
    severity severity_enum NOT NULL,
    alert_type TEXT NOT NULL,
    headline TEXT NOT NULL,
    summary TEXT NOT NULL,
    explanation_json JSONB NOT NULL,
    regime_label TEXT NOT NULL,
    regime_confidence DOUBLE PRECISION NOT NULL,
    final_alert_score DOUBLE PRECISION NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_alerts_active_severity ON alerts (is_active, severity, as_of_date DESC);
CREATE INDEX idx_alerts_commodity ON alerts (commodity_id, as_of_date DESC);

-- +goose Down
DROP TABLE IF EXISTS alerts;
DROP TYPE IF EXISTS severity_enum;
