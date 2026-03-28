"use client";

import { useEffect, useState } from "react";
import { getFreshness } from "@/lib/api";
import type { SourceFreshness } from "@/types/api";
import { StatusDot } from "@/components/common/Badge";
import { timeAgo } from "@/lib/format";

export function StatusBar() {
  const [sources, setSources] = useState<SourceFreshness[]>([]);

  useEffect(() => {
    getFreshness()
      .then((data) => setSources(data.sources))
      .catch(() => {});
  }, []);

  return (
    <footer className="border-t border-terminal-border bg-terminal-panel h-7 flex items-center px-4 text-[10px] text-text-muted gap-6">
      <span className="text-text-secondary">DATA STATUS</span>
      {sources.length === 0 ? (
        <span>Loading...</span>
      ) : (
        sources.map((s) => (
          <span key={s.source} className="flex items-center gap-1.5">
            <StatusDot status={s.status} />
            <span>{s.source}</span>
            {s.last_ingested && (
              <span className="text-text-muted">{timeAgo(s.last_ingested)}</span>
            )}
          </span>
        ))
      )}
      <span className="ml-auto flex items-center gap-1.5">
        <span className="w-1.5 h-1.5 rounded-full bg-green blink" />
        <span>CONNECTED</span>
      </span>
    </footer>
  );
}
