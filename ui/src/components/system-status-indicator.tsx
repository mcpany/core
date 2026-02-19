/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useEffect, useState } from "react";
import { apiClient, DoctorReport, CheckResult } from "@/lib/client";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { ScrollArea } from "@/components/ui/scroll-area";
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
  SheetTrigger,
} from "@/components/ui/sheet";
import {
  Activity,
  CheckCircle2,
  AlertTriangle,
  XCircle,
  WifiOff,
  Server,
  Cpu,
  Globe,
  Loader2,
  RefreshCw
} from "lucide-react";
import { cn } from "@/lib/utils";
import { ConfigDiffViewer } from "@/components/diagnostics/config-diff-viewer";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";

/**
 * SystemStatusIndicator component.
 * Displays a subtle status icon in the header and opens a detailed report sheet on click.
 * @returns The rendered component.
 */
export function SystemStatusIndicator() {
  const [report, setReport] = useState<DoctorReport | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  const [isOpen, setIsOpen] = useState(false);

  const fetchHealth = async () => {
    setLoading(true);
    try {
      const data = await apiClient.getDoctorStatus();
      setReport(data);
      setError(null);
    } catch (err) {
      console.error("SystemStatusIndicator: Fetch failed", err);
      setError(err instanceof Error ? err.message : "Unknown error");
      setReport(null);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchHealth();
    const interval = setInterval(fetchHealth, 10000); // Poll every 10s
    return () => clearInterval(interval);
  }, []);

  const getStatusColor = () => {
    if (error) return "text-muted-foreground";
    if (!report) return "text-muted-foreground";
    switch (report.status.toLowerCase()) {
      case "healthy":
      case "ok":
        return "text-green-500";
      case "degraded":
      case "warning":
        return "text-yellow-500";
      case "error":
      case "unhealthy":
      case "critical":
        return "text-destructive";
      default:
        return "text-muted-foreground";
    }
  };

  const getStatusIcon = () => {
    if (error) return <WifiOff className="h-4 w-4" />;
    if (!report) return <Activity className="h-4 w-4 animate-pulse" />;
    switch (report.status.toLowerCase()) {
      case "healthy":
      case "ok":
        return <CheckCircle2 className="h-4 w-4" />;
      case "degraded":
      case "warning":
        return <AlertTriangle className="h-4 w-4" />;
      case "error":
      case "unhealthy":
      case "critical":
        return <XCircle className="h-4 w-4" />;
      default:
        return <Activity className="h-4 w-4" />;
    }
  };

  const getCheckIcon = (name: string) => {
    const n = name.toLowerCase();
    if (n.includes("network") || n.includes("connectivity") || n.includes("internet")) return <Globe className="h-4 w-4" />;
    if (n.includes("database") || n.includes("storage")) return <Server className="h-4 w-4" />;
    if (n.includes("memory") || n.includes("cpu") || n.includes("runtime")) return <Cpu className="h-4 w-4" />;
    return <Activity className="h-4 w-4" />;
  };

  return (
    <Sheet open={isOpen} onOpenChange={setIsOpen}>
      <SheetTrigger asChild>
        <Button
          variant="ghost"
          size="sm"
          className={cn("h-8 gap-2 px-2", getStatusColor())}
          title="System Status"
        >
          {getStatusIcon()}
          <span className="text-xs font-medium hidden md:inline-block">
             {error ? "Disconnected" : report?.status === "healthy" ? "Healthy" : report?.status || "Loading"}
          </span>
        </Button>
      </SheetTrigger>
      <SheetContent className="w-full sm:max-w-xl overflow-hidden flex flex-col p-0 bg-background/95 backdrop-blur-md">
        <SheetHeader className="p-6 pb-2 border-b bg-muted/20">
          <div className="flex items-center justify-between">
             <SheetTitle className="flex items-center gap-2">
                <Activity className="h-5 w-5" /> System Status
             </SheetTitle>
             <Button variant="ghost" size="icon" onClick={fetchHealth} disabled={loading} className="h-8 w-8">
                <RefreshCw className={cn("h-4 w-4", loading && "animate-spin")} />
             </Button>
          </div>
          <SheetDescription>
            Real-time diagnostics and environment health check.
          </SheetDescription>
        </SheetHeader>

        <ScrollArea className="flex-1 p-6">
            <div className="space-y-6">
                {error && (
                    <div className="p-4 rounded-lg bg-destructive/10 text-destructive border border-destructive/20 flex items-start gap-3">
                        <WifiOff className="h-5 w-5 mt-0.5" />
                        <div>
                            <h4 className="font-semibold text-sm">Connection Error</h4>
                            <p className="text-xs mt-1">Could not contact the server. Is the backend running?</p>
                            <p className="text-[10px] mt-2 font-mono bg-background/50 p-1 rounded">{error}</p>
                        </div>
                    </div>
                )}

                {report && (
                    <div className="flex flex-col gap-4">
                        <div className="grid grid-cols-2 gap-4">
                            <Card className="bg-muted/30 border-dashed shadow-none">
                                <CardHeader className="p-4 pb-2">
                                    <CardTitle className="text-xs font-medium text-muted-foreground uppercase tracking-wider">Overall Status</CardTitle>
                                </CardHeader>
                                <CardContent className="p-4 pt-0">
                                    <div className={cn("text-xl font-bold capitalize flex items-center gap-2", getStatusColor())}>
                                        {getStatusIcon()}
                                        {report.status}
                                    </div>
                                </CardContent>
                            </Card>
                            <Card className="bg-muted/30 border-dashed shadow-none">
                                <CardHeader className="p-4 pb-2">
                                    <CardTitle className="text-xs font-medium text-muted-foreground uppercase tracking-wider">Last Checked</CardTitle>
                                </CardHeader>
                                <CardContent className="p-4 pt-0">
                                    <div className="text-sm font-mono text-foreground">
                                        {new Date(report.timestamp).toLocaleTimeString()}
                                    </div>
                                    <div className="text-[10px] text-muted-foreground mt-1">
                                        {new Date(report.timestamp).toLocaleDateString()}
                                    </div>
                                </CardContent>
                            </Card>
                        </div>

                        <div className="space-y-3">
                            <h3 className="text-sm font-semibold text-muted-foreground">Checks</h3>
                            {Object.entries(report.checks).map(([name, check]) => {
                                const result = check as CheckResult; // Type assertion
                                const isError = result.status !== "ok";
                                const hasDiff = !!result.diff;

                                return (
                                    <Card key={name} className={cn(
                                        "overflow-hidden transition-all",
                                        isError ? "border-l-4 border-l-destructive" : "border-l-4 border-l-green-500"
                                    )}>
                                        <CardHeader className="p-3 bg-muted/10 flex flex-row items-center justify-between space-y-0">
                                            <div className="flex items-center gap-2">
                                                {getCheckIcon(name)}
                                                <span className="font-medium capitalize text-sm">{name}</span>
                                            </div>
                                            <Badge variant={isError ? "destructive" : "outline"} className="text-[10px] h-5">
                                                {result.status.toUpperCase()}
                                            </Badge>
                                        </CardHeader>
                                        <CardContent className="p-3 text-sm space-y-3">
                                            <p className="text-muted-foreground">{result.message || "No status message."}</p>

                                            {result.latency && (
                                                <div className="flex items-center text-[10px] text-muted-foreground font-mono">
                                                    Latency: {result.latency}
                                                </div>
                                            )}

                                            {hasDiff && (
                                                <div className="mt-2">
                                                    <div className="text-xs font-semibold mb-2 flex items-center gap-2">
                                                        <AlertTriangle className="h-3 w-3 text-yellow-500" />
                                                        Configuration Drift
                                                    </div>
                                                    <ConfigDiffViewer diff={result.diff!} height="200px" />
                                                </div>
                                            )}
                                        </CardContent>
                                    </Card>
                                );
                            })}
                        </div>
                    </div>
                )}
            </div>
        </ScrollArea>
      </SheetContent>
    </Sheet>
  );
}
