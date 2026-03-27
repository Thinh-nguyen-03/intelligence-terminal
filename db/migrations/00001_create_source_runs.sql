-- +goose Up
CREATE TYPE source_enum AS ENUM ('fred', 'alfred', 'cftc', 'analytics');
CREATE TYPE job_status_enum AS ENUM ('pending', 'running', 'success', 'failed');

CREATE TABLE source_runs (
    id BIGSERIAL PRIMARY KEY,
    source source_enum NOT NULL,
    job_name TEXT NOT NULL,
    status job_status_enum NOT NULL DEFAULT 'pending',
    started_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    finished_at TIMESTAMPTZ,
    records_processed INTEGER DEFAULT 0,
    error_message TEXT,
    checksum TEXT
);

CREATE INDEX idx_source_runs_source_status ON source_runs (source, status);
CREATE INDEX idx_source_runs_started_at ON source_runs (started_at DESC);

-- +goose Down
DROP TABLE IF EXISTS source_runs;
DROP TYPE IF EXISTS job_status_enum;
DROP TYPE IF EXISTS source_enum;
