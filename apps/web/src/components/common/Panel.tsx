"use client";

interface PanelProps {
  title: string;
  children: React.ReactNode;
  className?: string;
  action?: React.ReactNode;
}

export function Panel({ title, children, className = "", action }: PanelProps) {
  return (
    <div className={`panel ${className}`}>
      <div className="panel-header flex items-center justify-between">
        <span>{title}</span>
        {action}
      </div>
      <div className="p-3">{children}</div>
    </div>
  );
}
