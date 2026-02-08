/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useEffect } from "react";
import { apiClient, DoctorReport } from "@/lib/client";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import {
  Activity,
  CheckCircle2,
  XCircle,
  AlertTriangle,
  RefreshCw,
  Server,
  Cpu,
  Globe,
  Loader2,
  Clock,
  History
} from "lucide-react";
import { cn } from "@/lib/utils";
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from "@/components/ui/tooltip";

interface ServiceHealthData {
  services: {
    id: string;
    name: string;
    status: string;
    latency: string;
    uptime: string;
    message?: string;
  }[];
  history: Record<string, { timestamp: number; status: string }[]>;
}

/**
 * SystemHealth component.
 * @returns The rendered component.
 */
export function SystemHealth() {
  const [report, setReport] = useState<DoctorReport | null>(null);
  const [serviceHealth, setServiceHealth] = useState<ServiceHealthData | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const fetchHealth = async () => {
    setLoading(true);
    setError(null);
    try {
      const [doctorData, healthData] = await Promise.all([
        apiClient.getDoctorStatus(),
        apiClient.getServiceHealth()
      ]);
      setReport(doctorData);
      setServiceHealth(healthData);
    } catch (err) {
      console.error("Failed to fetch system health", err);
      setError("Failed to retrieve diagnostics report. The backend might be unreachable.");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchHealth();
  }, []);

  const getStatusBadge = (status: string) => {
    switch (status.toLowerCase()) {
      case "ok":
      case "healthy":
        return <Badge variant="default" className="bg-green-600 hover:bg-green-700">Healthy</Badge>;
      case "degraded":
      case "warning":
        return <Badge variant="secondary" className="bg-yellow-500/10 text-yellow-600 hover:bg-yellow-500/20">Degraded</Badge>;
      case "error":
      case "unhealthy":
      case "critical":
        return <Badge variant="destructive">Critical</Badge>;
      default:
        return <Badge variant="outline">Unknown</Badge>;
    }
  };

  const getIconForCheck = (name: string) => {
    const n = name.toLowerCase();
    if (n.includes("network") || n.includes("connectivity") || n.includes("internet")) return <Globe className="h-4 w-4" />;
    if (n.includes("database") || n.includes("storage")) return <Server className="h-4 w-4" />;
    if (n.includes("memory") || n.includes("cpu") || n.includes("runtime")) return <Cpu className="h-4 w-4" />;
    return <Activity className="h-4 w-4" />;
  };

  const getStatusColor = (status: string) => {
      switch (status.toLowerCase()) {
          case 'healthy': return "bg-green-500";
          case 'degraded': return "bg-yellow-500";
          case 'unhealthy': return "bg-red-500";
          case 'inactive': return "bg-gray-500";
          default: return "bg-gray-300";
      }
  };

  if (loading && !report) {
    return (
      <div className="flex flex-col items-center justify-center h-64 space-y-4">
        <Loader2 className="h-8 w-8 animate-spin text-primary" />
        <p className="text-muted-foreground">Running diagnostics...</p>
      </div>
    );
  }

  if (error) {
    return (
      <div className="flex flex-col items-center justify-center h-64 space-y-4">
        <Alert variant="destructive" className="max-w-md">
          <AlertTriangle className="h-4 w-4" />
          <AlertTitle>Diagnostics Failed</AlertTitle>
          <AlertDescription>{error}</AlertDescription>
        </Alert>
        <Button onClick={fetchHealth} variant="outline">
          <RefreshCw className="mr-2 h-4 w-4" />
          Retry
        </Button>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Top Status Card */}
      <Card className="border-l-4 border-l-primary shadow-sm bg-gradient-to-r from-background to-muted/20">
        <CardContent className="p-6 flex items-center justify-between">
          <div className="flex items-center gap-4">
             <div className={cn("p-3 rounded-full bg-muted",
                 report?.status === 'healthy' ? "bg-green-100 dark:bg-green-900/30 text-green-600" :
                 report?.status === 'degraded' ? "bg-yellow-100 dark:bg-yellow-900/30 text-yellow-600" :
                 "bg-red-100 dark:bg-red-900/30 text-red-600"
             )}>
                 {report?.status === 'healthy' ? <CheckCircle2 className="h-8 w-8" /> :
                  report?.status === 'degraded' ? <AlertTriangle className="h-8 w-8" /> :
                  <XCircle className="h-8 w-8" />}
             </div>
             <div>
                <h3 className="text-2xl font-bold tracking-tight capitalize">{report?.status || "Unknown"}</h3>
                <p className="text-muted-foreground flex items-center gap-2 text-sm">
                    <Clock className="h-3 w-3" />
                    Last checked: {report?.timestamp ? new Date(report.timestamp).toLocaleString() : "-"}
                </p>
             </div>
          </div>
          <Button onClick={fetchHealth} disabled={loading} size="lg">
            {loading ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : <RefreshCw className="mr-2 h-4 w-4" />}
            Run Check
          </Button>
        </CardContent>
      </Card>

      {/* Checks Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
         {Object.entries(report?.checks || {}).map(([name, result]) => (
            <Card key={name} className="flex flex-col overflow-hidden transition-all hover:shadow-md">
                <CardHeader className="p-4 pb-2 flex flex-row items-center justify-between space-y-0">
                    <CardTitle className="text-base font-medium flex items-center gap-2">
                        {getIconForCheck(name)}
                        {name}
                    </CardTitle>
                    {getStatusBadge(result.status)}
                </CardHeader>
                <CardContent className="p-4 pt-2 flex-1 flex flex-col justify-between">
                    <div className="text-sm text-muted-foreground mb-4">
                        {result.message || "No status message available."}
                    </div>
                    <div className="flex items-center justify-between pt-2 border-t text-xs text-muted-foreground">
                        <span>Latency</span>
                        <span className="font-mono">{result.latency || "< 1ms"}</span>
                    </div>
                     {result.diff && (
                        <div className="mt-2 p-2 bg-muted/50 rounded text-xs font-mono break-all text-red-500">
                             {result.diff}
                        </div>
                    )}
                </CardContent>
            </Card>
         ))}
      </div>

      {/* Service Health History Timeline */}
      {serviceHealth && (
          <Card className="shadow-sm">
              <CardHeader className="p-6 pb-2">
                  <CardTitle className="text-lg font-medium flex items-center gap-2">
                      <History className="h-5 w-5" />
                      Service Health History
                  </CardTitle>
              </CardHeader>
              <CardContent className="p-6 pt-4">
                  <div className="space-y-4">
                      {serviceHealth.services.map((svc) => (
                          <div key={svc.id} className="flex flex-col sm:flex-row sm:items-center gap-4 border-b last:border-0 pb-4 last:pb-0">
                              <div className="w-32 shrink-0">
                                  <div className="font-medium truncate" title={svc.name}>{svc.name}</div>
                                  <div className="text-xs text-muted-foreground flex items-center gap-1">
                                      <div className={cn("w-2 h-2 rounded-full", getStatusColor(svc.status))} />
                                      {svc.status}
                                  </div>
                              </div>
                              <div className="flex-1 overflow-hidden">
                                  {/* Render heatmap blocks */}
                                  <div className="flex gap-1 h-6">
                                      <TooltipProvider>
                                          {(serviceHealth.history[svc.id] || []).slice(-50).map((point, i) => (
                                              <Tooltip key={i}>
                                                  <TooltipTrigger asChild>
                                                      <div className={cn(
                                                          "w-2 h-full rounded-sm transition-opacity hover:opacity-80 cursor-help",
                                                          getStatusColor(point.status)
                                                      )} />
                                                  </TooltipTrigger>
                                                  <TooltipContent>
                                                      <div className="text-xs font-mono">
                                                          <div>Status: {point.status}</div>
                                                          <div>Time: {new Date(point.timestamp).toLocaleTimeString()}</div>
                                                      </div>
                                                  </TooltipContent>
                                              </Tooltip>
                                          ))}
                                          {(!serviceHealth.history[svc.id] || serviceHealth.history[svc.id].length === 0) && (
                                              <div className="text-xs text-muted-foreground italic self-center">No history available</div>
                                          )}
                                      </TooltipProvider>
                                  </div>
                              </div>
                          </div>
                      ))}
                  </div>
              </CardContent>
          </Card>
      )}
    </div>
  );
}
