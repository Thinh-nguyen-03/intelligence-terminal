"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { getCommodities } from "@/lib/api";
import type { Commodity } from "@/types/api";

export default function CommoditiesPage() {
  const [commodities, setCommodities] = useState<Commodity[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    getCommodities()
      .then((data) => setCommodities(data.commodities))
      .catch(() => {})
      .finally(() => setLoading(false));
  }, []);

  if (loading) {
    return <div className="text-text-muted p-8">Loading commodities...<span className="blink">_</span></div>;
  }

  // Group by group_name
  const groups = commodities.reduce<Record<string, Commodity[]>>((acc, c) => {
    (acc[c.group_name] ||= []).push(c);
    return acc;
  }, {});

  return (
    <div className="max-w-4xl mx-auto space-y-6">
      <h1 className="text-green text-[16px] font-bold">COMMODITIES</h1>
      {Object.entries(groups).map(([group, items]) => (
        <div key={group}>
          <h2 className="text-[12px] text-text-muted uppercase tracking-wider mb-2 border-b border-terminal-border pb-1">
            {group}
          </h2>
          <div className="grid grid-cols-3 gap-2">
            {items.map((c) => (
              <Link
                key={c.slug}
                href={`/commodities/${c.slug}`}
                className="panel p-3 hover:border-terminal-border-bright transition-colors"
              >
                <div className="text-green text-[13px] font-medium">{c.name}</div>
                <div className="text-[10px] text-text-muted mt-1">{c.slug}</div>
              </Link>
            ))}
          </div>
        </div>
      ))}
    </div>
  );
}
