# Architecture Decision Log

Decisions made during implementation that aren't in the spec. Record the decision, why, and what was considered.

---

## ADR-001: Disaggregated COT over Legacy
**Date:** 2026-03-27
**Decision:** Use CFTC Disaggregated report as primary COT source.
**Why:** Managed money positions are a cleaner speculative signal than Legacy's non-commercial bucket, which blends swap dealers with speculators. Disaggregated has history back to 2006 (20 years).
**Alternative:** Legacy report (longer history, simpler schema). Rejected because the signal quality improvement outweighs the extra fields.

## ADR-002: v1 commodity set
**Date:** 2026-03-27
**Decision:** Gold, Silver, WTI Crude, Natural Gas, Copper.
**Why:** These 5 are the most macro-sensitive commodities. Copper is included specifically to showcase regime-conditioned analysis. Agricultural markets deferred — they need USDA context to get real value from the regime layer.

## ADR-003: Overlapping score ranges for regime model
**Date:** 2026-03-27
**Decision:** Use overlapping factor score ranges with transition states instead of hard thresholds.
**Why:** Hard cutoffs create cliff effects where trivial score changes flip the regime label. Overlapping ranges + confidence scoring makes transitions visible and reduces false precision.

## ADR-004: goose for migrations
**Date:** 2026-03-27
**Decision:** Use goose (pressly/goose) for database migrations.
**Why:** Go-native, supports plain SQL files, widely adopted in Go+Postgres projects. golang-migrate was considered but goose's SQL file support is cleaner.

## ADR-005: Domain types without ORM
**Date:** 2026-03-27
**Decision:** Plain Go structs in domain package, raw pgx queries in storage.
**Why:** No ORM overhead, full control over SQL, better performance with pgx. The schema is stable and well-defined — an ORM adds complexity without value here.
