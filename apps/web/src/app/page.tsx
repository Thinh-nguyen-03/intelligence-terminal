"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { RegimeCard } from "@/components/dashboard/RegimeCard";
import { AlertsFeed } from "@/components/dashboard/AlertsFeed";
import { CommodityGrid } from "@/components/dashboard/CommodityGrid";
import { Panel } from "@/components/common/Panel";
import { scoreColor, formatScore } from "@/lib/format";
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
  RankingEntry,
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
        <div className="text-text-muted text-[13px] tracking-widest">
          INTEL TERMINAL — LOADING<span className="blink">_</span>
        </div>
      </div>
    );
  }

  return (
    <div className="flex flex-col gap-3 h-full">
      {/* Row 1 — main panels */}
      <div className="grid grid-cols-12 gap-3 flex-1 min-h-0">
        <div className="col-span-3 overflow-auto">
          <RegimeCard regime={regime} />
        </div>
        <div className="col-span-6 overflow-auto">
          <CommodityGrid commodities={commodities} />
        </div>
        <div className="col-span-3 overflow-auto">
          <AlertsFeed alerts={alerts} />
        </div>
      </div>

      {/* Row 2 — rankings strip */}
      {rankings && (
        <div className="grid grid-cols-4 gap-3 flex-shrink-0">
          <MiniRanking
            title="Crowding"
            entries={rankings.crowding}
            href="/rankings"
          />
          <MiniRanking
            title="Squeeze Risk"
            entries={rankings.squeeze_risk}
            href="/rankings"
          />
          <MiniRanking
            title="Reversal Risk"
            entries={rankings.reversal_risk}
            href="/rankings"
          />
          <MiniRanking
            title="Trend Support"
            entries={rankings.trend_support}
            href="/rankings"
          />
        </div>
      )}
    </div>
  );
}

function MiniRanking({
  title,
  entries,
  href,
}: {
  title: string;
  entries: RankingEntry[];
  href: string;
}) {
  return (
    <Panel
      title={title}
      action={
        <Link href={href} className="text-[10px] text-amber hover:text-amber-bright">
          ALL
        </Link>
      }
    >
      <div className="space-y-1">
        {entries.slice(0, 5).map((e, i) => (
          <div key={e.slug} className="flex items-center gap-2 text-[11px]">
            <span className="w-3 text-text-muted text-[10px]">{i + 1}</span>
            <Link
              href={`/commodities/${e.slug}`}
              className="flex-1 text-text-secondary hover:text-amber transition-colors truncate"
            >
              {e.name}
            </Link>
            {/* mini bar */}
            <div className="w-12 h-[4px] bg-terminal-border rounded-full overflow-hidden">
              <div
                className={`h-full rounded-full ${e.score >= 0.7 ? "bg-red" : e.score >= 0.4 ? "bg-amber" : "bg-text-muted"}`}
                style={{ width: `${Math.min(e.score * 100, 100)}%` }}
              />
            </div>
            <span className={`w-7 text-right tabular-nums font-medium text-[11px] ${scoreColor(e.score)}`}>
              {formatScore(e.score)}
            </span>
          </div>
        ))}
      </div>
    </Panel>
  );
}
