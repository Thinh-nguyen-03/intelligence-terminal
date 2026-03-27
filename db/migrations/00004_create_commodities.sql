-- +goose Up
CREATE TABLE commodities (
    id BIGSERIAL PRIMARY KEY,
    slug TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    cftc_commodity_code TEXT NOT NULL,
    group_name TEXT NOT NULL CHECK (group_name IN ('metals', 'energy')),
    active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- +goose Down
DROP TABLE IF EXISTS commodities;
