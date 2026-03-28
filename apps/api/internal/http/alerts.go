package handler

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/0510t/intelligence-terminal/apps/api/internal/domain"
	"github.com/0510t/intelligence-terminal/apps/api/internal/storage"
)

// AlertHandlers holds dependencies for alert endpoints.
type AlertHandlers struct {
	alertRepo *storage.AlertRepo
	cotRepo   *storage.COTRepo
}

func NewAlertHandlers(alertRepo *storage.AlertRepo, cotRepo *storage.COTRepo) *AlertHandlers {
	return &AlertHandlers{alertRepo: alertRepo, cotRepo: cotRepo}
}

// List returns active alerts with optional filters.
// GET /api/v1/alerts?severity=critical&commodity=copper
func (h *AlertHandlers) List(w http.ResponseWriter, r *http.Request) {
	var severity *string
	var commoditySlug *string

	if s := r.URL.Query().Get("severity"); s != "" {
		severity = &s
	}
	if c := r.URL.Query().Get("commodity"); c != "" {
		commoditySlug = &c
	}

	alerts, err := h.alertRepo.ListActiveAlerts(r.Context(), severity, commoditySlug)
	if err != nil {
		WriteError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to query alerts")
		return
	}

	// Build commodity ID -> slug/name lookup
	commodityMap := h.buildCommodityMap(r)

	dtos := make([]AlertDTO, len(alerts))
	for i, a := range alerts {
		slug, name := commodityMap[a.CommodityID].slug, commodityMap[a.CommodityID].name
		dtos[i] = alertToDTO(a, slug, name)
	}

	w.Header().Set("Cache-Control", "public, max-age=60")
	WriteJSON(w, http.StatusOK, AlertListResponse{Alerts: dtos, Count: len(dtos)})
}

// GetByID returns a single alert with full explanation.
// GET /api/v1/alerts/{id}
func (h *AlertHandlers) GetByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		WriteError(w, r, http.StatusBadRequest, "VALIDATION_ERROR", "invalid alert ID")
		return
	}

	alert, err := h.alertRepo.GetAlertByID(r.Context(), id)
	if err != nil {
		WriteError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to query alert")
		return
	}
	if alert == nil {
		WriteError(w, r, http.StatusNotFound, "RESOURCE_NOT_FOUND", "alert not found")
		return
	}

	commodityMap := h.buildCommodityMap(r)
	slug, name := commodityMap[alert.CommodityID].slug, commodityMap[alert.CommodityID].name

	resp := AlertDetailResponse{
		AlertDTO:        alertToDTO(*alert, slug, name),
		ExplanationJSON: alert.ExplanationJSON,
	}

	WriteJSON(w, http.StatusOK, resp)
}

type commodityInfo struct {
	slug string
	name string
}

func (h *AlertHandlers) buildCommodityMap(r *http.Request) map[int64]commodityInfo {
	m := make(map[int64]commodityInfo)
	commodities, err := h.cotRepo.ListActiveCommodities(r.Context())
	if err != nil {
		return m
	}
	for _, c := range commodities {
		m[c.ID] = commodityInfo{slug: c.Slug, name: c.Name}
	}
	return m
}

func alertToDTO(a domain.Alert, slug, name string) AlertDTO {
	return AlertDTO{
		ID:               a.ID,
		CommoditySlug:    slug,
		CommodityName:    name,
		AsOfDate:         FormatDate(a.AsOfDate),
		Severity:         string(a.Severity),
		AlertType:        a.AlertType,
		Headline:         a.Headline,
		Summary:          a.Summary,
		RegimeLabel:      a.RegimeLabel,
		RegimeConfidence: a.RegimeConfidence,
		FinalAlertScore:  a.FinalAlertScore,
		IsActive:         a.IsActive,
		CreatedAt:        FormatDateTime(a.CreatedAt),
	}
}
