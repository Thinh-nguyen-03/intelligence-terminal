"use client";

import Link from "next/link";
import { Panel } from "@/components/common/Panel";
import { scoreColor, scoreCellBg, formatScore, formatPercent } from "@/lib/format";
import type { CommodityDetailResponse } from "@/types/api";

interface Props {
  commodities: CommodityDetailResponse[];
}

function ScoreCell({ score }: { score: number }) {
  return (
    <td className={`text-right py-2 px-2 tabular-nums font-medium ${scoreColor(score)} ${scoreCellBg(score)}`}>
      {formatScore(score)}
    </td>
  );
}

export function CommodityGrid({ commodities }: Props) {
  return (
    <Panel title="Commodity Signals">
      {commodities.length === 0 ? (
        <div className="text-text-muted text-center py-6">
          No commodity data available
        </div>
      ) : (
        <div className="overflow-x-auto">
          <table className="w-full text-[12px]">
            <thead>
              <tr className="text-[10px] text-text-muted uppercase tracking-wider border-b border-terminal-border">
                <th className="text-left py-2 pr-4">Commodity</th>
                <th className="text-right py-2 px-2">Net MM %OI</th>
                <th className="text-right py-2 px-2">Z-Score</th>
                <th className="text-right py-2 px-2">Pctile</th>
                <th className="text-right py-2 px-2">Crowding</th>
                <th className="text-right py-2 px-2">Squeeze</th>
                <th className="text-right py-2 px-2">Reversal</th>
                <th className="text-right py-2 px-2">Trend</th>
              </tr>
            </thead>
            <tbody>
              {commodities.map((c) => {
                const sig = c.latest_signal;
                return (
                  <tr
                    key={c.commodity.slug}
                    className="border-b border-terminal-border/50 hover:bg-terminal-border/20 transition-colors"
                  >
                    <td className="py-2 pr-4">
                      <Link
                        href={`/commodities/${c.commodity.slug}`}
                        className="text-amber hover:text-amber-bright transition-colors"
                      >
                        {c.commodity.name}
                      </Link>
                      <div className="text-[10px] text-text-muted uppercase tracking-wide mt-0.5">
                        {c.commodity.group_name}
                      </div>
                    </td>
                    {sig ? (
                      <>
                        <td className="text-right py-2 px-2 tabular-nums">
                          <span className={sig.net_mm_pct_oi >= 0 ? "num-positive" : "num-negative"}>
                            {formatPercent(sig.net_mm_pct_oi)}
                          </span>
                        </td>
                        <td className="text-right py-2 px-2 tabular-nums">
                          {sig.position_zscore_52w != null ? (
                            <span className={sig.position_zscore_52w >= 0 ? "num-positive" : "num-negative"}>
                              {sig.position_zscore_52w.toFixed(2)}
                            </span>
                          ) : (
                            <span className="text-text-muted">--</span>
                          )}
                        </td>
                        <td className="text-right py-2 px-2 tabular-nums text-text-secondary">
                          {sig.position_percentile_52w != null
                            ? formatPercent(sig.position_percentile_52w, 0)
                            : "--"}
                        </td>
                        <ScoreCell score={sig.crowding_score} />
                        <ScoreCell score={sig.squeeze_risk_score} />
                        <ScoreCell score={sig.reversal_risk_score} />
                        <ScoreCell score={sig.trend_support_score} />
                      </>
                    ) : (
                      <td colSpan={7} className="text-center py-2 text-text-muted">
                        No signal data
                      </td>
                    )}
                  </tr>
                );
              })}
            </tbody>
          </table>
        </div>
      )}
    </Panel>
  );
}
