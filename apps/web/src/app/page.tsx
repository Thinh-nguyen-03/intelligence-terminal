"use client";

import { useEffect, useState } from "react";
import { RegimeCard } from "@/components/dashboard/RegimeCard";
import { AlertsFeed } from "@/components/dashboard/AlertsFeed";
import { CommodityGrid } from "@/components/dashboard/CommodityGrid";
import {
  getRegimeCurrent,
  getAlerts,
  getCommodities,
  getCommodityDetail,
} from "@/lib/api";
import type {
  RegimeCurrentResponse,
  Alert,
  CommodityDetailResponse,
} from "@/types/api";

export default function DashboardPage() {
  const [regime, setRegime] = useState<RegimeCurrentResponse | null>(null);
  const [alerts, setAlerts] = useState<Alert[]>([]);
  const [commodities, setCommodities] = useState<CommodityDetailResponse[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    async function load() {
      try {
        const [regimeData, alertsData, commoditiesData] =
          await Promise.allSettled([
            getRegimeCurrent(),
            getAlerts(),
            getCommodities().then(async (list) => {
              const details = await Promise.all(
                list.commodities.map((c) => getCommodityDetail(c.slug))
              );
              return details;
            }),
          ]);

        if (regimeData.status === "fulfilled") setRegime(regimeData.value);
        if (alertsData.status === "fulfilled") setAlerts(alertsData.value.alerts);
        if (commoditiesData.status === "fulfilled") setCommodities(commoditiesData.value);
      } finally {
        setLoading(false);
      }
    }
    load();
  }, []);

  if (loading) {
    return (
      <div className="flex items-center justify-center h-full">
        <div className="text-text-muted text-[13px] tracking-widest">
          <span className="text-amber">INTEL TERMINAL</span> — LOADING
          <span className="blink">_</span>
        </div>
      </div>
    );
  }

  return (
    <div className="grid grid-cols-12 gap-3 h-full">
      {/* Left — Regime card, full height */}
      <div className="col-span-3 overflow-auto">
        <RegimeCard regime={regime} />
      </div>

      {/* Center — Commodity signal grid */}
      <div className="col-span-6 overflow-auto">
        <CommodityGrid commodities={commodities} />
      </div>

      {/* Right — Alerts feed */}
      <div className="col-span-3 overflow-auto">
        <AlertsFeed alerts={alerts} />
      </div>
    </div>
  );
}
