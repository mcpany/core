/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useEffect, useState } from "react";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { CheckCircle2, AlertTriangle, XCircle, Activity, PauseCircle } from "lucide-react";
import { cn } from "@/lib/utils";
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from "@/components/ui/tooltip";

interface ServiceHealth {
  id: string;
  name: string;
  status: "healthy" | "degraded" | "unhealthy" | "inactive";
  latency: string;
  uptime: string;
  message?: string;
}

export function ServiceHealthWidget() {
  const [services, setServices] = useState<ServiceHealth[]>([]);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    async function fetchHealth() {
      try {
        const res = await fetch("/api/dashboard/health");
        if (res.ok) {
          const data = await res.json();
          // Sort services: unhealthy first, then active, then inactive
          const sorted = Array.isArray(data) ? data.sort((a: ServiceHealth, b: ServiceHealth) => {
             const score = (s: string) => s === 'unhealthy' ? 0 : s === 'degraded' ? 1 : s === 'healthy' ? 2 : 3;
             return score(a.status) - score(b.status);
          }) : [];
          setServices(sorted);
        }
      } catch (error) {
        console.error("Failed to fetch health data", error);
      } finally {
        setIsLoading(false);
      }
    }
    fetchHealth();
    const interval = setInterval(() => {
      // ⚡ Bolt Optimization: Pause polling when tab is not visible
      if (!document.hidden) {
        fetchHealth();
      }
    }, 10000); // Poll every 10s

    // ⚡ Bolt Optimization: Refresh immediately when tab becomes visible
    const onVisibilityChange = () => {
      if (!document.hidden) {
        fetchHealth();
      }
    };
    document.addEventListener("visibilitychange", onVisibilityChange);

    return () => {
      clearInterval(interval);
      document.removeEventListener("visibilitychange", onVisibilityChange);
    };
  }, []);

  const getStatusIcon = (status: string) => {
    switch (status) {
      case "healthy":
        return <CheckCircle2 className="h-4 w-4 text-green-500" />;
      case "degraded":
        return <AlertTriangle className="h-4 w-4 text-amber-500" />;
      case "unhealthy":
        return <XCircle className="h-4 w-4 text-red-500" />;
      case "inactive":
        return <PauseCircle className="h-4 w-4 text-muted-foreground" />;
      default:
        return <Activity className="h-4 w-4 text-muted-foreground" />;
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

  if (isLoading) {
    return (
        <Card className="col-span-4 backdrop-blur-xl bg-background/60 border border-white/20 shadow-sm">
             <CardHeader>
                <CardTitle>System Health</CardTitle>
             </CardHeader>
             <CardContent>
                 <div className="flex items-center justify-center h-48">
                     <p className="text-muted-foreground animate-pulse">Checking system status...</p>
                 </div>
             </CardContent>
        </Card>
    )
  }

  if (services.length === 0) {
      return (
          <Card className="col-span-4 backdrop-blur-xl bg-background/60 border border-white/20 shadow-sm">
             <CardHeader>
                <CardTitle>System Health</CardTitle>
                 <CardDescription>
                  No services connected.
                </CardDescription>
             </CardHeader>
             <CardContent>
                 <div className="flex flex-col items-center justify-center h-32 text-muted-foreground text-sm">
                     <p>Register a service to see health status.</p>
                 </div>
             </CardContent>
        </Card>
      )
  }

  return (
    <Card className="col-span-4 backdrop-blur-xl bg-background/60 border border-white/20 shadow-sm transition-all duration-300">
      <CardHeader>
        <CardTitle>System Health</CardTitle>
        <CardDescription>
          Live health checks for {services.length} connected services.
        </CardDescription>
      </CardHeader>
      <CardContent>
        <div className="space-y-1">
          {services.map((service) => (
            <div
              key={service.id}
              className="group flex items-center justify-between p-3 hover:bg-muted/50 rounded-lg transition-colors"
            >
              <div className="flex items-center space-x-4">
                <div className={cn("p-2 rounded-full bg-background shadow-sm border", getStatusColor(service.status).split(" ")[0])}>
                    {getStatusIcon(service.status)}
                </div>
                <div>
                  <div className="flex items-center gap-2">
                     <p className="text-sm font-medium leading-none mb-1">{service.name}</p>
                     {service.message && (
                         <TooltipProvider>
                             <Tooltip>
                                 <TooltipTrigger>
                                     <AlertTriangle className="h-3 w-3 text-red-500 cursor-help" />
                                 </TooltipTrigger>
                                 <TooltipContent>
                                     <p className="max-w-xs">{service.message}</p>
                                 </TooltipContent>
                             </Tooltip>
                         </TooltipProvider>
                     )}
                  </div>
                  {service.status !== 'inactive' && service.latency !== '--' && (
                       <p className="text-xs text-muted-foreground flex items-center">
                        <Activity className="h-3 w-3 mr-1" />
                        Latency: <span className="font-mono ml-1">{service.latency}</span>
                      </p>
                  )}
                </div>
              </div>
              <div className="flex items-center space-x-4">
                 {(service.status !== 'inactive' && service.uptime !== '--') && (
                     <div className="text-right hidden sm:block">
                         <p className="text-xs text-muted-foreground">Uptime</p>
                         <p className="text-sm font-medium">{service.uptime}</p>
                     </div>
                 )}
                 <Badge variant="outline" className={cn("capitalize shadow-none", getStatusColor(service.status))}>
                    {service.status}
                 </Badge>
              </div>
            </div>
          ))}
        </div>
      </CardContent>
    </Card>
  );
}
