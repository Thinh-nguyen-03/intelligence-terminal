# Local Setup Runbook

## Prerequisites
- Go 1.22+
- Node.js 20+ (for frontend, later)
- A Neon Postgres database (free tier)
- A FRED API key (free)

## 1. Environment Variables

Create a `.env` file in the repo root (gitignored):

```
DATABASE_URL=postgresql://user:pass@ep-xxx.region.aws.neon.tech/dbname?sslmode=require
FRED_API_KEY=your_fred_api_key_here
PORT=8080
INTERNAL_AUTH_TOKEN=some-secret-token-for-internal-endpoints
```

## 2. Run Migrations

```bash
cd apps/api
go run ./cmd/migrate/ up
```

## 3. Seed Reference Data

```bash
psql $DATABASE_URL -f db/seeds/001_commodities.sql
psql $DATABASE_URL -f db/seeds/002_macro_series.sql
psql $DATABASE_URL -f db/seeds/003_model_config.sql
```

## 4. Start the API

```bash
cd apps/api
DATABASE_URL=... FRED_API_KEY=... go run ./cmd/server/
```

## 5. Verify

```bash
curl http://localhost:8080/api/v1/health
# Should return: {"status":"healthy","db":"connected"}
```

## 6. Run Backfill (first time only)

```bash
cd apps/api
go run ./cmd/backfill/ --source=fred --years=5
go run ./cmd/backfill/ --source=cftc --years=5
```
