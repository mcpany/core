import * as React from "react"
import { cn } from "@/lib/utils"
import { CheckCircle2, XCircle, AlertTriangle, Activity } from "lucide-react"

type StatusType = "active" | "inactive" | "error" | "warning"

interface StatusBadgeProps extends React.HTMLAttributes<HTMLDivElement> {
  status: StatusType | boolean
  text?: string
}

export function StatusBadge({ status, text, className, ...props }: StatusBadgeProps) {
  let type: StatusType = "inactive"
  if (typeof status === 'boolean') {
      type = status ? "active" : "inactive"
  } else {
      type = status
  }

  const config = {
    active: { icon: CheckCircle2, color: "text-emerald-500", bg: "bg-emerald-500/10", border: "border-emerald-500/20", label: "Active" },
    inactive: { icon: XCircle, color: "text-slate-500", bg: "bg-slate-500/10", border: "border-slate-500/20", label: "Inactive" },
    error: { icon: AlertTriangle, color: "text-red-500", bg: "bg-red-500/10", border: "border-red-500/20", label: "Error" },
    warning: { icon: Activity, color: "text-amber-500", bg: "bg-amber-500/10", border: "border-amber-500/20", label: "Warning" },
  }

  const current = config[type]
  const Icon = current.icon

  return (
    <div
      className={cn(
        "inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-full border text-xs font-medium",
        current.bg,
        current.color,
        current.border,
        className
      )}
      {...props}
    >
      <Icon className="w-3 h-3" />
      {text || current.label}
    </div>
  )
}
