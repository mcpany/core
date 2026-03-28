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
 * Intent: Document RecentActivityWidget
 *
 * Params:
 *   - None
 *
 * Returns:
 *   - Documented below.
 *
 * Errors:
 *   - None
 *
 * Side Effects:
 *   - None
 *
 * RecentActivityWidget component.
 * Displays the most recent tool executions.
 * @returns The rendered component.
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
    <Card className="col-span-3 bg-background/80 backdrop-blur-md shadow-sm overflow-hidden flex flex-col border border-border/40">
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

                    const reqPayload = trace.rootSpan.attributes?.['mcp.request_payload'] || trace.rootSpan.input;
                    const resPayload = trace.rootSpan.attributes?.['mcp.response_payload'] || trace.rootSpan.output;
                    const errMessage = trace.rootSpan.attributes?.['error.message'] || trace.rootSpan.errorMessage;

                    const isError = trace.status === "error" || trace.rootSpan.status === "error";
                    const isSuccess = trace.status === "success" || trace.rootSpan.status === "success";
                    const hasResponseDiff = trace.rootSpan.attributes?.['mcp.response_diff'] !== undefined || (trace.rootSpan.input && trace.rootSpan.input['mcp.response_diff'] !== undefined);
                    const diffContent = trace.rootSpan.attributes?.['mcp.response_diff'] || (trace.rootSpan.input && trace.rootSpan.input['mcp.response_diff']);


                    return (
                        <div key={trace.id} className="relative group">
                            {/* Timeline connector line for mobile */}
                            {index !== traces.length - 1 && (
                                <div className="absolute left-4 top-8 bottom-[-16px] w-px bg-border/50 z-0 sm:hidden" />
                            )}
                            <div
                                className={cn(
                                    "flex flex-col sm:flex-row gap-3 sm:gap-4 p-3 rounded-xl transition-all duration-300 ease-in-out border",
                                    "hover:bg-muted/30 hover:border-border/60 cursor-pointer shadow-sm",
                                    isExpanded ? "bg-muted/30 border-border/60" : "bg-background/50 border-border/40 hover:scale-[1.01] hover:shadow-md",
                                    isError && "hover:bg-destructive/5 border-destructive/10 hover:border-destructive/20"
                                )}
                                onClick={() => toggleExpand(trace.id)}
                            >
                                {/* Status Indicator */}
                                <div className="flex items-start shrink-0 pt-1 relative z-10">
                                    <div className={cn(
                                        "rounded-full p-2 ring-4 ring-background shadow-sm transition-transform duration-300 group-hover:scale-110",
                                        isSuccess ? "text-emerald-500 bg-emerald-500/10 border border-emerald-500/20" :
                                        isError ? "text-destructive bg-destructive/10 border border-destructive/20" : "text-amber-500 bg-amber-500/10 border border-amber-500/20"
                                    )}>
                                        {isSuccess ? <CheckCircle2 className="h-4 w-4" /> :
                                         isError ? <XCircle className="h-4 w-4" /> : <Clock className="h-4 w-4" />}
                                    </div>
                                    {/* Pulsating ring for the most recent item */}
                                    {index === 0 && (
                                        <div className={cn(
                                            "absolute inset-0 rounded-full animate-ping opacity-30",
                                            isSuccess ? "bg-emerald-400" : isError ? "bg-destructive/80" : "bg-amber-400"
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
                                        <div className="overflow-hidden space-y-3 pt-1">
                                            {/* Request Payload */}
                                            {reqPayload && (
                                                <div className="space-y-1.5">
                                                    <div className="text-[11px] font-semibold tracking-wide uppercase text-muted-foreground flex items-center gap-1.5">
                                                        <Code2 className="h-3 w-3" /> Request Payload
                                                    </div>
                                                    <div className="bg-[#1e1e1e] dark:bg-[#0d0d0d] rounded-md border border-border/40 overflow-hidden shadow-inner p-1">
                                                        <RichResultViewer result={safeParsePayload(reqPayload)} />
                                                    </div>
                                                </div>
                                            )}

                                            {/* Error Message */}
                                            {isError && errMessage && (
                                                <div className="space-y-1.5">
                                                    <div className="text-[11px] font-semibold tracking-wide uppercase text-destructive flex items-center gap-1.5">
                                                        <XCircle className="h-3 w-3" /> Error Details
                                                    </div>
                                                    <div className="bg-destructive/10 rounded-md p-3 border border-destructive/20 overflow-x-auto text-destructive">
                                                        <pre className="text-[12px] font-mono whitespace-pre-wrap break-all leading-relaxed">
                                                            {errMessage}
                                                        </pre>
                                                    </div>
                                                </div>
                                            )}

                                            {/* Response Payload (if not diff) */}
                                            {!hasResponseDiff && resPayload && !isError && (
                                                <div className="space-y-1.5">
                                                    <div className="text-[11px] font-semibold tracking-wide uppercase text-emerald-600 dark:text-emerald-500 flex items-center gap-1.5">
                                                        <CheckCircle2 className="h-3 w-3" /> Response Payload
                                                    </div>
                                                    <div className="bg-[#1e1e1e] dark:bg-[#0d0d0d] rounded-md border border-emerald-500/20 overflow-hidden shadow-inner p-1">
                                                        <RichResultViewer result={safeParsePayload(resPayload)} />
                                                    </div>
                                                </div>
                                            )}

                                            {/* Response Diff (Premium Feature) */}
                                            {hasResponseDiff && (
                                                 <div className="space-y-1.5">
                                                    <div className="text-[11px] font-semibold tracking-wide uppercase text-blue-600 dark:text-blue-500 flex items-center gap-1.5">
                                                        <Activity className="h-3 w-3" /> Inline Diff
                                                    </div>
                                                    <div className="bg-[#1e1e1e] dark:bg-[#0d0d0d] rounded-md border border-border/50 overflow-hidden shadow-inner">
                                                        <pre className="text-[12px] font-mono whitespace-pre-wrap break-all m-0 leading-[1.6]">
                                                            {/* Premium rendering of a unified diff */}
                                                            {String(diffContent).split('\n').map((line, i) => {
                                                                if (line.startsWith('+')) {
                                                                    return <div key={i} className="bg-emerald-500/10 text-emerald-400 px-3 py-0.5 border-l-2 border-emerald-500"><span className="select-none opacity-50 mr-2">+</span>{line.substring(1)}</div>;
                                                                }
                                                                if (line.startsWith('-')) {
                                                                    return <div key={i} className="bg-destructive/10 text-destructive-foreground/80 dark:text-red-400 px-3 py-0.5 border-l-2 border-destructive/80"><span className="select-none opacity-50 mr-2">-</span>{line.substring(1)}</div>;
                                                                }
                                                                if (line.startsWith('@')) {
                                                                    return <div key={i} className="bg-blue-500/5 text-blue-400 px-3 py-1 opacity-80 text-[11px] mt-1 mb-1 border-t border-b border-border/20"><span className="select-none">{line}</span></div>;
                                                                }
                                                                return <div key={i} className="px-3 py-0.5 text-foreground/80 opacity-90 border-l-2 border-transparent"><span className="select-none opacity-0 mr-2"> </span>{line}</div>;
                                                            })}
                                                        </pre>
                                                    </div>
                                                </div>
                                            )}

                                            <div className="flex justify-end pt-3 pb-1">
                                                <Button variant="outline" size="sm" className="h-8 text-xs font-medium shadow-sm" asChild>
                                                    <Link to={`/traces?id=${trace.id}`}>Inspect Trace Details</Link>
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
