"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { getAlerts } from "@/lib/api";
import { Panel } from "@/components/common/Panel";
import { SeverityBadge } from "@/components/common/Badge";
import { formatScore, regimeColor } from "@/lib/format";
import type { Alert } from "@/types/api";

const SEVERITY_FILTERS = ["all", "critical", "warning", "info"] as const;

export default function AlertsPage() {
  const [alerts, setAlerts] = useState<Alert[]>([]);
  const [loading, setLoading] = useState(true);
  const [filter, setFilter] = useState<string>("all");

  useEffect(() => {
    getAlerts()
      .then((data) => setAlerts(data.alerts))
      .catch(() => {})
      .finally(() => setLoading(false));
  }, []);

  const filtered =
    filter === "all" ? alerts : alerts.filter((a) => a.severity === filter);

  if (loading) {
    return <div className="text-text-muted p-8">Loading alerts...<span className="blink">_</span></div>;
  }

  return (
    <div className="space-y-3 max-w-5xl mx-auto">
      <div className="flex items-center justify-between">
        <h1 className="text-green text-[16px] font-bold">ACTIVE ALERTS</h1>
        <div className="flex gap-1">
          {SEVERITY_FILTERS.map((s) => (
            <button
              key={s}
              onClick={() => setFilter(s)}
              className={`px-3 py-1 text-[10px] uppercase tracking-wider border rounded transition-colors ${
                filter === s
                  ? "text-green border-green bg-green/10"
                  : "text-text-muted border-terminal-border hover:border-terminal-border-bright"
              }`}
            >
              {s}
              {s !== "all" && (
                <span className="ml-1 text-text-muted">
                  ({alerts.filter((a) => a.severity === s).length})
                </span>
              )}
            </button>
          ))}
        </div>
      </div>

      <Panel title={`${filtered.length} alerts`}>
        {filtered.length === 0 ? (
          <div className="text-text-muted text-center py-8">No alerts matching filter</div>
        ) : (
          <div className="space-y-2">
            {filtered.map((alert) => (
              <Link
                key={alert.id}
                href={`/alerts/${alert.id}`}
                className="block p-3 bg-terminal-bg border border-terminal-border rounded hover:border-terminal-border-bright transition-colors"
              >
                <div className="flex items-center gap-3 mb-2">
                  <SeverityBadge severity={alert.severity}>
                    {alert.severity}
                  </SeverityBadge>
                  <span className="text-green text-[12px] font-medium">
                    {alert.commodity_name}
                  </span>
                  <span className="text-[10px] text-text-muted">
                    {alert.alert_type.replace(/_/g, " ")}
                  </span>
                  <span className="ml-auto text-[11px] text-text-muted tabular-nums">
                    Score: {formatScore(alert.final_alert_score)}
                  </span>
                  <span className="text-[10px] text-text-muted">{alert.as_of_date}</span>
                </div>

                <div className="text-[13px] text-text-primary mb-1">
                  {alert.headline}
                </div>
                <div className="text-[11px] text-text-secondary line-clamp-2">
                  {alert.summary}
                </div>

                <div className="flex items-center gap-4 mt-2 text-[10px] text-text-muted">
                  <span>
                    Regime:{" "}
                    <span className={regimeColor(alert.regime_label)}>
                      {alert.regime_label}
                    </span>
                  </span>
                  <span>
                    Confidence: {alert.regime_confidence.toFixed(1)}%
                  </span>
                </div>
              </Link>
            ))}
          </div>
        )}
      </Panel>
    </div>
  );
}
