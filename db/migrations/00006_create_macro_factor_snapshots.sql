-- +goose Up
CREATE TABLE macro_factor_snapshots (
    id BIGSERIAL PRIMARY KEY,
    as_of_date DATE NOT NULL,
    growth_score DOUBLE PRECISION NOT NULL,
    inflation_score DOUBLE PRECISION NOT NULL,
    labor_score DOUBLE PRECISION NOT NULL,
    stress_score DOUBLE PRECISION NOT NULL,
    regime_label TEXT NOT NULL,
    regime_prior_label TEXT,
    confidence DOUBLE PRECISION NOT NULL CHECK (confidence >= 0 AND confidence <= 100),
    is_transitioning BOOLEAN NOT NULL DEFAULT false,
    transition_detail TEXT,
    model_version TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX idx_macro_factor_as_of ON macro_factor_snapshots (as_of_date, model_version);
CREATE INDEX idx_macro_factor_date ON macro_factor_snapshots (as_of_date DESC);

-- +goose Down
DROP TABLE IF EXISTS macro_factor_snapshots;
