export function formatScore(score: number): string {
  return (score * 100).toFixed(0);
}

export function formatPercent(value: number, decimals = 1): string {
  return `${value.toFixed(decimals)}%`;
}

export function formatNumber(value: number): string {
  return new Intl.NumberFormat("en-US").format(value);
}

export function formatCompactNumber(value: number): string {
  if (Math.abs(value) >= 1_000_000) {
    return `${(value / 1_000_000).toFixed(1)}M`;
  }
  if (Math.abs(value) >= 1_000) {
    return `${(value / 1_000).toFixed(1)}K`;
  }
  return value.toFixed(0);
}

export function formatDate(dateStr: string): string {
  return dateStr; // already YYYY-MM-DD from API
}

export function formatDateTime(dateStr: string): string {
  const d = new Date(dateStr);
  return d.toLocaleString("en-US", {
    month: "short",
    day: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  });
}

export function timeAgo(dateStr: string): string {
  const now = Date.now();
  const then = new Date(dateStr).getTime();
  const diff = now - then;
  const minutes = Math.floor(diff / 60000);
  if (minutes < 60) return `${minutes}m ago`;
  const hours = Math.floor(minutes / 60);
  if (hours < 24) return `${hours}h ago`;
  const days = Math.floor(hours / 24);
  return `${days}d ago`;
}

export function scoreColor(score: number): string {
  if (score >= 0.7) return "text-red";
  if (score >= 0.4) return "text-amber";
  return "text-text-muted";
}

// Background heat tint for score cells in tables
export function scoreCellBg(score: number): string {
  if (score >= 0.7) return "bg-red/10";
  if (score >= 0.4) return "bg-amber/10";
  return "";
}

// Format alert_type slug → readable ALL CAPS label
export function formatAlertType(alertType: string): string {
  return alertType.replace(/_/g, " ").toUpperCase();
}

// Format a label as ALL CAPS
export function label(text: string): string {
  return text.toUpperCase();
}

// Left border color class for alert severity
export function severityBorder(severity: string): string {
  switch (severity) {
    case "critical": return "border-l-4 border-l-red";
    case "warning":  return "border-l-4 border-l-amber";
    case "info":     return "border-l-4 border-l-blue";
    default:         return "border-l-4 border-l-terminal-border";
  }
}

export function severityColor(severity: string): string {
  switch (severity) {
    case "critical": return "text-severity-critical";
    case "warning": return "text-severity-warning";
    case "info": return "text-severity-info";
    default: return "text-text-secondary";
  }
}

export function severityBg(severity: string): string {
  switch (severity) {
    case "critical": return "bg-red/10 border-red/30";
    case "warning": return "bg-amber/10 border-amber/30";
    case "info": return "bg-blue/10 border-blue/30";
    default: return "bg-terminal-panel border-terminal-border";
  }
}

export function regimeColor(label: string): string {
  const l = label.toLowerCase();
  if (l.includes("inflationary growth")) return "text-regime-inflationary-growth";
  if (l.includes("inflationary slowdown")) return "text-regime-inflationary-slowdown";
  if (l.includes("disinflationary")) return "text-regime-disinflationary-slowdown";
  if (l.includes("recovery")) return "text-regime-recovery";
  if (l.includes("stress") || l.includes("defensive")) return "text-regime-stress";
  return "text-text-primary";
}

export function regimeBgColor(label: string): string {
  const l = label.toLowerCase();
  if (l.includes("inflationary growth")) return "bg-regime-inflationary-growth/15";
  if (l.includes("inflationary slowdown")) return "bg-regime-inflationary-slowdown/15";
  if (l.includes("disinflationary")) return "bg-regime-disinflationary-slowdown/15";
  if (l.includes("recovery")) return "bg-regime-recovery/15";
  if (l.includes("stress") || l.includes("defensive")) return "bg-regime-stress/15";
  return "bg-terminal-panel";
}
