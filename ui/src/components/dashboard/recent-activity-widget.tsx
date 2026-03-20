/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */



import { useState, useCallback, useEffect } from "react";
import { Link } from 'react-router-dom';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { CheckCircle2, XCircle, Clock, ArrowRight, Activity, Loader2 } from "lucide-react";
import { apiClient } from "@/lib/client";
import { cn } from "@/lib/utils";
import { usePolling } from "@/hooks/use-polling";

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

/**
 * RecentActivityWidget component.
 * Displays the most recent tool executions.
 * @returns The rendered component.
 */
export function RecentActivityWidget() {
  interface AuditLogEntry {
      timestamp: string;
      toolName: string;
      userId: string;
      profileId: string;
      arguments: string;
      result: string;
      error: string;
      duration: string;
      durationMs: number;
  }

  const [logs, setLogs] = useState<AuditLogEntry[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const fetchLogs = useCallback(async () => {
    try {
      const res = await apiClient.listAuditLogs({ limit: 5 });
      setLogs(res.entries || []);
      setError(null);
    } catch (err) {
      console.error("Failed to load recent activity", err);
      setError("Failed to load activity.");
    } finally {
      setLoading(false);
    }
  }, []);

  // ⚡ BOLT: Smart polling using Page Visibility API.
  // Stops polling when the tab is hidden to save bandwidth and resources.
  // Randomized Selection from Top 5 High-Impact Targets
  usePolling(fetchLogs, 5000);

  // Initial fetch on mount
  useEffect(() => {
      fetchLogs();
  }, [fetchLogs]);

  return (
    <Card className="col-span-3 backdrop-blur-sm bg-background/50">
      <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2 border-b border-border/50 bg-muted/20 rounded-t-xl px-4 py-3">
        <div className="space-y-1">
            <CardTitle className="text-base font-semibold flex items-center gap-2 tracking-tight">
                <Activity className="h-4 w-4 text-primary" />
                Recent Activity
            </CardTitle>
            <CardDescription className="text-xs">
                Historical record of tool executions.
            </CardDescription>
        </div>
        <Link to="/audit" className="text-xs font-medium text-muted-foreground hover:text-primary flex items-center gap-1 transition-colors px-2 py-1 rounded-md hover:bg-muted">
            View Audit Log <ArrowRight className="h-3 w-3" />
        </Link>
      </CardHeader>
      <CardContent className="p-0">
        {loading && logs.length === 0 ? (
            <div className="flex items-center justify-center h-[200px] text-muted-foreground">
                <Loader2 className="h-6 w-6 animate-spin mr-2" /> Loading activity...
            </div>
        ) : error && logs.length === 0 ? (
            <div className="flex items-center justify-center h-[200px] text-destructive">
                {error}
            </div>
        ) : logs.length === 0 ? (
            <div className="flex flex-col items-center justify-center h-[200px] text-muted-foreground">
                <Clock className="h-8 w-8 mb-2 opacity-20" />
                <p>No recent activity recorded.</p>
                <p className="text-xs opacity-70 mt-1">Execute a tool to see it here.</p>
            </div>
        ) : (
            <div className="divide-y divide-border/50">
                {logs.map((log, i) => {
                    const status = log.error ? 'error' : 'success';
                    return (
                    <div key={`${log.timestamp}-${i}`} className="flex items-center justify-between p-4 hover:bg-muted/30 transition-colors group">
                        <div className="flex items-center gap-4">
                            <div className={cn("rounded-full p-2 ring-1 ring-inset shadow-sm",
                                status === 'success' ? "text-green-500 bg-green-500/10 ring-green-500/20" :
                                "text-red-500 bg-red-500/10 ring-red-500/20"
                            )}>
                                {status === 'success' ? <CheckCircle2 className="h-4 w-4" /> :
                                 <XCircle className="h-4 w-4" />}
                            </div>
                            <div className="space-y-1.5">
                                <div className="text-sm font-semibold leading-none flex items-center gap-2">
                                    {log.toolName}
                                    {status === 'error' && (
                                        <Badge variant="destructive" className="text-[10px] h-4 px-1.5 font-medium tracking-wide">FAILED</Badge>
                                    )}
                                </div>
                                <div className="flex items-center gap-2 text-xs font-medium text-muted-foreground/80">
                                    <span>{formatTime(log.timestamp)}</span>
                                    <span className="opacity-50">•</span>
                                    <span className={cn("font-mono", getDurationColor(log.durationMs))}>{log.durationMs}ms</span>
                                    {log.userId && (
                                        <>
                                            <span className="opacity-50">•</span>
                                            <span className="truncate max-w-[100px]" title={log.userId}>{log.userId}</span>
                                        </>
                                    )}
                                </div>
                            </div>
                        </div>
                    </div>
                )})}
            </div>
        )}
      </CardContent>
    </Card>
  );
}
