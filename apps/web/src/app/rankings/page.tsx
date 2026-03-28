"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { getCommodityRankings } from "@/lib/api";
import { Panel } from "@/components/common/Panel";
import { scoreColor, formatScore } from "@/lib/format";
import type { CommodityRankingsResponse, RankingEntry } from "@/types/api";

export default function RankingsPage() {
  const [rankings, setRankings] = useState<CommodityRankingsResponse | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    getCommodityRankings()
      .then(setRankings)
      .catch(() => {})
      .finally(() => setLoading(false));
  }, []);

  if (loading) {
    return <div className="text-text-muted p-8">Loading rankings...<span className="blink">_</span></div>;
  }

  if (!rankings) {
    return <div className="text-text-muted p-8">No ranking data available</div>;
  }

  return (
    <div className="space-y-3">
      <h1 className="text-amber text-[16px] font-bold">COMMODITY RANKINGS</h1>

      <div className="grid grid-cols-2 gap-3">
        <RankingTable title="Crowding Score" entries={rankings.crowding} />
        <RankingTable title="Squeeze Risk" entries={rankings.squeeze_risk} />
        <RankingTable title="Reversal Risk" entries={rankings.reversal_risk} />
        <RankingTable title="Trend Support" entries={rankings.trend_support} />
      </div>
    </div>
  );
}

function RankingTable({ title, entries }: { title: string; entries: RankingEntry[] }) {
  return (
    <Panel title={title}>
      <table className="w-full text-[12px]">
        <thead>
          <tr className="text-[10px] text-text-muted uppercase tracking-wider border-b border-terminal-border">
            <th className="text-left py-1.5 w-8">#</th>
            <th className="text-left py-1.5">Commodity</th>
            <th className="text-right py-1.5">Score</th>
            <th className="text-right py-1.5 w-24">Bar</th>
          </tr>
        </thead>
        <tbody>
          {entries.map((entry, i) => (
            <tr
              key={entry.slug}
              className="border-b border-terminal-border/30 hover:bg-terminal-border/30"
            >
              <td className="py-1.5 text-text-muted">{i + 1}</td>
              <td className="py-1.5">
                <Link
                  href={`/commodities/${entry.slug}`}
                  className="text-text-primary hover:text-amber transition-colors"
                >
                  {entry.name}
                </Link>
              </td>
              <td className={`text-right py-1.5 tabular-nums font-medium ${scoreColor(entry.score)}`}>
                {formatScore(entry.score)}
              </td>
              <td className="py-1.5 pl-3">
                <div className="h-[5px] bg-terminal-border rounded-sm overflow-hidden">
                  <div
                    className={`h-full rounded-sm ${
                      entry.score >= 0.7 ? "bg-red" : entry.score >= 0.4 ? "bg-amber" : "bg-blue"
                    }`}
                    style={{ width: `${Math.min(entry.score * 100, 100)}%` }}
                  />
                </div>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </Panel>
  );
}
