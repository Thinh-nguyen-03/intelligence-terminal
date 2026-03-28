"use client";

import { severityColor, severityBg } from "@/lib/format";

interface BadgeProps {
  severity: string;
  children: React.ReactNode;
}

export function SeverityBadge({ severity, children }: BadgeProps) {
  return (
    <span
      className={`inline-block px-2 py-0.5 text-[10px] uppercase tracking-wider border rounded ${severityColor(severity)} ${severityBg(severity)}`}
    >
      {children}
    </span>
  );
}

export function StatusDot({ status }: { status: "ok" | "no_data" | "stale" | string }) {
  const color =
    status === "ok" ? "bg-green" : status === "stale" ? "bg-amber" : "bg-text-muted";
  return <span className={`inline-block w-2 h-2 rounded-full ${color}`} />;
}
