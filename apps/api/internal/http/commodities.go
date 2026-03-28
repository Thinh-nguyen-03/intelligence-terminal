package handler

import (
	"net/http"
	"sort"

	"github.com/go-chi/chi/v5"

	"github.com/0510t/intelligence-terminal/apps/api/internal/domain"
	"github.com/0510t/intelligence-terminal/apps/api/internal/storage"
)

// CommodityHandlers holds dependencies for commodity endpoints.
type CommodityHandlers struct {
	cotRepo    *storage.COTRepo
	signalRepo *storage.SignalRepo
	alertRepo  *storage.AlertRepo
	configRepo *storage.ConfigRepo
}

func NewCommodityHandlers(cotRepo *storage.COTRepo, signalRepo *storage.SignalRepo, alertRepo *storage.AlertRepo, configRepo *storage.ConfigRepo) *CommodityHandlers {
	return &CommodityHandlers{
		cotRepo:    cotRepo,
		signalRepo: signalRepo,
		alertRepo:  alertRepo,
		configRepo: configRepo,
	}
}

// List returns all active commodities.
// GET /api/v1/commodities
func (h *CommodityHandlers) List(w http.ResponseWriter, r *http.Request) {
	commodities, err := h.cotRepo.ListActiveCommodities(r.Context())
	if err != nil {
		WriteError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to query commodities")
		return
	}

	dtos := make([]CommodityDTO, len(commodities))
	for i, c := range commodities {
		dtos[i] = commodityToDTO(c)
	}

	w.Header().Set("Cache-Control", "public, max-age=3600")
	WriteJSON(w, http.StatusOK, CommodityListResponse{Commodities: dtos, Count: len(dtos)})
}

// GetDetail returns current detail for one commodity.
// GET /api/v1/commodities/{slug}
func (h *CommodityHandlers) GetDetail(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")

	// Find commodity by slug
	commodity, err := h.findCommodityBySlug(r, slug)
	if err != nil {
		WriteError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to query commodities")
		return
	}
	if commodity == nil {
		WriteError(w, r, http.StatusNotFound, "RESOURCE_NOT_FOUND", "commodity not found: "+slug)
		return
	}

	params, err := h.configRepo.LoadModelParams(r.Context())
	if err != nil {
		WriteError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to load config")
		return
	}

	resp := CommodityDetailResponse{
		Commodity:    commodityToDTO(*commodity),
		ActiveAlerts: []AlertDTO{},
	}

	// Latest signal
	sig, err := h.signalRepo.GetLatestSignal(r.Context(), commodity.ID, params.ModelVersion)
	if err == nil && sig != nil {
		dto := signalToDTO(*sig)
		resp.LatestSignal = &dto
	}

	// Active alerts for this commodity
	alerts, err := h.alertRepo.ListActiveAlerts(r.Context(), nil, &slug)
	if err == nil {
		for _, a := range alerts {
			resp.ActiveAlerts = append(resp.ActiveAlerts, alertToDTO(a, commodity.Slug, commodity.Name))
		}
	}

	w.Header().Set("Cache-Control", "public, max-age=300")
	WriteJSON(w, http.StatusOK, resp)
}

// GetHistory returns historical signal snapshots for a commodity.
// GET /api/v1/commodities/{slug}/history?from=YYYY-MM-DD&to=YYYY-MM-DD
func (h *CommodityHandlers) GetHistory(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")

	commodity, err := h.findCommodityBySlug(r, slug)
	if err != nil {
		WriteError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to query commodities")
		return
	}
	if commodity == nil {
		WriteError(w, r, http.StatusNotFound, "RESOURCE_NOT_FOUND", "commodity not found: "+slug)
		return
	}

	from, to, ok := parseDateRange(w, r)
	if !ok {
		return
	}

	params, err := h.configRepo.LoadModelParams(r.Context())
	if err != nil {
		WriteError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to load config")
		return
	}

	signals, err := h.signalRepo.ListSignals(r.Context(), commodity.ID, from, to, params.ModelVersion)
	if err != nil {
		WriteError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to query signals")
		return
	}

	dtos := make([]SignalSnapshotDTO, len(signals))
	for i, s := range signals {
		dtos[i] = signalToDTO(s)
	}

	w.Header().Set("Cache-Control", "public, max-age=3600")
	WriteJSON(w, http.StatusOK, CommodityHistoryResponse{Slug: slug, Snapshots: dtos, Count: len(dtos)})
}

