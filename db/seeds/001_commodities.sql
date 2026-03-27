-- Seed: v1 commodities (5 markets)
-- CFTC commodity codes are from the Disaggregated Futures-Only report

INSERT INTO commodities (slug, name, cftc_commodity_code, group_name, active) VALUES
    ('gold',        'Gold',           '088691', 'metals', true),
    ('silver',      'Silver',         '084691', 'metals', true),
    ('wti-crude',   'WTI Crude Oil',  '067651', 'energy', true),
    ('natural-gas', 'Natural Gas',    '023651', 'energy', true),
    ('copper',      'Copper',         '085692', 'metals', true)
ON CONFLICT (slug) DO NOTHING;
