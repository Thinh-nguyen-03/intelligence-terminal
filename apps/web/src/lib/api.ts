import type {
  RegimeCurrentResponse,
  RegimeHistoryResponse,
  CommodityListResponse,
  CommodityDetailResponse,
  CommodityHistoryResponse,
  CommodityRankingsResponse,
  AlertListResponse,
  AlertDetailResponse,
  FreshnessResponse,
} from "@/types/api";

const API_BASE = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

async function fetchAPI<T>(path: string): Promise<T> {
  const res = await fetch(`${API_BASE}${path}`, {
    next: { revalidate: 300 },
  });
  if (!res.ok) {
    throw new Error(`API error: ${res.status} ${res.statusText}`);
  }
  return res.json();
}

// Regime
export function getRegimeCurrent() {
  return fetchAPI<RegimeCurrentResponse>("/api/v1/regime/current");
}

export function getRegimeHistory(from?: string, to?: string) {
  const params = new URLSearchParams();
  if (from) params.set("from", from);
  if (to) params.set("to", to);
  const qs = params.toString();
  return fetchAPI<RegimeHistoryResponse>(`/api/v1/regime/history${qs ? `?${qs}` : ""}`);
}

// Commodities
export function getCommodities() {
  return fetchAPI<CommodityListResponse>("/api/v1/commodities");
}

export function getCommodityDetail(slug: string) {
  return fetchAPI<CommodityDetailResponse>(`/api/v1/commodities/${slug}`);
}

export function getCommodityHistory(slug: string, from?: string, to?: string) {
  const params = new URLSearchParams();
  if (from) params.set("from", from);
  if (to) params.set("to", to);
  const qs = params.toString();
  return fetchAPI<CommodityHistoryResponse>(`/api/v1/commodities/${slug}/history${qs ? `?${qs}` : ""}`);
}

export function getCommodityRankings() {
  return fetchAPI<CommodityRankingsResponse>("/api/v1/commodities/rankings");
}

// Alerts
export function getAlerts(severity?: string, commodity?: string) {
  const params = new URLSearchParams();
  if (severity) params.set("severity", severity);
  if (commodity) params.set("commodity", commodity);
  const qs = params.toString();
  return fetchAPI<AlertListResponse>(`/api/v1/alerts${qs ? `?${qs}` : ""}`);
}

export function getAlertDetail(id: number) {
  return fetchAPI<AlertDetailResponse>(`/api/v1/alerts/${id}`);
}

// Freshness
export function getFreshness() {
  return fetchAPI<FreshnessResponse>("/api/v1/freshness");
}
