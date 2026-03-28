"use client";

import Link from "next/link";
import { Panel } from "@/components/common/Panel";
import { scoreColor, formatScore } from "@/lib/format";
import type { RankingEntry } from "@/types/api";

interface Props {
  title: string;
  entries: RankingEntry[];
  limit?: number;
}

export function TopMovers({ title, entries, limit = 5 }: Props) {
  const top = entries.slice(0, limit);

  return (
    <Panel title={title}>
      {top.length === 0 ? (
        <div className="text-text-muted text-center py-4">No data</div>
      ) : (
        <div className="space-y-1.5">
          {top.map((entry, i) => (
            <div
              key={entry.slug}
              className="flex items-center gap-2 text-[12px]"
            >
              <span className="w-4 text-text-muted text-right">{i + 1}</span>
              <Link
                href={`/commodities/${entry.slug}`}
                className="flex-1 text-text-primary hover:text-amber transition-colors truncate"
              >
                {entry.name}
              </Link>
              <span className={`tabular-nums font-medium ${scoreColor(entry.score)}`}>
                {formatScore(entry.score)}
              </span>
            </div>
          ))}
        </div>
      )}
    </Panel>
  );
}
