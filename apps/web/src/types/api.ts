// --- Regime ---

export interface RegimeCurrentResponse {
  as_of_date: string;
  regime_label: string;
  prior_label?: string;
  confidence: number;
  is_transitioning: boolean;
  transition_detail?: string;
  factor_scores: FactorScores;
  model_version: string;
}

export interface FactorScores {
  growth: number;
  inflation: number;
  labor: number;
  stress: number;
}

export interface RegimeHistoryResponse {
  snapshots: RegimeSnapshot[];
  count: number;
}

export interface RegimeSnapshot {
  as_of_date: string;
  regime_label: string;
  confidence: number;
  is_transitioning: boolean;
  factor_scores: FactorScores;
}

// --- Commodities ---

export interface CommodityListResponse {
  commodities: Commodity[];
  count: number;
}

export interface Commodity {
  slug: string;
  name: string;
  group_name: string;
  active: boolean;
}

export interface CommodityDetailResponse {
  commodity: Commodity;
  latest_signal?: SignalSnapshot;
  active_alerts: Alert[];
}

export interface SignalSnapshot {
  as_of_date: string;
  net_managed_money: number;
  net_mm_pct_oi: number;
  position_zscore_26w?: number;
  position_zscore_52w?: number;
  position_percentile_52w?: number;
  weekly_change_net_mm?: number;
  crowding_score: number;
  squeeze_risk_score: number;
  reversal_risk_score: number;
  trend_support_score: number;
  model_version: string;
}

export interface CommodityHistoryResponse {
  slug: string;
  snapshots: SignalSnapshot[];
  count: number;
}

export interface RankingEntry {
  slug: string;
  name: string;
  score: number;
}

export interface CommodityRankingsResponse {
  crowding: RankingEntry[];
  squeeze_risk: RankingEntry[];
  reversal_risk: RankingEntry[];
  trend_support: RankingEntry[];
}

// --- Alerts ---

export interface AlertListResponse {
  alerts: Alert[];
  count: number;
}

export interface Alert {
  id: number;
  commodity_slug: string;
  commodity_name: string;
  as_of_date: string;
  severity: "critical" | "warning" | "info";
  alert_type: string;
  headline: string;
  summary: string;
  regime_label: string;
  regime_confidence: number;
  final_alert_score: number;
  is_active: boolean;
  created_at: string;
}

export interface AlertDetailResponse extends Alert {
  explanation: ExplanationPayload;
}

export interface ExplanationPayload {
  factors: ExplanationFactor[];
  regime_context: {
    label: string;
    confidence: number;
    impact: string;
    is_transitioning: boolean;
  };
  positioning_context: {
    net_mm_pct_oi: number;
    zscore_52w: number;
    percentile_52w: number;
    weekly_change: number;
  };
}

export interface ExplanationFactor {
  name: string;
  score: number;
  weight: number;
  weighted: number;
  detail: string;
}

// --- Freshness ---

export interface FreshnessResponse {
  sources: SourceFreshness[];
}

export interface SourceFreshness {
  source: string;
  last_ingested?: string;
  positions_as_of?: string;
  status: "ok" | "no_data" | "stale";
}
