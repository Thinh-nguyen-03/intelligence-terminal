"use client";

import { Panel } from "@/components/common/Panel";
import { ScoreBar } from "@/components/common/ScoreBar";
import { regimeColor, regimeBgColor, formatPercent } from "@/lib/format";
import type { RegimeCurrentResponse } from "@/types/api";

interface Props {
  regime: RegimeCurrentResponse | null;
}

export function RegimeCard({ regime }: Props) {
  if (!regime) {
    return (
      <Panel title="Macro Regime">
        <div className="text-text-muted text-center py-8">No regime data available</div>
      </Panel>
    );
  }

  const factors = [
    { label: "Growth", score: regime.factor_scores.growth, color: "bg-green" },
    { label: "Inflation", score: regime.factor_scores.inflation, color: "bg-amber" },
    { label: "Labor", score: regime.factor_scores.labor, color: "bg-blue" },
    { label: "Stress", score: regime.factor_scores.stress, color: "bg-red" },
  ];

  return (
    <Panel title="Macro Regime">
      <div className="space-y-4">
        {/* Regime label */}
        <div className="flex items-start justify-between">
          <div>
            <div
              className={`text-[18px] font-bold ${regimeColor(regime.regime_label)}`}
            >
              {regime.regime_label}
            </div>
            {regime.is_transitioning && regime.transition_detail && (
              <div className="text-[11px] text-amber mt-1">
                {regime.transition_detail}
              </div>
            )}
          </div>
          <div className="text-right">
            <div className="text-[10px] text-text-muted uppercase">Confidence</div>
            <div className="text-[16px] font-bold tabular-nums">
              {formatPercent(regime.confidence)}
            </div>
          </div>
        </div>

        {/* Regime badge */}
        <div
          className={`px-3 py-2 rounded text-center text-[11px] border border-terminal-border ${regimeBgColor(regime.regime_label)}`}
        >
          <span className="text-text-muted">As of </span>
          <span>{regime.as_of_date}</span>
          {regime.prior_label && (
            <>
              <span className="text-text-muted"> | Prior: </span>
              <span className={regimeColor(regime.prior_label)}>
                {regime.prior_label}
              </span>
            </>
          )}
        </div>

        {/* Factor scores */}
        <div className="space-y-2">
          <div className="text-[10px] text-text-muted uppercase tracking-wider">
            Factor Scores
          </div>
          {factors.map((f) => (
            <ScoreBar
              key={f.label}
              label={f.label}
              score={Math.abs(f.score)}
              colorClass={f.score >= 0 ? f.color : "bg-text-muted"}
            />
          ))}
        </div>
      </div>
    </Panel>
  );
}
