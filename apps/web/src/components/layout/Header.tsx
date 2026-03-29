"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";

const NAV_ITEMS = [
  { href: "/", label: "DASHBOARD" },
  { href: "/commodities", label: "COMMODITIES" },
  { href: "/rankings", label: "RANKINGS" },
  { href: "/alerts", label: "ALERTS" },
];

export function Header() {
  const pathname = usePathname();

  return (
    <header className="border-b border-terminal-border bg-terminal-panel">
      <div className="flex items-center h-10 px-4">
        {/* Logo / Title */}
        <Link href="/" className="flex items-center gap-2 mr-8">
          <span className="text-green font-bold text-[14px] tracking-tight">
            INTEL
          </span>
          <span className="text-amber font-bold text-[14px] tracking-tight">
            TERMINAL
          </span>
          <span className="text-text-muted text-[10px]">v2</span>
        </Link>

        {/* Navigation */}
        <nav className="flex items-center gap-1">
          {NAV_ITEMS.map((item) => {
            const isActive =
              item.href === "/"
                ? pathname === "/"
                : pathname.startsWith(item.href);
            return (
              <Link
                key={item.href}
                href={item.href}
                className={`px-3 py-1.5 text-[11px] tracking-wider transition-colors ${
                  isActive
                    ? "text-green bg-green/10 border-b-2 border-green"
                    : "text-text-secondary hover:text-text-primary hover:bg-terminal-border/50"
                }`}
              >
                {item.label}
              </Link>
            );
          })}
        </nav>

        {/* Right side - clock */}
        <div className="ml-auto text-[11px] text-text-muted tabular-nums">
          <Clock />
        </div>
      </div>
    </header>
  );
}

function Clock() {
  return (
    <span suppressHydrationWarning>
      {new Date().toLocaleString("en-US", {
        weekday: "short",
        month: "short",
        day: "numeric",
        hour: "2-digit",
        minute: "2-digit",
        timeZoneName: "short",
      })}
    </span>
  );
}
