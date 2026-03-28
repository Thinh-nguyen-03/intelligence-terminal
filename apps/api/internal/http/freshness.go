package handler

import (
	"context"
	"net/http"

	"github.com/0510t/intelligence-terminal/apps/api/internal/domain"
	"github.com/0510t/intelligence-terminal/apps/api/internal/storage"
)

// FreshnessHandlers holds dependencies for freshness endpoints.
type FreshnessHandlers struct {
	sourceRunRepo *storage.SourceRunRepo
	cotRepo       *storage.COTRepo
}

func NewFreshnessHandlers(sourceRunRepo *storage.SourceRunRepo, cotRepo *storage.COTRepo) *FreshnessHandlers {
	return &FreshnessHandlers{sourceRunRepo: sourceRunRepo, cotRepo: cotRepo}
}

// Get returns dataset freshness info.
// GET /api/v1/freshness
func (h *FreshnessHandlers) Get(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	sources := []SourceFreshness{
		h.sourceFreshness(ctx, domain.SourceFRED, "ingest-macro", "FRED Macro Series"),
		h.sourceFreshness(ctx, domain.SourceCFTC, "ingest-cot", "CFTC COT Report"),
		h.sourceFreshness(ctx, domain.SourceAnalytics, "rebuild-snapshots", "Analytics Snapshots"),
	}

	// Add COT positions-as-of date
	latestReport, err := h.cotRepo.GetLatestReportDate(ctx)
	if err == nil && latestReport != nil {
		d := FormatDate(*latestReport)
		for i := range sources {
			if sources[i].Source == "CFTC COT Report" {
				sources[i].PositionsAsOf = &d
			}
		}
	}

	WriteJSON(w, http.StatusOK, FreshnessResponse{Sources: sources})
}

func (h *FreshnessHandlers) sourceFreshness(ctx context.Context, source domain.Source, jobName, label string) SourceFreshness {
	sf := SourceFreshness{
		Source: label,
		Status: "no_data",
	}

	run, err := h.sourceRunRepo.GetLatestSuccessful(ctx, source, jobName)
	if err != nil || run == nil {
		return sf
	}

	if run.FinishedAt != nil {
		d := FormatDateTime(*run.FinishedAt)
		sf.LastIngested = &d
		sf.Status = "ok"
	}

	return sf
}
