"use client";

interface ScoreBarProps {
  label: string;
  score: number;
  maxScore?: number;
  colorClass?: string;
}

export function ScoreBar({ label, score, maxScore = 1, colorClass }: ScoreBarProps) {
  const pct = Math.min((score / maxScore) * 100, 100);
  const color = colorClass || (score >= 0.7 ? "bg-red" : score >= 0.4 ? "bg-amber" : "bg-blue");

  return (
    <div className="flex items-center gap-3 text-[12px]">
      <span className="w-24 text-text-secondary truncate">{label}</span>
      <div className="flex-1 h-[6px] bg-terminal-border rounded-sm overflow-hidden">
        <div
          className={`h-full rounded-sm transition-all ${color}`}
          style={{ width: `${pct}%` }}
        />
      </div>
      <span className="w-10 text-right tabular-nums">{(score * 100).toFixed(0)}</span>
    </div>
  );
}
