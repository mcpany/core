/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { CheckCircle2, XCircle, Clock, ArrowRight, Activity, Loader2 } from "lucide-react";
import { Trace } from "@/app/api/traces/route";
import { cn } from "@/lib/utils";
import { useDashboardDensity } from "@/contexts/dashboard-density-context";

/**
 * RecentActivityWidget component.
 * Displays the most recent tool executions.
 * @returns The rendered component.
 */
export function RecentActivityWidget() {
  const [traces, setTraces] = useState<Trace[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const { density } = useDashboardDensity();
  const isCompact = density === "compact";

  const fetchTraces = async () => {
    try {
      const res = await fetch('/api/traces');
      if (!res.ok) {
          throw new Error(`Failed to fetch traces: ${res.status}`);
      }
      const data = await res.json();
      // Take top 5
      setTraces(data.slice(0, 5));
      setError(null);
    } catch (err) {
      console.error("Failed to load recent activity", err);
      setError("Failed to load activity.");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchTraces();
    const interval = setInterval(fetchTraces, 5000);
    return () => clearInterval(interval);
  }, []);

  const formatTime = (timestamp: string) => {
      const date = new Date(timestamp);
      const now = new Date();
      const diffMs = now.getTime() - date.getTime();
      const diffSec = Math.floor(diffMs / 1000);
      const diffMin = Math.floor(diffSec / 60);

      if (diffSec < 60) return "Just now";
      if (diffMin < 60) return `${diffMin}m ago`;
      const diffHour = Math.floor(diffMin / 60);
      if (diffHour < 24) return `${diffHour}h ago`;
      return date.toLocaleDateString();
  };

  const getDurationColor = (ms: number) => {
      if (ms > 1000) return "text-amber-500";
      return "text-muted-foreground";
  };

  return (
    <Card className="col-span-3 backdrop-blur-sm bg-background/50">
      <CardHeader className={cn(
          "flex flex-row items-center justify-between space-y-0",
          isCompact ? "p-3 pb-1" : "pb-2"
      )}>
        <div className="space-y-1">
            <CardTitle className="text-base font-medium flex items-center gap-2">
                <Activity className="h-4 w-4 text-primary" />
                Recent Activity
            </CardTitle>
            <CardDescription className={isCompact ? "text-[10px]" : ""}>
                Real-time monitor of tool executions.
            </CardDescription>
        </div>
        <Link href="/traces" className="text-xs text-muted-foreground hover:text-primary flex items-center gap-1 transition-colors">
            View All <ArrowRight className="h-3 w-3" />
        </Link>
      </CardHeader>
      <CardContent className={isCompact ? "p-3 pt-0" : ""}>
        {loading && traces.length === 0 ? (
            <div className={cn("flex items-center justify-center text-muted-foreground", isCompact ? "h-[150px]" : "h-[200px]")}>
                <Loader2 className="h-6 w-6 animate-spin mr-2" /> Loading activity...
            </div>
        ) : error && traces.length === 0 ? (
            <div className={cn("flex items-center justify-center text-destructive", isCompact ? "h-[150px]" : "h-[200px]")}>
                {error}
            </div>
        ) : traces.length === 0 ? (
            <div className={cn("flex flex-col items-center justify-center text-muted-foreground", isCompact ? "h-[150px]" : "h-[200px]")}>
                <Clock className="h-8 w-8 mb-2 opacity-20" />
                <p>No recent activity recorded.</p>
                <p className="text-xs opacity-70 mt-1">Execute a tool to see it here.</p>
            </div>
        ) : (
            <div className={cn(isCompact ? "space-y-2" : "space-y-4")}>
                {traces.map((trace) => (
                    <div key={trace.id} className={cn("flex items-center justify-between border-b last:border-0 last:pb-0", isCompact ? "pb-2" : "pb-4")}>
                        <div className={cn("flex items-center", isCompact ? "gap-2" : "gap-4")}>
                            <div className={cn("rounded-full bg-muted/50",
                                isCompact ? "p-1.5" : "p-2",
                                trace.status === 'success' ? "text-green-500 bg-green-500/10" :
                                trace.status === 'error' ? "text-red-500 bg-red-500/10" : "text-yellow-500"
                            )}>
                                {trace.status === 'success' ? <CheckCircle2 className="h-4 w-4" /> :
                                 trace.status === 'error' ? <XCircle className="h-4 w-4" /> : <Clock className="h-4 w-4" />}
                            </div>
                            <div className="space-y-1">
                                <div className="text-sm font-medium leading-none flex items-center gap-2">
                                    {trace.rootSpan.name.replace('POST /', '').replace('GET /', '')}
                                    {trace.status === 'error' && (
                                        <Badge variant="destructive" className="text-[10px] h-4 px-1">Failed</Badge>
                                    )}
                                </div>
                                <div className="flex items-center gap-2 text-xs text-muted-foreground">
                                    <span>{formatTime(trace.timestamp)}</span>
                                    <span>•</span>
                                    <span className={getDurationColor(trace.totalDuration)}>{trace.totalDuration.toFixed(0)}ms</span>
                                </div>
                            </div>
                        </div>
                        <Button variant="ghost" size="sm" className="h-8 w-8 p-0" asChild>
                            <Link href={`/traces?id=${trace.id}`}>
                                <ArrowRight className="h-4 w-4" />
                            </Link>
                        </Button>
                    </div>
                ))}
            </div>
        )}
      </CardContent>
    </Card>
  );
}
