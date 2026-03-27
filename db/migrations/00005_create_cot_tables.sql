-- +goose Up
CREATE TABLE cot_reports_raw (
    id BIGSERIAL PRIMARY KEY,
    report_date DATE NOT NULL,
    report_type TEXT NOT NULL DEFAULT 'disaggregated',
    file_checksum TEXT NOT NULL,
    payload_text TEXT,
    ingested_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX idx_cot_raw_date_type ON cot_reports_raw (report_date, report_type);
CREATE UNIQUE INDEX idx_cot_raw_checksum ON cot_reports_raw (file_checksum);

CREATE TABLE cot_positions_clean (
    id BIGSERIAL PRIMARY KEY,
    commodity_id BIGINT NOT NULL REFERENCES commodities(id),
    report_date DATE NOT NULL,
    open_interest BIGINT NOT NULL,
    producer_merchant_long BIGINT NOT NULL,
    producer_merchant_short BIGINT NOT NULL,
    swap_dealer_long BIGINT NOT NULL,
    swap_dealer_short BIGINT NOT NULL,
    managed_money_long BIGINT NOT NULL,
    managed_money_short BIGINT NOT NULL,
    other_reportable_long BIGINT NOT NULL,
    other_reportable_short BIGINT NOT NULL,
    nonreportable_long BIGINT NOT NULL,
    nonreportable_short BIGINT NOT NULL
);

CREATE UNIQUE INDEX idx_cot_clean_commodity_date ON cot_positions_clean (commodity_id, report_date);
CREATE INDEX idx_cot_clean_report_date ON cot_positions_clean (report_date);

-- +goose Down
DROP TABLE IF EXISTS cot_positions_clean;
DROP TABLE IF EXISTS cot_reports_raw;
