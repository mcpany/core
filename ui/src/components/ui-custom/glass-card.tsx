import * as React from "react"
import { cn } from "@/lib/utils"

interface GlassCardProps extends React.HTMLAttributes<HTMLDivElement> {
  children: React.ReactNode
  hoverEffect?: boolean
}

export function GlassCard({ className, children, hoverEffect = false, ...props }: GlassCardProps) {
  return (
    <div
      className={cn(
        "rounded-xl border bg-background/50 backdrop-blur-xl shadow-sm",
        hoverEffect && "transition-all duration-200 hover:shadow-md hover:bg-background/60",
        className
      )}
      {...props}
    >
      {children}
    </div>
  )
}
