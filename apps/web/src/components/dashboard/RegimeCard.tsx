"use client";

import { Panel } from "@/components/common/Panel";
import { regimeColor, regimeBgColor, formatPercent } from "@/lib/format";
import type { RegimeCurrentResponse } from "@/types/api";

interface Props {
  regime: RegimeCurrentResponse | null;
}

const FACTOR_CONFIG = [
  { key: "growth",    label: "Growth",    bar: "bg-green" },
  { key: "inflation", label: "Inflation", bar: "bg-amber" },
  { key: "labor",     label: "Labor",     bar: "bg-blue" },
  { key: "stress",    label: "Stress",    bar: "bg-red" },
] as const;

export function RegimeCard({ regime }: Props) {
  if (!regime) {
    return (
      <Panel title="Macro Regime">
        <div className="text-text-muted text-center py-12">No regime data</div>
      </Panel>
    );
  }

  return (
    <Panel title="Macro Regime">
      <div className="space-y-5">

        {/* Regime label — dominant */}
        <div className={`rounded p-4 border border-terminal-border ${regimeBgColor(regime.regime_label)}`}>
          <div className={`text-[17px] font-bold leading-snug ${regimeColor(regime.regime_label)}`}>
            {/* Strip "(transitioning)" suffix for the headline */}
            {regime.regime_label.replace(" (transitioning)", "")}
          </div>
          {regime.is_transitioning && (
            <div className="text-[10px] text-amber mt-1 uppercase tracking-widest">
              Transitioning
            </div>
          )}
          {regime.transition_detail && (
            <div className="text-[11px] text-text-secondary mt-1 uppercase tracking-wide">
              {regime.transition_detail}
            </div>
          )}
        </div>

        {/* Confidence */}
        <div className="flex items-center justify-between text-[12px]">
          <span className="text-text-muted uppercase tracking-wider text-[10px]">Confidence</span>
          <span className="text-[20px] font-bold tabular-nums text-text-primary">
            {formatPercent(regime.confidence)}
          </span>
        </div>

        {/* Confidence bar */}
        <div className="h-1.5 bg-terminal-border rounded-full overflow-hidden -mt-3">
          <div
            className={`h-full rounded-full ${regimeColor(regime.regime_label).replace("text-", "bg-")}`}
            style={{ width: `${regime.confidence}%` }}
          />
        </div>

        {/* As of / Prior */}
        <div className="text-[11px] space-y-1 border-t border-terminal-border pt-3">
          <div className="flex justify-between">
            <span className="text-text-muted uppercase tracking-wide text-[10px]">As of</span>
            <span className="text-text-secondary tabular-nums">{regime.as_of_date}</span>
          </div>
          {regime.prior_label && (
            <div className="flex justify-between">
              <span className="text-text-muted uppercase tracking-wide text-[10px]">Prior</span>
              <span className={`text-[11px] ${regimeColor(regime.prior_label)}`}>
                {regime.prior_label.replace(" (transitioning)", "")}
              </span>
            </div>
          )}
        </div>

        {/* Factor scores */}
        <div className="space-y-2.5 border-t border-terminal-border pt-3">
          <div className="text-[10px] text-text-muted uppercase tracking-wider">Factor Scores</div>
          {FACTOR_CONFIG.map((f) => {
            const raw = regime.factor_scores[f.key];
            const abs = Math.abs(raw);
            const pct = Math.min(abs * 200, 100); // scale ±0.5 → 100%
            return (
              <div key={f.key} className="space-y-1">
                <div className="flex justify-between text-[11px]">
                  <span className="text-text-secondary uppercase tracking-wide text-[10px]">{f.label}</span>
                  <span className={`tabular-nums font-medium ${raw >= 0 ? "text-text-primary" : "text-text-muted"}`}>
                    {raw >= 0 ? "+" : ""}{(raw * 100).toFixed(1)}
                  </span>
                </div>
                <div className="h-1 bg-terminal-border rounded-full overflow-hidden">
                  <div
                    className={`h-full rounded-full ${abs < 0.05 ? "bg-text-muted" : f.bar}`}
                    style={{ width: `${pct}%` }}
                  />
                </div>
              </div>
            );
          })}
        </div>

        <div className="text-[10px] text-text-muted border-t border-terminal-border pt-2">
          Model {regime.model_version}
        </div>
      </div>
    </Panel>
  );
}
