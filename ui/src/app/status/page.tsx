/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useServiceHealthHistory, ServiceHealth } from "@/hooks/use-service-health-history";
import { StatusTimeline } from "@/components/status/status-timeline";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { CheckCircle2, AlertTriangle, XCircle, Activity, PauseCircle, RefreshCw } from "lucide-react";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";
import { useMemo } from "react";

const getStatusIcon = (status: string) => {
  switch (status) {
    case "healthy": return <CheckCircle2 className="h-5 w-5 text-green-500" />;
    case "degraded": return <AlertTriangle className="h-5 w-5 text-amber-500" />;
    case "unhealthy": return <XCircle className="h-5 w-5 text-red-500" />;
    case "inactive": return <PauseCircle className="h-5 w-5 text-muted-foreground" />;
    default: return <Activity className="h-5 w-5 text-muted-foreground" />;
  }
};

const getStatusLabel = (status: string) => {
    switch (status) {
        case "healthy": return "Operational";
        case "degraded": return "Degraded Performance";
        case "unhealthy": return "Service Outage";
        case "inactive": return "Inactive";
        default: return "Unknown";
    }
};

const getStatusColor = (status: string) => {
    switch (status) {
      case "healthy": return "text-green-600 dark:text-green-400";
      case "degraded": return "text-amber-600 dark:text-amber-400";
      case "unhealthy": return "text-red-600 dark:text-red-400";
      case "inactive": return "text-muted-foreground";
      default: return "text-muted-foreground";
    }
};

export default function StatusPage() {
    const { services, history, isLoading, refresh } = useServiceHealthHistory();

    const overallStatus = useMemo(() => {
        if (services.length === 0) return "inactive";
        if (services.some(s => s.status === "unhealthy")) return "unhealthy";
        if (services.some(s => s.status === "degraded")) return "degraded";
        if (services.every(s => s.status === "inactive")) return "inactive";
        return "healthy";
    }, [services]);

    const calculateUptime = (serviceId: string) => {
        const h = history[serviceId];
        if (!h || h.length === 0) return 100;
        const total = h.length;
        const unhealthy = h.filter(p => p.status === 'unhealthy').length;
        return ((total - unhealthy) / total) * 100;
    };

    if (isLoading) {
        return <div className="p-8 text-center text-muted-foreground">Loading system status...</div>;
    }

    return (
        <div className="flex-1 space-y-6 p-8 pt-6 pb-12">
             <div className="flex items-center justify-between">
                <h1 className="text-3xl font-bold tracking-tight">System Status</h1>
                <Button variant="outline" size="sm" onClick={() => refresh()}>
                    <RefreshCw className="mr-2 h-4 w-4" /> Refresh
                </Button>
            </div>

            {/* Overall Status Banner */}
            <div className={cn(
                "rounded-lg border p-6 flex items-center gap-4 shadow-sm",
                overallStatus === "healthy" ? "bg-green-50/50 border-green-200 dark:bg-green-900/10 dark:border-green-900/30" :
                overallStatus === "degraded" ? "bg-amber-50/50 border-amber-200 dark:bg-amber-900/10 dark:border-amber-900/30" :
                overallStatus === "unhealthy" ? "bg-red-50/50 border-red-200 dark:bg-red-900/10 dark:border-red-900/30" :
                "bg-muted/30"
            )}>
                 <div className={cn("p-3 rounded-full bg-background border shadow-sm", getStatusColor(overallStatus))}>
                    {getStatusIcon(overallStatus)}
                 </div>
                 <div>
                     <h2 className="text-xl font-semibold">{getStatusLabel(overallStatus)}</h2>
                     <p className="text-muted-foreground">
                         {overallStatus === "healthy" ? "All systems operational." :
                          overallStatus === "degraded" ? "Some services are experiencing performance issues." :
                          overallStatus === "unhealthy" ? "Major system outage detected." :
                          "System is currently inactive."}
                     </p>
                 </div>
            </div>

            <div className="grid gap-6">
                 {services.map(service => {
                     const uptime = calculateUptime(service.id);
                     return (
                        <Card key={service.id}>
                            <CardHeader className="pb-3">
                                <div className="flex items-center justify-between">
                                    <div className="flex items-center gap-3">
                                        {getStatusIcon(service.status)}
                                        <div>
                                            <CardTitle className="text-base">{service.name}</CardTitle>
                                            <CardDescription>
                                                {service.status === 'inactive' ? 'Service Disconnected' :
                                                 service.message || 'Operational'}
                                            </CardDescription>
                                        </div>
                                    </div>
                                    <div className="text-right">
                                        <div className="text-2xl font-bold font-mono">{uptime.toFixed(2)}%</div>
                                        <div className="text-xs text-muted-foreground uppercase tracking-wider">24h Uptime</div>
                                    </div>
                                </div>
                            </CardHeader>
                            <CardContent>
                                <div className="space-y-2">
                                    <div className="flex justify-between text-xs text-muted-foreground px-0.5">
                                        <span>24 hours ago</span>
                                        <span>Now</span>
                                    </div>
                                    <StatusTimeline history={history[service.id]} />
                                </div>
                            </CardContent>
                        </Card>
                     );
                 })}

                 {services.length === 0 && (
                     <div className="text-center py-12 border-2 border-dashed rounded-lg text-muted-foreground">
                         No services monitored.
                     </div>
                 )}
            </div>
        </div>
    );
}
