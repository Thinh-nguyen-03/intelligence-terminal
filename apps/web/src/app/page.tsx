"use client";

import { useEffect, useState } from "react";
import { RegimeCard } from "@/components/dashboard/RegimeCard";
import { AlertsFeed } from "@/components/dashboard/AlertsFeed";
import { CommodityGrid } from "@/components/dashboard/CommodityGrid";
import { TopMovers } from "@/components/dashboard/TopMovers";
import {
  getRegimeCurrent,
  getAlerts,
  getCommodities,
  getCommodityDetail,
  getCommodityRankings,
} from "@/lib/api";
import type {
  RegimeCurrentResponse,
  Alert,
  CommodityDetailResponse,
  CommodityRankingsResponse,
} from "@/types/api";

export default function DashboardPage() {
  const [regime, setRegime] = useState<RegimeCurrentResponse | null>(null);
  const [alerts, setAlerts] = useState<Alert[]>([]);
  const [commodities, setCommodities] = useState<CommodityDetailResponse[]>([]);
  const [rankings, setRankings] = useState<CommodityRankingsResponse | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    async function load() {
      try {
        const [regimeData, alertsData, commoditiesData, rankingsData] =
          await Promise.allSettled([
            getRegimeCurrent(),
            getAlerts(),
            getCommodities().then(async (list) => {
              const details = await Promise.all(
                list.commodities.map((c) => getCommodityDetail(c.slug))
              );
              return details;
            }),
            getCommodityRankings(),
          ]);

        if (regimeData.status === "fulfilled") setRegime(regimeData.value);
        if (alertsData.status === "fulfilled") setAlerts(alertsData.value.alerts);
        if (commoditiesData.status === "fulfilled") setCommodities(commoditiesData.value);
        if (rankingsData.status === "fulfilled") setRankings(rankingsData.value);
      } finally {
        setLoading(false);
      }
    }
    load();
  }, []);

  if (loading) {
    return (
      <div className="flex items-center justify-center h-full">
        <div className="text-text-muted text-[14px]">
          <span className="text-amber">INTEL TERMINAL</span> Loading...
          <span className="blink">_</span>
        </div>
      </div>
    );
  }

  return (
    <div className="grid grid-cols-12 gap-3 h-full">
      {/* Left column: Regime + Rankings */}
      <div className="col-span-3 space-y-3 overflow-auto">
        <RegimeCard regime={regime} />
        {rankings && (
          <>
            <TopMovers title="Top Crowding" entries={rankings.crowding} />
            <TopMovers title="Top Squeeze Risk" entries={rankings.squeeze_risk} />
          </>
        )}
      </div>

      {/* Center column: Commodity Grid */}
      <div className="col-span-6 overflow-auto">
        <CommodityGrid commodities={commodities} />
      </div>

      {/* Right column: Alerts */}
      <div className="col-span-3 overflow-auto">
        <AlertsFeed alerts={alerts} />
        {rankings && (
          <>
            <div className="mt-3">
              <TopMovers title="Top Reversal Risk" entries={rankings.reversal_risk} />
            </div>
            <div className="mt-3">
              <TopMovers title="Top Trend Support" entries={rankings.trend_support} />
            </div>
          </>
        )}
      </div>
    </div>
  );
}
