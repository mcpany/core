/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useServiceHealthHistory, ServiceHealth } from "@/hooks/use-service-health-history";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { CheckCircle2, AlertTriangle, XCircle, Activity, PauseCircle } from "lucide-react";
import { StatusTimeline } from "@/components/status/status-timeline";
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

const getStatusColor = (status: string) => {
    switch (status) {
      case "healthy": return "border-green-200 bg-green-50 text-green-700 dark:border-green-900/30 dark:bg-green-900/20 dark:text-green-400";
      case "degraded": return "border-amber-200 bg-amber-50 text-amber-700 dark:border-amber-900/30 dark:bg-amber-900/20 dark:text-amber-400";
      case "unhealthy": return "border-red-200 bg-red-50 text-red-700 dark:border-red-900/30 dark:bg-red-900/20 dark:text-red-400";
      case "inactive": return "border-muted bg-muted/50 text-muted-foreground";
      default: return "border-gray-200 bg-gray-50 text-gray-700 dark:border-gray-800 dark:bg-gray-800/50 dark:text-gray-400";
    }
};

export default function StatusPage() {
    const { services, history, isLoading } = useServiceHealthHistory();

    const overallStatus = useMemo(() => {
        if (services.length === 0) return "No Services";
        if (services.some(s => s.status === 'unhealthy')) return "Critical Outage";
        if (services.some(s => s.status === 'degraded')) return "Degraded Performance";
        return "All Systems Operational";
    }, [services]);

    const overallColor = useMemo(() => {
        if (overallStatus === "Critical Outage") return "bg-red-500";
        if (overallStatus === "Degraded Performance") return "bg-amber-500";
        if (overallStatus === "All Systems Operational") return "bg-green-500";
        return "bg-muted";
    }, [overallStatus]);

    const calculateUptime = (serviceId: string) => {
        const h = history[serviceId];
        if (!h || h.length === 0) return 0;
        const up = h.filter(p => p.status === 'healthy').length;
        return ((up / h.length) * 100).toFixed(2);
    };

    return (
        <div className="flex-1 space-y-8 p-8 pt-6 pb-20">
             <div className="flex flex-col gap-2">
                <h1 className="text-3xl font-bold tracking-tight">System Status</h1>
                <p className="text-muted-foreground">Real-time availability and performance monitoring.</p>
             </div>

             {/* Status Banner */}
             <div className={cn("w-full p-6 rounded-lg text-white shadow-lg flex items-center justify-between", overallColor)}>
                 <div className="flex items-center gap-4">
                     <Activity className="h-8 w-8" />
                     <div>
                         <h2 className="text-2xl font-bold">{overallStatus}</h2>
                         <p className="opacity-90">Last updated: {new Date().toLocaleTimeString()}</p>
                     </div>
                 </div>
                 {/* Maybe some metrics summary here */}
             </div>

             <div className="grid gap-6">
                 {isLoading && (
                     <Card>
                         <CardContent className="p-8 text-center text-muted-foreground animate-pulse">
                             Loading status data...
                         </CardContent>
                     </Card>
                 )}

                 {!isLoading && services.length === 0 && (
                     <Card>
                         <CardContent className="p-8 text-center text-muted-foreground">
                             No services monitored.
                         </CardContent>
                     </Card>
                 )}

                 {services.map(service => {
                     const uptime = calculateUptime(service.id);
                     return (
                         <Card key={service.id} className="overflow-hidden">
                             <div className="p-6">
                                 <div className="flex items-start justify-between mb-6">
                                     <div className="flex items-center gap-4">
                                         <div className={cn("p-2 rounded-full border shadow-sm", getStatusColor(service.status).split(' ')[0])}>
                                             {getStatusIcon(service.status)}
                                         </div>
                                         <div>
                                             <div className="flex items-center gap-2">
                                                 <h3 className="text-lg font-bold">{service.name}</h3>
                                                 <Badge variant="outline" className={cn("capitalize", getStatusColor(service.status))}>
                                                     {service.status}
                                                 </Badge>
                                             </div>
                                             <p className="text-sm text-muted-foreground mt-1 font-mono text-[10px]">{service.id}</p>
                                         </div>
                                     </div>
                                     <div className="text-right">
                                         <p className="text-xs text-muted-foreground uppercase tracking-wider font-semibold">Observed Uptime</p>
                                         <p className={cn("text-2xl font-bold", Number(uptime) > 99 ? "text-green-600" : Number(uptime) > 95 ? "text-amber-600" : "text-red-600")}>
                                             {uptime}%
                                         </p>
                                     </div>
                                 </div>

                                 <div className="space-y-2">
                                     <div className="flex justify-between text-xs text-muted-foreground px-1">
                                         <span>24 hours ago</span>
                                         <span>Now</span>
                                     </div>
                                     <StatusTimeline history={history[service.id]} />
                                 </div>
                             </div>
                         </Card>
                     )
                 })}
             </div>
        </div>
    );
}
