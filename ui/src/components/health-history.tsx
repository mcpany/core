/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useEffect, useState } from "react";
import { apiClient } from "@/lib/client";
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from "@/components/ui/tooltip";
import { cn } from "@/lib/utils";

interface HealthRecord {
  timestamp: string;
  status: boolean;
}

export function HealthHistory({ serviceName }: { serviceName: string }) {
  const [history, setHistory] = useState<HealthRecord[]>([]);

  useEffect(() => {
    apiClient.getServiceHealthHistory(serviceName).then((res) => {
      if (res && res.history) {
        setHistory(res.history);
      }
    });
    // Poll every 5 seconds
    const interval = setInterval(() => {
        apiClient.getServiceHealthHistory(serviceName).then((res) => {
            if (res && res.history) {
              setHistory(res.history);
            }
          });
    }, 5000);
    return () => clearInterval(interval);
  }, [serviceName]);

  if (history.length === 0) {
    return <div className="text-sm text-muted-foreground">No health history available.</div>;
  }

  // Sort by timestamp
  const sortedHistory = [...history].sort((a, b) => new Date(a.timestamp).getTime() - new Date(b.timestamp).getTime());

  // Take last 50 for display
  const displayHistory = sortedHistory.slice(-50);

  return (
    <div className="flex items-center gap-1 h-8 mt-2">
      <TooltipProvider>
        {displayHistory.map((record, i) => (
          <Tooltip key={i}>
            <TooltipTrigger asChild>
              <div
                className={cn(
                  "w-2 h-6 rounded-sm transition-all hover:scale-110 cursor-default",
                  record.status ? "bg-green-500 hover:bg-green-400" : "bg-red-500 hover:bg-red-400"
                )}
              />
            </TooltipTrigger>
            <TooltipContent>
              <p>{record.status ? "Up" : "Down"}</p>
              <p className="text-xs text-muted-foreground">{new Date(record.timestamp).toLocaleString()}</p>
            </TooltipContent>
          </Tooltip>
        ))}
      </TooltipProvider>
    </div>
  );
}
