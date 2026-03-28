# Intelligence Terminal - Development Guide

## Project
Regime + Positioning Intelligence Terminal. Full spec: `regime_positioning_intelligence_terminal_v2.md`

## Tech Stack
- **API:** Go + chi + pgx (in `apps/api/`)
- **Frontend:** Next.js + TypeScript + Tailwind (in `apps/web/`, not started yet)
- **Database:** Neon Postgres, migrations via goose (in `db/migrations/`)
- **Scheduling:** GitHub Actions (in `.github/workflows/`)

## Commands
```bash
# Build API
cd apps/api && go build ./cmd/server/

# Run API (requires DATABASE_URL env var)
cd apps/api && DATABASE_URL=... go run ./cmd/server/

# Run migrations (once migration runner is built)
cd apps/api && go run ./cmd/migrate/ up

# Run tests
cd apps/api && go test ./...
```

## Architecture
- Raw data is never overwritten (append-only ingestion)
- Normalized data stored separately from raw payloads
- All analytics are precomputed asynchronously, served from snapshot tables
- Every score is explainable via explanation_json payloads
- Model weights/thresholds live in `model_config` table, not hardcoded

## Conventions
- Go packages follow `internal/` layout: domain, storage, service, analytics, jobs, http, config, auth, cache
- Domain types are plain structs with no ORM tags
- Storage layer uses raw pgx queries, no ORM
- API responses use DTOs defined in `internal/http/`, not domain types directly
- Migrations are sequential numbered SQL files with goose annotations
- Seed files in `db/seeds/` are idempotent (ON CONFLICT DO NOTHING / DO UPDATE)

## Data Sources
- FRED API (macro series) — requires FRED_API_KEY env var
- CFTC Disaggregated COT report (commodity positioning) — public CSV download
- ALFRED (point-in-time vintages) — same API key as FRED