// GetRankings returns ranked commodity lists across all score dimensions.
// GET /api/v1/commodities/rankings
func (h *CommodityHandlers) GetRankings(w http.ResponseWriter, r *http.Request) {
	commodities, err := h.cotRepo.ListActiveCommodities(r.Context())
	if err != nil {
		WriteError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to query commodities")
		return
	}

	params, err := h.configRepo.LoadModelParams(r.Context())
	if err != nil {
		WriteError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to load config")
		return
	}

	type rankEntry struct {
		slug, name string
		score      float64
	}

	var crowding, squeeze, reversal, trend []rankEntry
	for _, c := range commodities {
		sig, err := h.signalRepo.GetLatestSignal(r.Context(), c.ID, params.ModelVersion)
		if err != nil || sig == nil {
			continue
		}
		crowding = append(crowding, rankEntry{c.Slug, c.Name, sig.CrowdingScore})
		squeeze = append(squeeze, rankEntry{c.Slug, c.Name, sig.SqueezeRiskScore})
		reversal = append(reversal, rankEntry{c.Slug, c.Name, sig.ReversalRiskScore})
		trend = append(trend, rankEntry{c.Slug, c.Name, sig.TrendSupportScore})
	}

	sortDesc := func(entries []rankEntry) []RankingEntry {
		sort.Slice(entries, func(i, j int) bool { return entries[i].score > entries[j].score })
		result := make([]RankingEntry, len(entries))
		for i, e := range entries {
			result[i] = RankingEntry{Slug: e.slug, Name: e.name, Score: e.score}
		}
		return result
	}

	resp := CommodityRankingsResponse{
		Crowding:     sortDesc(crowding),
		SqueezeRisk:  sortDesc(squeeze),
		ReversalRisk: sortDesc(reversal),
		TrendSupport: sortDesc(trend),
	}

	w.Header().Set("Cache-Control", "public, max-age=300")
	WriteJSON(w, http.StatusOK, resp)
}

func (h *CommodityHandlers) findCommodityBySlug(r *http.Request, slug string) (*domain.Commodity, error) {
	commodities, err := h.cotRepo.ListActiveCommodities(r.Context())
	if err != nil {
		return nil, err
	}
	for _, c := range commodities {
		if c.Slug == slug {
			return &c, nil
		}
	}
	return nil, nil
}

func commodityToDTO(c domain.Commodity) CommodityDTO {
	return CommodityDTO{
		Slug:      c.Slug,
		Name:      c.Name,
		GroupName: c.GroupName,
		Active:    c.Active,
	}
}

func signalToDTO(s domain.CommoditySignalSnapshot) SignalSnapshotDTO {
	return SignalSnapshotDTO{
		AsOfDate:           FormatDate(s.AsOfDate),
		NetManagedMoney:    s.NetManagedMoney,
		NetMMPctOI:         s.NetMMPctOI,
		PositionZScore26W:  s.PositionZScore26W,
		PositionZScore52W:  s.PositionZScore52W,
		PositionPercentile: s.PositionPercentile,
		WeeklyChangeNetMM:  s.WeeklyChangeNetMM,
		CrowdingScore:      s.CrowdingScore,
		SqueezeRiskScore:   s.SqueezeRiskScore,
		ReversalRiskScore:  s.ReversalRiskScore,
		TrendSupportScore:  s.TrendSupportScore,
		ModelVersion:       s.ModelVersion,
	}
}

