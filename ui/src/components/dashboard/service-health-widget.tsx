/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { memo, useMemo } from "react";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { CheckCircle2, AlertTriangle, XCircle, Activity, PauseCircle, Clock } from "lucide-react";
import { cn } from "@/lib/utils";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";
import { analyzeConnectionError } from "@/lib/diagnostics-utils";
import { useServiceHealthHistory, ServiceHealth, HealthHistoryPoint } from "@/hooks/use-service-health-history";
import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip";

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

// ⚡ Bolt Optimization: Extracted diagnosis logic to a memoized component.
const ServiceDiagnosisPopover = memo(function ServiceDiagnosisPopover({ message }: { message: string }) {
    const diagnosis = useMemo(() => analyzeConnectionError(message), [message]);

    return (
        <Popover>
          <PopoverTrigger>
             <AlertTriangle className="h-3 w-3 text-red-500 cursor-help hover:text-red-600 transition-colors" />
          </PopoverTrigger>
          <PopoverContent className="w-80">
            <div className="space-y-2">
                <div className="flex items-center gap-2 border-b pb-2">
                    <AlertTriangle className="h-4 w-4 text-red-500" />
                    <h4 className="font-medium text-red-900 dark:text-red-200">{diagnosis.title}</h4>
                </div>
                <p className="text-sm text-muted-foreground">{diagnosis.description}</p>
                <div className="bg-muted/50 p-2 rounded text-xs font-mono break-all max-h-24 overflow-y-auto">
                    {message}
                </div>
                <div className="pt-2">
                    <p className="text-xs font-medium mb-1">Suggestion:</p>
                    <p className="text-xs text-muted-foreground whitespace-pre-wrap">{diagnosis.suggestion}</p>
                </div>
            </div>
          </PopoverContent>
        </Popover>
    );
});

const HealthTimeline = memo(function HealthTimeline({ history }: { history: HealthHistoryPoint[] }) {
  if (!history || history.length === 0) return null;

  return (
    <div className="flex items-center gap-[2px] h-3 ml-4">
      {history.map((point, i) => {
        let colorClass = "bg-muted";
        switch (point.status) {
          case "healthy": colorClass = "bg-green-500/80 hover:bg-green-500"; break;
          case "degraded": colorClass = "bg-amber-500/80 hover:bg-amber-500"; break;
          case "unhealthy": colorClass = "bg-red-500/80 hover:bg-red-500"; break;
          case "inactive": colorClass = "bg-slate-200 dark:bg-slate-700 opacity-50"; break;
        }

        return (
          <Tooltip key={point.timestamp} delayDuration={0}>
            <TooltipTrigger asChild>
                <div
                  className={cn("w-1.5 h-full rounded-[1px] transition-all cursor-crosshair", colorClass)}
                />
            </TooltipTrigger>
            <TooltipContent className="text-[10px] p-1 px-2">
              <div className="font-semibold capitalize flex items-center gap-1">
                  {point.status === 'healthy' && <CheckCircle2 className="h-3 w-3 text-green-500" />}
                  {point.status === 'unhealthy' && <XCircle className="h-3 w-3 text-red-500" />}
                  {point.status}
              </div>
              <div className="text-muted-foreground">{new Date(point.timestamp).toLocaleTimeString()}</div>
            </TooltipContent>
          </Tooltip>
        );
      })}
    </div>
  );
});


// ⚡ Bolt Optimization: Memoized individual service items.
const ServiceHealthItem = memo(function ServiceHealthItem({ service, history }: { service: ServiceHealth, history: HealthHistoryPoint[] }) {
    return (
        <div
            className="group flex items-center justify-between p-3 hover:bg-muted/50 rounded-lg transition-colors"
        >
            <div className="flex items-center space-x-4 flex-1 min-w-0">
                <div className={cn("p-2 rounded-full bg-background shadow-sm border shrink-0", getStatusColor(service.status).split(" ")[0])}>
                    {getStatusIcon(service.status)}
                </div>
                <div className="min-w-0 flex-1">
                    <div className="flex items-center gap-2">
                        <p className="text-sm font-medium leading-none mb-1 truncate">{service.name}</p>
                        {service.message && (
                            <ServiceDiagnosisPopover message={service.message} />
                        )}
                    </div>
                    {service.status !== 'inactive' && service.latency !== '--' && (
                        <p className="text-xs text-muted-foreground flex items-center">
                        <Activity className="h-3 w-3 mr-1" />
                        Latency: <span className="font-mono ml-1">{service.latency}</span>
                        </p>
                    )}
                </div>

                {/* Timeline Visualization */}
                <div className="hidden md:flex flex-1 justify-end px-4">
                     <HealthTimeline history={history} />
                </div>
            </div>

            <div className="flex items-center space-x-4 shrink-0">
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
    );
});

/**
 * ServiceHealthWidget component.
 * @returns The rendered component.
 */
export function ServiceHealthWidget() {
  const { services, history, isLoading } = useServiceHealthHistory();

  const sortedServices = useMemo(() => {
    return Array.isArray(services) ? [...services].sort((a, b) => {
         const score = (s: string) => s === 'unhealthy' ? 0 : s === 'degraded' ? 1 : s === 'healthy' ? 2 : 3;
         return score(a.status) - score(b.status);
    }) : [];
  }, [services]);

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

  if (sortedServices.length === 0) {
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
        <div className="flex items-center justify-between">
            <div>
                <CardTitle>System Health</CardTitle>
                <CardDescription>
                Live health checks for {sortedServices.length} connected services.
                </CardDescription>
            </div>
             <div className="flex items-center gap-1 text-[10px] text-muted-foreground bg-muted/50 px-2 py-1 rounded">
                <Clock className="h-3 w-3" />
                <span>History (10m)</span>
            </div>
        </div>
      </CardHeader>
      <CardContent>
        <div className="space-y-1">
          {sortedServices.map((service) => (
            <ServiceHealthItem
                key={service.id}
                service={service}
                history={history[service.id]}
            />
          ))}
        </div>
      </CardContent>
    </Card>
  );
}
