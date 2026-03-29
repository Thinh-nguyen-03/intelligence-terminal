"use client";

import { useEffect, useState } from "react";
import { useParams } from "next/navigation";
import Link from "next/link";
import { getAlertDetail } from "@/lib/api";
import { Panel } from "@/components/common/Panel";
import { SeverityBadge } from "@/components/common/Badge";
import { ScoreBar } from "@/components/common/ScoreBar";
import { regimeColor, formatPercent, formatCompactNumber, formatAlertType, severityBorder } from "@/lib/format";
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
    return <div className="text-text-muted p-8 tracking-widest">LOADING<span className="blink">_</span></div>;
  }

  if (!alert) {
    return <div className="text-red p-8">Alert not found</div>;
  }

  const exp = alert.explanation;

  return (
    <div className="max-w-4xl mx-auto space-y-3">
      {/* Breadcrumb */}
      <div className="flex items-center gap-3 text-[11px]">
        <Link href="/alerts" className="text-text-muted hover:text-amber transition-colors uppercase tracking-wide">
          ← Alerts
        </Link>
        <span className="text-terminal-border">/</span>
        <span className="text-text-muted uppercase tracking-wide">{formatAlertType(alert.alert_type)}</span>
      </div>

      {/* Headline card */}
      <div className={`panel p-4 ${severityBorder(alert.severity)}`}>
        <div className="flex items-center gap-3 mb-3">
          <SeverityBadge severity={alert.severity}>{alert.severity}</SeverityBadge>
          <span className="text-amber font-medium text-[13px]">{alert.commodity_name}</span>
          <span className="text-[10px] text-text-muted uppercase tracking-wide">{formatAlertType(alert.alert_type)}</span>
          <span className="ml-auto text-[11px] text-text-muted tabular-nums">{alert.as_of_date}</span>
        </div>
        <h1 className="text-[15px] text-text-primary font-bold mb-2">{alert.headline}</h1>
        <p className="text-[12px] text-text-secondary leading-relaxed">{alert.summary}</p>
        <div className="flex gap-4 mt-3 text-[11px] text-text-muted border-t border-terminal-border pt-3">
          <span>
            Commodity:{" "}
            <Link href={`/commodities/${alert.commodity_slug}`} className="text-amber hover:text-amber-bright">
              {alert.commodity_name}
            </Link>
          </span>
          <span>
            Final Score: <span className="text-text-primary font-bold">{(alert.final_alert_score * 100).toFixed(0)}</span>
          </span>
        </div>
      </div>

      <div className="grid grid-cols-2 gap-3">
        {/* Score factors */}
        <Panel title="Score Breakdown">
          {exp?.factors ? (
            <div className="space-y-4">
              {exp.factors.map((f) => (
                <div key={f.name}>
                  <div className="flex justify-between text-[10px] text-text-muted uppercase tracking-wide mb-1">
                    <span>{f.name}</span>
                    <span>Weight {(f.weight * 100).toFixed(0)}%</span>
                  </div>
                  <ScoreBar label="" score={f.score} />
                  <div className="text-[11px] text-text-secondary mt-1 leading-relaxed">
                    {f.detail}
                  </div>
                </div>
              ))}
            </div>
          ) : (
            <div className="text-text-muted py-4">No explanation data</div>
          )}
        </Panel>

        <div className="space-y-3">
          {/* Regime context */}
          <Panel title="Regime Context">
            {exp?.regime_context ? (
              <div className="space-y-2.5">
                <Row label="Regime">
                  <span className={regimeColor(exp.regime_context.label)}>
                    {exp.regime_context.label.replace(" (transitioning)", "")}
                  </span>
                </Row>
                <Row label="Confidence">
                  <span>{formatPercent(exp.regime_context.confidence)}</span>
                </Row>
                <Row label="Impact">
                  <span className="uppercase text-[10px] tracking-wide">{exp.regime_context.impact}</span>
                </Row>
                <Row label="Transitioning">
                  <span className={exp.regime_context.is_transitioning ? "text-amber" : "text-text-muted"}>
                    {exp.regime_context.is_transitioning ? "Yes" : "No"}
                  </span>
                </Row>
              </div>
            ) : (
              <div className="text-text-muted py-4">No regime context</div>
            )}
          </Panel>

          {/* Positioning context */}
          <Panel title="Positioning Context">
            {exp?.positioning_context ? (
              <div className="space-y-2.5">
                <Row label="Net MM %OI">
                  <span className={exp.positioning_context.net_mm_pct_oi >= 0 ? "num-positive" : "num-negative"}>
                    {formatPercent(exp.positioning_context.net_mm_pct_oi)}
                  </span>
                </Row>
                <Row label="Z-Score 52w">
                  <span>{exp.positioning_context.zscore_52w.toFixed(2)}</span>
                </Row>
                <Row label="Percentile 52w">
                  <span>{formatPercent(exp.positioning_context.percentile_52w, 0)}</span>
                </Row>
                <Row label="Weekly Change">
                  <span>{formatCompactNumber(exp.positioning_context.weekly_change)}</span>
                </Row>
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

function Row({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <div className="flex items-center justify-between text-[12px] border-b border-terminal-border/40 pb-2">
      <span className="text-text-muted uppercase tracking-wide text-[10px]">{label}</span>
      <span className="text-text-primary">{children}</span>
    </div>
  );
}
