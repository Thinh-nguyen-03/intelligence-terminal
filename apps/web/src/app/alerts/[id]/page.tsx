"use client";

import { useEffect, useState } from "react";
import { useParams } from "next/navigation";
import Link from "next/link";
import { getAlertDetail } from "@/lib/api";
import { Panel } from "@/components/common/Panel";
import { SeverityBadge } from "@/components/common/Badge";
import { ScoreBar } from "@/components/common/ScoreBar";
import { regimeColor, formatPercent, formatCompactNumber } from "@/lib/format";
import type { AlertDetailResponse } from "@/types/api";

export default function AlertDetailPage() {
  const params = useParams();
  const id = Number(params.id);
  const [alert, setAlert] = useState<AlertDetailResponse | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    getAlertDetail(id)
      .then(setAlert)
      .catch(() => {})
      .finally(() => setLoading(false));
  }, [id]);

  if (loading) {
    return <div className="text-text-muted p-8">Loading alert...<span className="blink">_</span></div>;
  }

  if (!alert) {
    return <div className="text-red p-8">Alert not found</div>;
  }

  const exp = alert.explanation;

  return (
    <div className="max-w-4xl mx-auto space-y-3">
      {/* Header */}
      <div className="flex items-center gap-4">
        <Link href="/alerts" className="text-text-muted hover:text-text-primary text-[12px]">
          &larr; ALERTS
        </Link>
        <SeverityBadge severity={alert.severity}>{alert.severity}</SeverityBadge>
        <span className="text-text-muted text-[12px]">{alert.alert_type.replace(/_/g, " ")}</span>
        <span className="ml-auto text-[11px] text-text-muted">{alert.as_of_date}</span>
      </div>

      {/* Headline */}
      <div className="panel p-4">
        <h1 className="text-[16px] text-text-primary font-bold mb-2">{alert.headline}</h1>
        <p className="text-[13px] text-text-secondary leading-relaxed">{alert.summary}</p>
        <div className="flex gap-4 mt-3 text-[11px]">
          <span className="text-text-muted">
            Commodity:{" "}
            <Link href={`/commodities/${alert.commodity_slug}`} className="text-green hover:text-green-bright">
              {alert.commodity_name}
            </Link>
          </span>
          <span className="text-text-muted">
            Final Score: <span className="text-text-primary font-bold">{(alert.final_alert_score * 100).toFixed(0)}</span>
          </span>
        </div>
      </div>

      <div className="grid grid-cols-2 gap-3">
        {/* Score Factors */}
        <Panel title="Score Breakdown">
          {exp?.factors ? (
            <div className="space-y-3">
              {exp.factors.map((f) => (
                <div key={f.name}>
                  <ScoreBar label={f.name} score={f.score} />
                  <div className="text-[10px] text-text-muted ml-[108px] mt-0.5">
                    Weight: {(f.weight * 100).toFixed(0)}% | Weighted: {(f.weighted * 100).toFixed(0)}
                  </div>
                  <div className="text-[10px] text-text-secondary ml-[108px]">
                    {f.detail}
                  </div>
                </div>
              ))}
            </div>
          ) : (
            <div className="text-text-muted py-4">No explanation data</div>
          )}
        </Panel>

        {/* Context */}
        <div className="space-y-3">
          <Panel title="Regime Context">
            {exp?.regime_context ? (
              <div className="space-y-2 text-[12px]">
                <div className="flex justify-between">
                  <span className="text-text-muted">Regime</span>
                  <span className={regimeColor(exp.regime_context.label)}>
                    {exp.regime_context.label}
                  </span>
                </div>
                <div className="flex justify-between">
                  <span className="text-text-muted">Confidence</span>
                  <span>{formatPercent(exp.regime_context.confidence)}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-text-muted">Impact</span>
                  <span>{exp.regime_context.impact}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-text-muted">Transitioning</span>
                  <span>{exp.regime_context.is_transitioning ? "Yes" : "No"}</span>
                </div>
              </div>
            ) : (
              <div className="text-text-muted py-4">No regime context</div>
            )}
          </Panel>

          <Panel title="Positioning Context">
            {exp?.positioning_context ? (
              <div className="space-y-2 text-[12px]">
                <div className="flex justify-between">
                  <span className="text-text-muted">Net MM %OI</span>
                  <span className={exp.positioning_context.net_mm_pct_oi >= 0 ? "num-positive" : "num-negative"}>
                    {formatPercent(exp.positioning_context.net_mm_pct_oi)}
                  </span>
                </div>
                <div className="flex justify-between">
                  <span className="text-text-muted">Z-Score 52w</span>
                  <span>{exp.positioning_context.zscore_52w.toFixed(2)}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-text-muted">Percentile 52w</span>
                  <span>{formatPercent(exp.positioning_context.percentile_52w, 0)}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-text-muted">Weekly Change</span>
                  <span>{formatCompactNumber(exp.positioning_context.weekly_change)}</span>
                </div>
              </div>
            ) : (
              <div className="text-text-muted py-4">No positioning context</div>
            )}
          </Panel>
        </div>
      </div>
    </div>
  );
}
