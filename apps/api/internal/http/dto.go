package handler

import (
	"encoding/json"
	"time"
)

// --- Regime DTOs ---

type RegimeCurrentResponse struct {
	AsOfDate        string          `json:"as_of_date"`
	RegimeLabel     string          `json:"regime_label"`
	PriorLabel      *string         `json:"prior_label,omitempty"`
	Confidence      float64         `json:"confidence"`
	IsTransitioning bool            `json:"is_transitioning"`
	TransitionDetail *string        `json:"transition_detail,omitempty"`
	FactorScores    FactorScoresDTO `json:"factor_scores"`
	ModelVersion    string          `json:"model_version"`
}

type FactorScoresDTO struct {
	Growth    float64 `json:"growth"`
	Inflation float64 `json:"inflation"`
	Labor     float64 `json:"labor"`
	Stress    float64 `json:"stress"`
}

type RegimeHistoryResponse struct {
	Snapshots []RegimeSnapshotDTO `json:"snapshots"`
	Count     int                 `json:"count"`
}

type RegimeSnapshotDTO struct {
	AsOfDate        string          `json:"as_of_date"`
	RegimeLabel     string          `json:"regime_label"`
	Confidence      float64         `json:"confidence"`
	IsTransitioning bool            `json:"is_transitioning"`
	FactorScores    FactorScoresDTO `json:"factor_scores"`
}

// --- Commodity DTOs ---

type CommodityListResponse struct {
	Commodities []CommodityDTO `json:"commodities"`
	Count       int            `json:"count"`
}

type CommodityDTO struct {
	Slug      string `json:"slug"`
	Name      string `json:"name"`
	GroupName string `json:"group_name"`
	Active    bool   `json:"active"`
}

type CommodityDetailResponse struct {
	Commodity    CommodityDTO       `json:"commodity"`
	LatestSignal *SignalSnapshotDTO `json:"latest_signal,omitempty"`
	ActiveAlerts []AlertDTO         `json:"active_alerts"`
}

type SignalSnapshotDTO struct {
	AsOfDate           string   `json:"as_of_date"`
	NetManagedMoney    int64    `json:"net_managed_money"`
	NetMMPctOI         float64  `json:"net_mm_pct_oi"`
	PositionZScore26W  *float64 `json:"position_zscore_26w,omitempty"`
	PositionZScore52W  *float64 `json:"position_zscore_52w,omitempty"`
	PositionPercentile *float64 `json:"position_percentile_52w,omitempty"`
	WeeklyChangeNetMM  *int64   `json:"weekly_change_net_mm,omitempty"`
	CrowdingScore      float64  `json:"crowding_score"`
	SqueezeRiskScore   float64  `json:"squeeze_risk_score"`
	ReversalRiskScore  float64  `json:"reversal_risk_score"`
	TrendSupportScore  float64  `json:"trend_support_score"`
	ModelVersion       string   `json:"model_version"`
}

type CommodityHistoryResponse struct {
	Slug      string              `json:"slug"`
	Snapshots []SignalSnapshotDTO `json:"snapshots"`
	Count     int                 `json:"count"`
}

type RankingEntry struct {
	Slug  string  `json:"slug"`
	Name  string  `json:"name"`
	Score float64 `json:"score"`
}

type CommodityRankingsResponse struct {
	Crowding     []RankingEntry `json:"crowding"`
	SqueezeRisk  []RankingEntry `json:"squeeze_risk"`
	ReversalRisk []RankingEntry `json:"reversal_risk"`
	TrendSupport []RankingEntry `json:"trend_support"`
}

// --- Alert DTOs ---

type AlertListResponse struct {
	Alerts []AlertDTO `json:"alerts"`
	Count  int        `json:"count"`
}

type AlertDTO struct {
	ID               int64           `json:"id"`
	CommoditySlug    string          `json:"commodity_slug,omitempty"`
	CommodityName    string          `json:"commodity_name,omitempty"`
	AsOfDate         string          `json:"as_of_date"`
	Severity         string          `json:"severity"`
	AlertType        string          `json:"alert_type"`
	Headline         string          `json:"headline"`
	Summary          string          `json:"summary"`
	RegimeLabel      string          `json:"regime_label"`
	RegimeConfidence float64         `json:"regime_confidence"`
	FinalAlertScore  float64         `json:"final_alert_score"`
	IsActive         bool            `json:"is_active"`
	CreatedAt        string          `json:"created_at"`
}

type AlertDetailResponse struct {
	AlertDTO
	ExplanationJSON json.RawMessage `json:"explanation"`
}

// --- Freshness DTO ---

type FreshnessResponse struct {
	Sources []SourceFreshness `json:"sources"`
}

type SourceFreshness struct {
	Source        string  `json:"source"`
	LastIngested  *string `json:"last_ingested,omitempty"`
	PositionsAsOf *string `json:"positions_as_of,omitempty"`
	Status        string  `json:"status"`
}

// --- helpers ---

func FormatDate(t time.Time) string {
	return t.Format("2006-01-02")
}

func FormatDateTime(t time.Time) string {
	return t.Format(time.RFC3339)
}
