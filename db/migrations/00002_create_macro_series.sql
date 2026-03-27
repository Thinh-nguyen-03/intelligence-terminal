-- +goose Up
CREATE TABLE macro_series (
    id BIGSERIAL PRIMARY KEY,
    source TEXT NOT NULL CHECK (source IN ('fred', 'alfred')),
    source_series_id TEXT NOT NULL,
    slug TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    frequency TEXT NOT NULL,
    units TEXT NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX idx_macro_series_source_id ON macro_series (source, source_series_id);

-- +goose Down
DROP TABLE IF EXISTS macro_series;
