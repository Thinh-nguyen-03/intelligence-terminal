package handler

import (
	"net/http"
	"time"

	"github.com/0510t/intelligence-terminal/apps/api/internal/storage"
)

// RegimeHandlers holds dependencies for regime endpoints.
type RegimeHandlers struct {
	snapshotRepo *storage.SnapshotRepo
	configRepo   *storage.ConfigRepo
}

func NewRegimeHandlers(snapshotRepo *storage.SnapshotRepo, configRepo *storage.ConfigRepo) *RegimeHandlers {
	return &RegimeHandlers{snapshotRepo: snapshotRepo, configRepo: configRepo}
}

// GetCurrent returns the latest regime snapshot.
// GET /api/v1/regime/current
func (h *RegimeHandlers) GetCurrent(w http.ResponseWriter, r *http.Request) {
	params, err := h.configRepo.LoadModelParams(r.Context())
	if err != nil {
		WriteError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to load config")
		return
	}

	snap, err := h.snapshotRepo.GetLatestSnapshot(r.Context(), params.ModelVersion)
	if err != nil {
		WriteError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to query snapshot")
		return
	}
	if snap == nil {
		WriteError(w, r, http.StatusServiceUnavailable, "DATA_STALE", "no regime data available yet")
		return
	}

	resp := RegimeCurrentResponse{
		AsOfDate:         FormatDate(snap.AsOfDate),
		RegimeLabel:      snap.RegimeLabel,
		PriorLabel:       snap.RegimePriorLabel,
		Confidence:       snap.Confidence,
		IsTransitioning:  snap.IsTransitioning,
		TransitionDetail: snap.TransitionDetail,
		FactorScores: FactorScoresDTO{
			Growth:    snap.GrowthScore,
			Inflation: snap.InflationScore,
			Labor:     snap.LaborScore,
			Stress:    snap.StressScore,
		},
		ModelVersion: snap.ModelVersion,
	}

	w.Header().Set("Cache-Control", "public, max-age=300")
	WriteJSON(w, http.StatusOK, resp)
}

// GetHistory returns regime snapshots within a date range.
// GET /api/v1/regime/history?from=YYYY-MM-DD&to=YYYY-MM-DD
func (h *RegimeHandlers) GetHistory(w http.ResponseWriter, r *http.Request) {
	from, to, ok := parseDateRange(w, r)
	if !ok {
		return
	}

	params, err := h.configRepo.LoadModelParams(r.Context())
	if err != nil {
		WriteError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to load config")
		return
	}

	snaps, err := h.snapshotRepo.ListSnapshots(r.Context(), from, to, params.ModelVersion)
	if err != nil {
		WriteError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to query snapshots")
		return
	}

	dtos := make([]RegimeSnapshotDTO, len(snaps))
	for i, s := range snaps {
		dtos[i] = RegimeSnapshotDTO{
			AsOfDate:        FormatDate(s.AsOfDate),
			RegimeLabel:     s.RegimeLabel,
			Confidence:      s.Confidence,
			IsTransitioning: s.IsTransitioning,
			FactorScores: FactorScoresDTO{
				Growth:    s.GrowthScore,
				Inflation: s.InflationScore,
				Labor:     s.LaborScore,
				Stress:    s.StressScore,
			},
		}
	}

	w.Header().Set("Cache-Control", "public, max-age=3600")
	WriteJSON(w, http.StatusOK, RegimeHistoryResponse{Snapshots: dtos, Count: len(dtos)})
}

func parseDateRange(w http.ResponseWriter, r *http.Request) (time.Time, time.Time, bool) {
	fromStr := r.URL.Query().Get("from")
	toStr := r.URL.Query().Get("to")

	if fromStr == "" || toStr == "" {
		WriteError(w, r, http.StatusBadRequest, "VALIDATION_ERROR", "from and to query params are required (YYYY-MM-DD)")
		return time.Time{}, time.Time{}, false
	}

	from, err := time.Parse("2006-01-02", fromStr)
	if err != nil {
		WriteError(w, r, http.StatusBadRequest, "VALIDATION_ERROR", "invalid from date format, use YYYY-MM-DD")
		return time.Time{}, time.Time{}, false
	}

	to, err := time.Parse("2006-01-02", toStr)
	if err != nil {
		WriteError(w, r, http.StatusBadRequest, "VALIDATION_ERROR", "invalid to date format, use YYYY-MM-DD")
		return time.Time{}, time.Time{}, false
	}

	return from, to, true
}
