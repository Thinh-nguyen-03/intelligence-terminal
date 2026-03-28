"use client";

import Link from "next/link";
import { Panel } from "@/components/common/Panel";
import { SeverityBadge } from "@/components/common/Badge";
import type { Alert } from "@/types/api";

interface Props {
  alerts: Alert[];
}

export function AlertsFeed({ alerts }: Props) {
  return (
    <Panel
      title="Active Alerts"
      action={
        <Link href="/alerts" className="text-[10px] text-amber hover:text-amber-bright">
          VIEW ALL
        </Link>
      }
    >
      {alerts.length === 0 ? (
        <div className="text-text-muted text-center py-6">No active alerts</div>
      ) : (
        <div className="space-y-2 max-h-[400px] overflow-auto">
          {alerts.slice(0, 10).map((alert) => (
            <Link
              key={alert.id}
              href={`/alerts/${alert.id}`}
              className="block p-2.5 bg-terminal-bg border border-terminal-border rounded hover:border-terminal-border-bright transition-colors"
            >
              <div className="flex items-center gap-2 mb-1">
                <SeverityBadge severity={alert.severity}>
                  {alert.severity}
                </SeverityBadge>
                <span className="text-[11px] text-text-muted">
                  {alert.commodity_name}
                </span>
                <span className="ml-auto text-[10px] text-text-muted tabular-nums">
                  {alert.as_of_date}
                </span>
              </div>
              <div className="text-[12px] text-text-primary leading-snug">
                {alert.headline}
              </div>
            </Link>
          ))}
        </div>
      )}
    </Panel>
  );
}
