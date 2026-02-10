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

interface HealthPoint {
  timestamp: number;
  status: string;
}

interface ServiceHealth {
  id: string;
  name: string;
  status: string;
  latency: string;
  uptime: string;
  message?: string;
}

interface HealthResponse {
  services: ServiceHealth[];
  history: Record<string, HealthPoint[]>;
}

function HealthTimeline({ points }: { points: HealthPoint[] }) {
  // Show last 40 points to fit nicely
  const displayPoints = points ? points.slice(-40) : [];
  // Fill with gray if empty to show "no data" or fixed width?
  // Let's just show what we have.

  return (
    <div className="flex items-center gap-[2px] h-4 mt-2">
      {displayPoints.map((p, i) => {
        let color = "bg-muted";
        const s = p.status.toLowerCase();
        if (s === "up" || s === "healthy") color = "bg-green-500";
        else if (s === "down" || s === "unhealthy" || s === "error") color = "bg-red-500";
        else if (s === "degraded") color = "bg-yellow-500";

        return (
          <div
            key={i}
            className={`w-1.5 h-full rounded-[1px] ${color} transition-opacity hover:opacity-80`}
            title={`${new Date(p.timestamp).toLocaleTimeString()}: ${p.status}`}
          />
        );
      })}
      {displayPoints.length === 0 && <span className="text-xs text-muted-foreground">No history</span>}
    </div>
  )
}

/**
 * SystemHealth component.
 * @returns The rendered component.
 */
export function SystemHealth() {
  const [doctorReport, setDoctorReport] = useState<DoctorReport | null>(null);
  const [dashboardHealth, setDashboardHealth] = useState<HealthResponse | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const fetchData = async () => {
    setLoading(true);
    setError(null);
    try {
      const [doctor, health] = await Promise.all([
        apiClient.getDoctorStatus(),
        apiClient.getDashboardHealth()
      ]);
      setDoctorReport(doctor);
      setDashboardHealth(health);
    } catch (err) {
      console.error("Failed to fetch system health", err);
      // Fallback: if one fails, maybe we still show partial?
      // For now, show error.
      setError("Failed to retrieve system health data.");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchData();
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
      case "down":
        return <Badge variant="destructive">Unhealthy</Badge>;
      case "inactive":
        return <Badge variant="outline">Inactive</Badge>;
      default:
        return <Badge variant="outline">{status}</Badge>;
    }
  };

  if (loading && !doctorReport) {
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
        <Button onClick={fetchData} variant="outline">
          <RefreshCw className="mr-2 h-4 w-4" />
          Retry
        </Button>
      </div>
    );
  }

  return (
    <div className="space-y-8">
      {/* Top Status Card (System Doctor) */}
      <Card className="border-l-4 border-l-primary shadow-sm bg-gradient-to-r from-background to-muted/20">
        <CardContent className="p-6 flex items-center justify-between">
          <div className="flex items-center gap-4">
             <div className={cn("p-3 rounded-full bg-muted",
                 doctorReport?.status === 'healthy' ? "bg-green-100 dark:bg-green-900/30 text-green-600" :
                 doctorReport?.status === 'degraded' ? "bg-yellow-100 dark:bg-yellow-900/30 text-yellow-600" :
                 "bg-red-100 dark:bg-red-900/30 text-red-600"
             )}>
                 {doctorReport?.status === 'healthy' ? <CheckCircle2 className="h-8 w-8" /> :
                  doctorReport?.status === 'degraded' ? <AlertTriangle className="h-8 w-8" /> :
                  <XCircle className="h-8 w-8" />}
             </div>
             <div>
                <h3 className="text-2xl font-bold tracking-tight capitalize">System {doctorReport?.status || "Unknown"}</h3>
                <p className="text-muted-foreground flex items-center gap-2 text-sm">
                    <Clock className="h-3 w-3" />
                    Last checked: {doctorReport?.timestamp ? new Date(doctorReport.timestamp).toLocaleString() : "-"}
                </p>
             </div>
          </div>
          <Button onClick={fetchData} disabled={loading} size="lg">
            {loading ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : <RefreshCw className="mr-2 h-4 w-4" />}
            Refresh
          </Button>
        </CardContent>
      </Card>

      {/* Upstream Services Health (With Timeline) */}
      <div>
        <h3 className="text-lg font-semibold mb-4 flex items-center gap-2">
            <Activity className="h-5 w-5" /> Upstream Services
        </h3>
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            {dashboardHealth?.services.map((svc) => (
                <Card key={svc.id} className="flex flex-col overflow-hidden transition-all hover:shadow-md">
                    <CardHeader className="p-4 pb-2 flex flex-row items-center justify-between space-y-0">
                        <CardTitle className="text-base font-medium flex items-center gap-2 truncate">
                            <Server className="h-4 w-4 text-muted-foreground" />
                            {svc.name}
                        </CardTitle>
                        {getStatusBadge(svc.status)}
                    </CardHeader>
                    <CardContent className="p-4 pt-2 flex-1 flex flex-col justify-between">
                        <div className="text-sm text-muted-foreground mb-2">
                            {svc.message || (svc.status === 'healthy' ? "Operational" : "No status message")}
                        </div>

                        {/* Timeline Visualization */}
                        <div className="mt-auto pt-2">
                            <div className="flex items-center justify-between text-xs text-muted-foreground mb-1">
                                <span className="flex items-center gap-1"><History className="h-3 w-3"/> History</span>
                                <span className="font-mono">{svc.latency}</span>
                            </div>
                            <HealthTimeline points={dashboardHealth.history[svc.id] || []} />
                        </div>
                    </CardContent>
                </Card>
            ))}
            {(!dashboardHealth?.services || dashboardHealth.services.length === 0) && (
                <div className="col-span-full p-8 text-center text-muted-foreground bg-muted/20 rounded-lg border border-dashed">
                    No upstream services configured.
                </div>
            )}
        </div>
      </div>

      {/* Internal System Checks (Doctor) */}
      <div>
        <h3 className="text-lg font-semibold mb-4 flex items-center gap-2">
            <Cpu className="h-5 w-5" /> Internal System Checks
        </h3>
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
            {Object.entries(doctorReport?.checks || {}).map(([name, result]) => (
                <Card key={name} className="flex flex-col overflow-hidden bg-muted/10">
                    <CardContent className="p-4 flex flex-col gap-2">
                        <div className="flex items-center justify-between">
                            <span className="font-medium text-sm capitalize">{name}</span>
                            {getStatusBadge(result.status)}
                        </div>
                        {result.message && <div className="text-xs text-muted-foreground">{result.message}</div>}
                        {result.diff && (
                            <div className="mt-1 p-1 bg-background rounded text-[10px] font-mono break-all text-red-500 border">
                                {result.diff.substring(0, 100)}...
                            </div>
                        )}
                    </CardContent>
                </Card>
            ))}
        </div>
      </div>
    </div>
  );
}
