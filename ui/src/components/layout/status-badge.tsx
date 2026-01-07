
import { cn } from "@/lib/utils";

interface StatusBadgeProps {
  status: "active" | "inactive" | "error" | "warning" | "pending";
  text?: string;
  className?: string;
}

export function StatusBadge({ status, text, className }: StatusBadgeProps) {
  const styles = {
    active: "bg-green-500/15 text-green-700 dark:text-green-400 border-green-500/20",
    inactive: "bg-slate-500/15 text-slate-700 dark:text-slate-400 border-slate-500/20",
    error: "bg-red-500/15 text-red-700 dark:text-red-400 border-red-500/20",
    warning: "bg-yellow-500/15 text-yellow-700 dark:text-yellow-400 border-yellow-500/20",
    pending: "bg-blue-500/15 text-blue-700 dark:text-blue-400 border-blue-500/20",
  };

  const indicatorStyles = {
    active: "bg-green-500",
    inactive: "bg-slate-500",
    error: "bg-red-500",
    warning: "bg-yellow-500",
    pending: "bg-blue-500",
  };

  return (
    <span
      className={cn(
        "inline-flex items-center gap-1.5 rounded-full border px-2.5 py-0.5 text-xs font-medium transition-colors",
        styles[status],
        className
      )}
    >
      <span className={cn("h-1.5 w-1.5 rounded-full", indicatorStyles[status])} />
      {text || status.charAt(0).toUpperCase() + status.slice(1)}
    </span>
  );
}
