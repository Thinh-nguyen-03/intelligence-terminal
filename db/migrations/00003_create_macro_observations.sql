-- +goose Up
CREATE TABLE macro_observations_raw (
    id BIGSERIAL PRIMARY KEY,
    series_id BIGINT NOT NULL REFERENCES macro_series(id),
    observation_date DATE NOT NULL,
    realtime_start DATE,
    realtime_end DATE,
    raw_value TEXT NOT NULL,
    payload_json JSONB,
    ingested_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_macro_obs_raw_series_date ON macro_observations_raw (series_id, observation_date);

CREATE TABLE macro_observations_clean (
    id BIGSERIAL PRIMARY KEY,
    series_id BIGINT NOT NULL REFERENCES macro_series(id),
    observation_date DATE NOT NULL,
    value DOUBLE PRECISION NOT NULL,
    vintage_date DATE,
    is_latest BOOLEAN NOT NULL DEFAULT true,
    frequency TEXT NOT NULL
);

CREATE UNIQUE INDEX idx_macro_obs_clean_series_date_vintage ON macro_observations_clean (series_id, observation_date, vintage_date)
    WHERE vintage_date IS NOT NULL;
CREATE UNIQUE INDEX idx_macro_obs_clean_series_date_latest ON macro_observations_clean (series_id, observation_date)
    WHERE vintage_date IS NULL;
CREATE INDEX idx_macro_obs_clean_series_date ON macro_observations_clean (series_id, observation_date);

-- +goose Down
DROP TABLE IF EXISTS macro_observations_clean;
DROP TABLE IF EXISTS macro_observations_raw;
