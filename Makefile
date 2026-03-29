.PHONY: help build run migrate-up migrate-status seed ingest-all ingest-macro ingest-cot rebuild web

# Load .env if it exists
-include .env
export

API_DIR := apps/api
WEB_DIR := apps/web
API_URL ?= http://localhost:8080
TOKEN   ?= $(INTERNAL_AUTH_TOKEN)

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

# ── Build ──────────────────────────────────────────────────────────────────────

build: ## Build the API server binary
	cd $(API_DIR) && go build ./cmd/server/

# ── Run ───────────────────────────────────────────────────────────────────────

run: ## Run the API server (requires DATABASE_URL env var)
	cd $(API_DIR) && go run ./cmd/server/

web: ## Run the Next.js dev server
	cd $(WEB_DIR) && npm run dev

# ── Database ──────────────────────────────────────────────────────────────────

migrate-up: ## Run all pending migrations
	cd $(API_DIR) && go run ./cmd/migrate/ up

migrate-status: ## Show migration status
	cd $(API_DIR) && go run ./cmd/migrate/ status

migrate-reset: ## Roll back ALL migrations (destructive)
	cd $(API_DIR) && go run ./cmd/migrate/ reset

seed: ## Apply all seed files to the database
	@echo "Seeding commodities..."
	psql "$(DATABASE_URL)" -f db/seeds/001_commodities.sql
	@echo "Seeding macro series..."
	psql "$(DATABASE_URL)" -f db/seeds/002_macro_series.sql
	@echo "Seeding model config..."
	psql "$(DATABASE_URL)" -f db/seeds/003_model_config.sql
	@echo "Done."

setup: migrate-up seed ## Run migrations + seed (first-time setup)

# ── Jobs (calls running API server) ───────────────────────────────────────────

ingest-macro: ## Trigger FRED macro ingestion (5yr lookback)
	curl -s -X POST $(API_URL)/api/v1/internal/jobs/ingest-macro \
	  -H "Authorization: Bearer $(TOKEN)" \
	  -H "Content-Type: application/json" \
	  -d '{"lookback_years": 5}' | jq .

ingest-macro-full: ## Trigger FRED macro ingestion (10yr lookback, first run)
	curl -s -X POST $(API_URL)/api/v1/internal/jobs/ingest-macro \
	  -H "Authorization: Bearer $(TOKEN)" \
	  -H "Content-Type: application/json" \
	  -d '{"lookback_years": 10}' | jq .

ingest-cot: ## Trigger CFTC COT ingestion (5yr lookback)
	curl -s -X POST $(API_URL)/api/v1/internal/jobs/ingest-cot \
	  -H "Authorization: Bearer $(TOKEN)" \
	  -H "Content-Type: application/json" \
	  -d '{"lookback_years": 5}' | jq .

ingest-cot-full: ## Trigger CFTC COT ingestion (10yr lookback, first run)
	curl -s -X POST $(API_URL)/api/v1/internal/jobs/ingest-cot \
	  -H "Authorization: Bearer $(TOKEN)" \
	  -H "Content-Type: application/json" \
	  -d '{"lookback_years": 10}' | jq .

rebuild: ## Trigger snapshot + signal + alert rebuild
	curl -s -X POST $(API_URL)/api/v1/internal/jobs/rebuild-snapshots \
	  -H "Authorization: Bearer $(TOKEN)" \
	  -H "Content-Type: application/json" \
	  -d '{}' | jq .

ingest-all: ingest-macro ingest-cot rebuild ## Run all three jobs in sequence

# ── First-time full bootstrap ─────────────────────────────────────────────────

bootstrap: ## Full first-time data load: 10yr macro + 10yr COT + rebuild
	@echo "==> Ingesting 10yr FRED macro data..."
	curl -s -X POST $(API_URL)/api/v1/internal/jobs/ingest-macro \
	  -H "Authorization: Bearer $(TOKEN)" \
	  -H "Content-Type: application/json" \
	  -d '{"lookback_years": 10}' | jq .
	@echo "==> Ingesting 10yr CFTC COT data..."
	curl -s -X POST $(API_URL)/api/v1/internal/jobs/ingest-cot \
	  -H "Authorization: Bearer $(TOKEN)" \
	  -H "Content-Type: application/json" \
	  -d '{"lookback_years": 10}' | jq .
	@echo "==> Rebuilding snapshots + signals + alerts..."
	curl -s -X POST $(API_URL)/api/v1/internal/jobs/rebuild-snapshots \
	  -H "Authorization: Bearer $(TOKEN)" \
	  -H "Content-Type: application/json" \
	  -d '{}' | jq .
	@echo "==> Bootstrap complete."

# ── Tests ─────────────────────────────────────────────────────────────────────

test: ## Run all Go tests
	cd $(API_DIR) && go test ./...

test-v: ## Run all Go tests (verbose)
	cd $(API_DIR) && go test -v ./...
