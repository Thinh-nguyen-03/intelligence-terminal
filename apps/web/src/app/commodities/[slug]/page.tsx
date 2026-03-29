"use client";

import { useEffect, useState } from "react";
import { useParams } from "next/navigation";
import Link from "next/link";
import { getCommodityDetail, getCommodityHistory } from "@/lib/api";
import { Panel } from "@/components/common/Panel";
import { ScoreBar } from "@/components/common/ScoreBar";
import { SeverityBadge } from "@/components/common/Badge";
import { formatPercent, formatCompactNumber, formatDate, severityBorder, formatAlertType } from "@/lib/format";
import type { CommodityDetailResponse, SignalSnapshot } from "@/types/api";

export default function CommodityDetailPage() {
  const params = useParams();
  const slug = params.slug as string;
  const [detail, setDetail] = useState<CommodityDetailResponse | null>(null);
  const [history, setHistory] = useState<SignalSnapshot[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    async function load() {
      try {
        const [d, h] = await Promise.allSettled([
          getCommodityDetail(slug),
          getCommodityHistory(slug),
        ]);
        if (d.status === "fulfilled") setDetail(d.value);
        if (h.status === "fulfilled") setHistory(h.value.snapshots);
      } finally {
        setLoading(false);
      }
    }
    load();
  }, [slug]);

  if (loading) {
    return <div className="text-text-muted p-8 tracking-widest">LOADING<span className="blink">_</span></div>;
  }

  if (!detail) {
    return <div className="text-red p-8">Commodity not found: {slug}</div>;
  }

  const sig = detail.latest_signal;

  return (
    <div className="space-y-3">
      {/* Header */}
      <div className="flex items-center gap-3">
        <Link href="/commodities" className="text-text-muted hover:text-amber text-[11px] uppercase tracking-wide transition-colors">
          ← Commodities
        </Link>
        <span className="text-terminal-border">/</span>
        <h1 className="text-amber text-[16px] font-bold">{detail.commodity.name}</h1>
        <span className="text-[10px] text-text-muted uppercase tracking-widest">{detail.commodity.group_name}</span>
      </div>

      <div className="grid grid-cols-12 gap-3">
        {/* Signal Scores */}
        <div className="col-span-4 space-y-3">
          <Panel title="Positioning Scores">
            {sig ? (
              <div className="space-y-3">
                <ScoreBar label="Crowding" score={sig.crowding_score} />
                <ScoreBar label="Squeeze Risk" score={sig.squeeze_risk_score} />
                <ScoreBar label="Reversal Risk" score={sig.reversal_risk_score} />
                <ScoreBar label="Trend Support" score={sig.trend_support_score} />
              </div>
            ) : (
              <div className="text-text-muted py-4">No signal data</div>
            )}
          </Panel>
        </div>

        {/* Key Metrics */}
        <div className="col-span-4">
          <Panel title="Key Metrics">
            {sig ? (
              <div className="grid grid-cols-2 gap-2">
                <MetricBox label="Net MM %OI" value={formatPercent(sig.net_mm_pct_oi)} positive={sig.net_mm_pct_oi >= 0} />
                <MetricBox label="Net Managed Money" value={formatCompactNumber(sig.net_managed_money)} positive={sig.net_managed_money >= 0} />
                <MetricBox label="Z-Score 52w" value={sig.position_zscore_52w?.toFixed(2) ?? "--"} positive={(sig.position_zscore_52w ?? 0) >= 0} />
                <MetricBox label="Percentile 52w" value={sig.position_percentile_52w != null ? formatPercent(sig.position_percentile_52w, 0) : "--"} />
                <MetricBox label="Z-Score 26w" value={sig.position_zscore_26w?.toFixed(2) ?? "--"} positive={(sig.position_zscore_26w ?? 0) >= 0} />
                <MetricBox label="Weekly Change" value={sig.weekly_change_net_mm != null ? formatCompactNumber(sig.weekly_change_net_mm) : "--"} positive={(sig.weekly_change_net_mm ?? 0) >= 0} />
              </div>
            ) : (
              <div className="text-text-muted py-4">No signal data</div>
            )}
          </Panel>
        </div>

        {/* Active Alerts */}
        <div className="col-span-4">
          <Panel title={`Alerts (${detail.active_alerts.length})`}>
            {detail.active_alerts.length === 0 ? (
              <div className="text-text-muted py-4 text-center">No active alerts</div>
            ) : (
              <div className="space-y-2">
                {detail.active_alerts.map((a) => (
                  <Link
                    key={a.id}
                    href={`/alerts/${a.id}`}
                    className={`block p-2.5 bg-terminal-bg border border-terminal-border rounded hover:border-terminal-border-bright transition-colors ${severityBorder(a.severity)}`}
                  >
                    <div className="flex items-center gap-2 mb-1">
                      <SeverityBadge severity={a.severity}>{a.severity}</SeverityBadge>
                      <span className="text-[10px] text-text-muted uppercase tracking-wide">{formatAlertType(a.alert_type)}</span>
                      <span className="ml-auto text-[10px] text-text-muted tabular-nums">{a.as_of_date}</span>
                    </div>
                    <div className="text-[11px] text-text-secondary">{a.headline}</div>
                  </Link>
                ))}
              </div>
            )}
          </Panel>
        </div>
      </div>

      {/* History Table */}
      <Panel title="Signal History">
        {history.length === 0 ? (
          <div className="text-text-muted py-4 text-center">No history available</div>
        ) : (
          <div className="overflow-x-auto max-h-[400px] overflow-auto">
            <table className="w-full text-[11px]">
              <thead className="sticky top-0 bg-terminal-panel">
                <tr className="text-[10px] text-text-muted uppercase tracking-wider border-b border-terminal-border">
                  <th className="text-left py-2 pr-3">Date</th>
                  <th className="text-right py-2 px-2">Net MM %OI</th>
                  <th className="text-right py-2 px-2">Z-52w</th>
                  <th className="text-right py-2 px-2">Pctile</th>
                  <th className="text-right py-2 px-2">Crowding</th>
                  <th className="text-right py-2 px-2">Squeeze</th>
                  <th className="text-right py-2 px-2">Reversal</th>
                  <th className="text-right py-2 px-2">Trend</th>
                </tr>
              </thead>
              <tbody>
                {history.map((s) => (
                  <tr key={s.as_of_date} className="border-b border-terminal-border/30 hover:bg-terminal-border/20">
                    <td className="py-1.5 pr-3 text-text-muted tabular-nums">{formatDate(s.as_of_date)}</td>
                    <td className="text-right py-1.5 px-2 tabular-nums">
                      <span className={s.net_mm_pct_oi >= 0 ? "num-positive" : "num-negative"}>
                        {formatPercent(s.net_mm_pct_oi)}
                      </span>
                    </td>
                    <td className="text-right py-1.5 px-2 tabular-nums">
                      {s.position_zscore_52w != null
                        ? <span className={s.position_zscore_52w >= 0 ? "num-positive" : "num-negative"}>{s.position_zscore_52w.toFixed(2)}</span>
                        : "--"}
                    </td>
                    <td className="text-right py-1.5 px-2 tabular-nums text-text-secondary">
                      {s.position_percentile_52w != null ? formatPercent(s.position_percentile_52w, 0) : "--"}
                    </td>
                    <td className="text-right py-1.5 px-2 tabular-nums">{(s.crowding_score * 100).toFixed(0)}</td>
                    <td className="text-right py-1.5 px-2 tabular-nums">{(s.squeeze_risk_score * 100).toFixed(0)}</td>
                    <td className="text-right py-1.5 px-2 tabular-nums">{(s.reversal_risk_score * 100).toFixed(0)}</td>
                    <td className="text-right py-1.5 px-2 tabular-nums">{(s.trend_support_score * 100).toFixed(0)}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </Panel>
    </div>
  );
}

function MetricBox({ label, value, positive }: { label: string; value: string; positive?: boolean }) {
  const colorClass = positive === undefined ? "text-text-primary" : positive ? "num-positive" : "num-negative";
  return (
    <div className="bg-terminal-bg border border-terminal-border rounded p-2.5">
      <div className="text-[10px] text-text-muted uppercase tracking-wider mb-1">{label}</div>
      <div className={`text-[15px] font-bold tabular-nums ${colorClass}`}>{value}</div>
    </div>
  );
}
