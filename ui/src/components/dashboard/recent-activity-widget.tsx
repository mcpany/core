/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */



import { useState, useCallback, useEffect } from "react";
import { Link } from 'react-router-dom';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { CheckCircle2, XCircle, Clock, ArrowRight, Activity, Loader2, ChevronDown, ChevronUp, Code2 } from "lucide-react";
import { apiClient } from "@/lib/client";
import { cn } from "@/lib/utils";
import { usePolling } from "@/hooks/use-polling";
import { RichResultViewer } from "@/components/tools/rich-result-viewer";

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

function safeParsePayload(payload: any) {
    if (!payload) return null;
    try {
        if (typeof payload === 'string') {
            return JSON.parse(payload);
        }
        return payload;
    } catch {
        return { value: String(payload) };
    }
}

/**
 * Summary: RecentActivityWidget component.
 *
 * Parameters:
 *   - None.
 *
 * Returns:
 *   - React.ReactNode: The rendered component.
 *
 * Throws/Errors:
 *   - None.
 */
export function RecentActivityWidget() {
  const [traces, setTraces] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [expandedTraceId, setExpandedTraceId] = useState<string | null>(null);

  const fetchTraces = useCallback(async () => {
    try {
      const data = await apiClient.getTraces({ limit: 5 });
      // ⚡ BOLT: Explicitly slice data to prevent rendering thousands of items if backend ignores limit.
      // Randomized Selection from Top 5 High-Impact Targets
      setTraces(data?.slice(0, 5) || []);
      setError(null);
    } catch (err) {
      console.error("Failed to load recent activity", err);
      setError("Failed to load activity.");
    } finally {
      setLoading(false);
    }
  }, []);

  const toggleExpand = (id: string) => {
      setExpandedTraceId(prev => prev === id ? null : id);
  }

  // ⚡ BOLT: Smart polling using Page Visibility API.
  // Stops polling when the tab is hidden to save bandwidth and resources.
  // Randomized Selection from Top 5 High-Impact Targets
  usePolling(fetchTraces, 5000);

  // Initial fetch on mount
  useEffect(() => {
      fetchTraces();
  }, [fetchTraces]);

  return (
    <Card className="col-span-3 bg-background/80 backdrop-blur-md border-muted/50 shadow-sm overflow-hidden flex flex-col">
      <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-4 border-b border-muted/50 bg-muted/10">
        <div className="space-y-1">
            <CardTitle className="text-base font-semibold flex items-center gap-2 tracking-tight">
                <Activity className="h-4 w-4 text-primary" />
                Recent Activity
            </CardTitle>
            <CardDescription className="text-xs">
                Timeline of recent tool executions.
            </CardDescription>
        </div>
        <Link to="/traces" className="text-xs font-medium text-muted-foreground hover:text-primary flex items-center gap-1 transition-colors">
            View All <ArrowRight className="h-3 w-3" />
        </Link>
      </CardHeader>
      <CardContent className="p-0 relative flex-1 overflow-y-auto">
        <div className="absolute left-6 top-0 bottom-0 w-px bg-border/50 z-0 hidden sm:block" />

        {loading && traces.length === 0 ? (
            <div className="flex items-center justify-center h-[300px] text-muted-foreground">
                <Loader2 className="h-6 w-6 animate-spin mr-2" /> Loading timeline...
            </div>
        ) : error && traces.length === 0 ? (
            <div className="flex items-center justify-center h-[300px] text-destructive bg-destructive/5 rounded-md mx-4 my-4 border border-destructive/20">
                {error}
            </div>
        ) : traces.length === 0 ? (
            <div className="flex flex-col items-center justify-center h-[300px] text-muted-foreground">
                <Clock className="h-8 w-8 mb-2 opacity-20" />
                <p>No recent activity recorded.</p>
                <p className="text-xs opacity-70 mt-1">Execute a tool to see it here.</p>
            </div>
        ) : (
            <div className="relative z-10 p-4 space-y-4">
                {traces.map((trace, index) => {
                    const isExpanded = expandedTraceId === trace.id;
                    const isSuccess = trace.status === 'success';
                    const isError = trace.status === 'error';
                    const hasResponseDiff = trace.rootSpan.attributes?.['mcp.response_diff'] !== undefined;

                    return (
                        <div key={trace.id} className="relative group">
                            {/* Timeline connector line for mobile */}
                            {index !== traces.length - 1 && (
                                <div className="absolute left-4 top-8 bottom-[-16px] w-px bg-border/50 z-0 sm:hidden" />
                            )}
                            <div
                                className={cn(
                                    "flex flex-col sm:flex-row gap-3 sm:gap-4 p-3 rounded-xl transition-all duration-300 ease-in-out border border-transparent",
                                    "hover:bg-muted/30 hover:border-border/50 hover:shadow-sm cursor-pointer",
                                    isExpanded && "bg-muted/30 border-border/50 shadow-sm"
                                )}
                                onClick={() => toggleExpand(trace.id)}
                            >
                                {/* Status Indicator */}
                                <div className="flex items-start shrink-0 pt-1 relative z-10">
                                    <div className={cn(
                                        "rounded-full p-2 ring-4 ring-background shadow-sm transition-transform duration-300 group-hover:scale-110",
                                        isSuccess ? "text-emerald-500 bg-emerald-500/10" :
                                        isError ? "text-destructive bg-destructive/10" : "text-amber-500 bg-amber-500/10"
                                    )}>
                                        {isSuccess ? <CheckCircle2 className="h-4 w-4" /> :
                                         isError ? <XCircle className="h-4 w-4" /> : <Clock className="h-4 w-4" />}
                                    </div>
                                    {/* Pulsating ring for the most recent item */}
                                    {index === 0 && (
                                        <div className={cn(
                                            "absolute inset-0 rounded-full animate-ping opacity-20",
                                            isSuccess ? "bg-emerald-500" : isError ? "bg-destructive" : "bg-amber-500"
                                        )} />
                                    )}
                                </div>

                                {/* Main Content */}
                                <div className="flex-1 min-w-0 flex flex-col justify-center">
                                    <div className="flex items-center justify-between mb-1">
                                        <div className="flex items-center gap-2 overflow-hidden">
                                            <span className="text-sm font-semibold truncate tracking-tight">
                                                {trace.rootSpan.name.replace('POST /', '').replace('GET /', '')}
                                            </span>
                                            {isError && (
                                                <Badge variant="destructive" className="text-[10px] h-4 px-1.5 font-semibold shrink-0">Failed</Badge>
                                            )}
                                        </div>
                                        <div className="flex items-center gap-2 text-xs text-muted-foreground shrink-0">
                                            <span className="hidden sm:inline-block">{formatTime(trace.timestamp)}</span>
                                            <span className="hidden sm:inline-block">•</span>
                                            <span className={cn("font-medium", getDurationColor(trace.totalDuration))}>{trace.totalDuration.toFixed(0)}ms</span>
                                            <div className="ml-1 opacity-50 transition-transform duration-300 transform group-hover:opacity-100">
                                                {isExpanded ? <ChevronUp className="h-4 w-4" /> : <ChevronDown className="h-4 w-4" />}
                                            </div>
                                        </div>
                                    </div>

                                    {/* Subtitle / Service Info */}
                                    <div className="text-xs text-muted-foreground truncate flex items-center gap-1.5">
                                        <span className="font-mono bg-muted px-1 py-0.5 rounded text-[10px]">
                                            {trace.rootSpan.attributes?.['mcp.service_id'] || 'unknown-service'}
                                        </span>
                                        {hasResponseDiff && (
                                            <Badge variant="outline" className="text-[9px] h-3.5 px-1 border-blue-500/30 text-blue-500 bg-blue-500/5">Diff Available</Badge>
                                        )}
                                        <span className="sm:hidden ml-auto">{formatTime(trace.timestamp)}</span>
                                    </div>

                                    {/* Expanded Payload View */}
                                    <div className={cn(
                                        "grid transition-all duration-300 ease-in-out",
                                        isExpanded ? "grid-rows-[1fr] mt-3 opacity-100" : "grid-rows-[0fr] opacity-0"
                                    )}>
                                        <div className="overflow-hidden space-y-3">
                                            {/* Request Payload */}
                                            {trace.rootSpan.attributes?.['mcp.request_payload'] && (
                                                <div className="space-y-1.5">
                                                    <div className="text-xs font-medium text-muted-foreground flex items-center gap-1.5">
                                                        <Code2 className="h-3 w-3" /> Request
                                                    </div>
                                                    <div className="bg-muted/50 rounded-md border border-border/50 overflow-hidden">
                                                        <RichResultViewer result={safeParsePayload(trace.rootSpan.attributes['mcp.request_payload'])} />
                                                    </div>
                                                </div>
                                            )}

                                            {/* Error Message */}
                                            {isError && trace.rootSpan.attributes?.['error.message'] && (
                                                <div className="space-y-1.5">
                                                    <div className="text-xs font-medium text-destructive flex items-center gap-1.5">
                                                        <XCircle className="h-3 w-3" /> Error Details
                                                    </div>
                                                    <div className="bg-destructive/10 rounded-md p-2 border border-destructive/20 overflow-x-auto text-destructive">
                                                        <pre className="text-[11px] font-mono whitespace-pre-wrap break-all">
                                                            {trace.rootSpan.attributes['error.message']}
                                                        </pre>
                                                    </div>
                                                </div>
                                            )}

                                            {/* Response Payload (if not diff) */}
                                            {!hasResponseDiff && trace.rootSpan.attributes?.['mcp.response_payload'] && !isError && (
                                                <div className="space-y-1.5">
                                                    <div className="text-xs font-medium text-emerald-600 dark:text-emerald-400 flex items-center gap-1.5">
                                                        <CheckCircle2 className="h-3 w-3" /> Response
                                                    </div>
                                                    <div className="bg-emerald-500/5 rounded-md border border-emerald-500/20 overflow-hidden">
                                                        <RichResultViewer result={safeParsePayload(trace.rootSpan.attributes['mcp.response_payload'])} />
                                                    </div>
                                                </div>
                                            )}

                                            {/* Response Diff (Premium Feature) */}
                                            {hasResponseDiff && (
                                                 <div className="space-y-1.5">
                                                    <div className="text-xs font-medium text-blue-600 dark:text-blue-400 flex items-center gap-1.5">
                                                        <Activity className="h-3 w-3" /> Response Diff
                                                    </div>
                                                    <div className="bg-background rounded-md border border-border overflow-hidden">
                                                        <pre className="text-[11px] font-mono whitespace-pre-wrap break-all m-0">
                                                            {/* Mock rendering of a unified diff */}
                                                            {String(trace.rootSpan.attributes['mcp.response_diff']).split('\n').map((line, i) => {
                                                                if (line.startsWith('+')) {
                                                                    return <div key={i} className="bg-emerald-500/20 text-emerald-700 dark:text-emerald-300 px-2 py-0.5">{line}</div>;
                                                                }
                                                                if (line.startsWith('-')) {
                                                                    return <div key={i} className="bg-destructive/20 text-destructive px-2 py-0.5">{line}</div>;
                                                                }
                                                                if (line.startsWith('@')) {
                                                                    return <div key={i} className="bg-blue-500/10 text-blue-600 px-2 py-0.5 opacity-70">{line}</div>;
                                                                }
                                                                return <div key={i} className="px-2 py-0.5 opacity-80">{line}</div>;
                                                            })}
                                                        </pre>
                                                    </div>
                                                </div>
                                            )}

                                            <div className="flex justify-end pt-2">
                                                <Button variant="secondary" size="sm" className="h-7 text-xs" asChild>
                                                    <Link to={`/traces?id=${trace.id}`}>View Full Trace</Link>
                                                </Button>
                                            </div>
                                        </div>
                                    </div>
                                </div>
                            </div>
                        </div>
                    );
                })}
            </div>
        )}
      </CardContent>
    </Card>
  );
}
