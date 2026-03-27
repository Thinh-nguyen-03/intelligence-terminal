-- Seed: v1 macro series (10 FRED series)

INSERT INTO macro_series (source, source_series_id, slug, name, frequency, units, enabled) VALUES
    ('fred', 'CPIAUCSL',  'cpi',                'CPI All Urban Consumers',         'monthly', 'index',   true),
    ('fred', 'CPILFESL',  'core-cpi',           'Core CPI (Less Food and Energy)', 'monthly', 'index',   true),
    ('fred', 'UNRATE',    'unemployment',        'Unemployment Rate',               'monthly', 'percent', true),
    ('fred', 'PAYEMS',    'nonfarm-payrolls',    'Total Nonfarm Payrolls',          'monthly', 'thousands', true),
    ('fred', 'INDPRO',    'industrial-prod',     'Industrial Production Index',     'monthly', 'index',   true),
    ('fred', 'RSAFS',     'retail-sales',        'Retail Sales',                    'monthly', 'millions', true),
    ('fred', 'T10Y2Y',    'yield-curve-10y2y',   '10Y-2Y Treasury Spread',         'daily',   'percent', true),
    ('fred', 'FEDFUNDS',  'fed-funds',           'Federal Funds Effective Rate',    'monthly', 'percent', true),
    ('fred', 'DCOILWTICO','wti-spot',            'WTI Crude Oil Spot Price',        'daily',   'dollars', true),
    ('fred', 'DTWEXBGS',  'dollar-index',        'Trade Weighted Dollar Index',     'daily',   'index',   true)
ON CONFLICT (slug) DO NOTHING;
