/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from "@/components/ui/tooltip";
import { cn } from "@/lib/utils";
import { formatDistanceToNow } from "date-fns";

interface HealthRecord {
  timestamp: string;
  status: string; // "UP" or "DOWN"
  latency_ms: number;
  error?: string;
}

interface HealthTimelineProps {
  history: HealthRecord[];
  limit?: number;
}

export function HealthTimeline({ history, limit = 50 }: HealthTimelineProps) {
  // Take last N records
  const recent = history ? history.slice(-limit) : [];

  if (recent.length === 0) {
    return <div className="text-sm text-muted-foreground">No health history available.</div>;
  }

  return (
    <div className="flex flex-col gap-2 w-full">
      <div className="flex items-end gap-[2px] h-12 w-full">
        {recent.map((rec, i) => (
          <TooltipProvider key={i}>
            <Tooltip delayDuration={0}>
              <TooltipTrigger asChild>
                <div
                  className={cn(
                    "flex-1 rounded-sm transition-all hover:opacity-80",
                    rec.status === "UP" ? "bg-green-500" : "bg-red-500"
                  )}
                  style={{
                    height: rec.status === "UP"
                        ? `${Math.min(100, Math.max(20, (rec.latency_ms / 100) * 100))}%` // Height based on latency, max 100ms = 100%
                        : "100%", // Full height for error
                    minHeight: "10%"
                  }}
                />
              </TooltipTrigger>
              <TooltipContent className="p-2 text-xs">
                <div className="flex flex-col gap-1">
                  <div className="font-bold flex items-center gap-2">
                    <span className={cn("w-2 h-2 rounded-full", rec.status === "UP" ? "bg-green-500" : "bg-red-500")} />
                    {rec.status}
                  </div>
                  <p className="text-muted-foreground">{new Date(rec.timestamp).toLocaleString()}</p>
                  <p>Latency: <span className="font-mono">{rec.latency_ms}ms</span></p>
                  {rec.error && <p className="text-red-400 font-mono mt-1 max-w-[200px] truncate">{rec.error}</p>}
                </div>
              </TooltipContent>
            </Tooltip>
          </TooltipProvider>
        ))}
      </div>
       <div className="flex justify-between text-xs text-muted-foreground px-1">
        <span>{recent.length > 0 ? formatDistanceToNow(new Date(recent[0].timestamp)) + " ago" : ""}</span>
        <span>Now</span>
      </div>
    </div>
  );
}
