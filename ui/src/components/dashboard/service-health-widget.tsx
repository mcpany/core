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
// ⚡ BOLT: Replaced legacy hook with optimized context
import { useServiceHealth, MetricPoint } from "@/contexts/service-health-context";
import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip";
import { NodeStatus } from "@/types/topology";

// Helper type for the view
interface ServiceHealthView {
    id: string;
    name: string;
    status: string;
    latency: string;
    uptime: string;
    message?: string;
}

const getStatusIcon = (status: string) => {
  switch (status) {
    case "NODE_STATUS_ACTIVE":
    case "healthy":
      return <CheckCircle2 className="h-4 w-4 text-green-500" />;
    case "NODE_STATUS_DEGRADED":
    case "degraded":
      return <AlertTriangle className="h-4 w-4 text-amber-500" />;
    case "NODE_STATUS_ERROR":
    case "unhealthy":
      return <XCircle className="h-4 w-4 text-red-500" />;
    case "NODE_STATUS_INACTIVE":
    case "inactive":
      return <PauseCircle className="h-4 w-4 text-muted-foreground" />;
    default:
      return <Activity className="h-4 w-4 text-muted-foreground" />;
  }
};

const mapStatus = (status: NodeStatus | string): string => {
    if (status === 'NODE_STATUS_ACTIVE') return 'healthy';
    if (status === 'NODE_STATUS_DEGRADED') return 'degraded';
    if (status === 'NODE_STATUS_ERROR') return 'unhealthy';
    if (status === 'NODE_STATUS_INACTIVE') return 'inactive';
    return 'unknown';
}

const getStatusColor = (status: string) => {
    // Map protobuf status to simple status if needed, but UI seems to mix them
    const s = mapStatus(status as NodeStatus) === 'unknown' ? status : mapStatus(status as NodeStatus);
    switch (s) {
      case "healthy": return "border-green-200 bg-green-50 text-green-700 dark:border-green-900/30 dark:bg-green-900/20 dark:text-green-400";
      case "degraded": return "border-amber-200 bg-amber-50 text-amber-700 dark:border-amber-900/30 dark:bg-amber-900/20 dark:text-amber-400";
      case "unhealthy": return "border-red-200 bg-red-50 text-red-700 dark:border-red-900/30 dark:bg-red-900/20 dark:text-red-400";
      case "inactive": return "border-muted bg-muted/50 text-muted-foreground dark:bg-muted/80 dark:text-muted-foreground/80";
      default: return "border-gray-200 bg-gray-50 text-gray-700 dark:border-gray-800 dark:bg-gray-800/50 dark:text-gray-400";
    }
};

const formatUptime = (ms: number) => {
    if (ms < 1000) return "0s";
    const sec = Math.floor(ms / 1000);
    const min = Math.floor(sec / 60);
    const hr = Math.floor(min / 60);
    const day = Math.floor(hr / 24);

    if (day > 0) return `${day}d ${hr % 24}h`;
    if (hr > 0) return `${hr}h ${min % 60}m`;
    if (min > 0) return `${min}m ${sec % 60}s`;
    return `${sec}s`;
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

const HealthTimeline = memo(function HealthTimeline({ history }: { history: MetricPoint[] }) {
  if (!history || history.length === 0) return null;

  return (
    <div className="flex items-center gap-[2px] h-3 ml-4">
      {history.map((point) => {
        let colorClass = "bg-muted";
        const simpleStatus = mapStatus(point.status);
        switch (simpleStatus) {
          case "healthy": colorClass = "bg-green-500/80 hover:bg-green-500"; break;
          case "degraded": colorClass = "bg-amber-500/80 hover:bg-amber-500"; break;
          case "unhealthy": colorClass = "bg-red-500/80 hover:bg-red-500"; break;
          case "inactive": colorClass = "bg-slate-300 dark:bg-slate-600"; break;
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
                  {simpleStatus === 'healthy' && <CheckCircle2 className="h-3 w-3 text-green-500" />}
                  {simpleStatus === 'unhealthy' && <XCircle className="h-3 w-3 text-red-500" />}
                  {simpleStatus}
              </div>
              <div className="text-muted-foreground">{new Date(point.timestamp).toLocaleTimeString()}</div>
              <div className="text-[9px] text-muted-foreground/80 mt-1">
                  Latency: {point.latencyMs.toFixed(0)}ms
              </div>
            </TooltipContent>
          </Tooltip>
        );
      })}
    </div>
  );
});


const ServiceHealthItem = memo(function ServiceHealthItem({ service, history }: { service: ServiceHealthView, history: MetricPoint[] }) {
    const status = mapStatus(service.status as NodeStatus); // normalizing
    return (
        <div
            className="group flex items-center justify-between p-3 hover:bg-muted/50 rounded-lg transition-colors"
        >
            <div className="flex items-center space-x-4 flex-1 min-w-0">
                <div className={cn("p-2 rounded-full bg-background shadow-sm border shrink-0", getStatusColor(status).split(" ")[0])}>
                    {getStatusIcon(service.status)}
                </div>
                <div className="min-w-0 flex-1">
                    <div className="flex items-center gap-2">
                        <p className="text-sm font-medium leading-none mb-1 truncate">{service.name}</p>
                        {service.message && (
                            <ServiceDiagnosisPopover message={service.message} />
                        )}
                    </div>
                    {status !== 'inactive' && service.latency !== '--' && (
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
                {(status !== 'inactive' && service.uptime !== '--') && (
                    <div className="text-right hidden sm:block">
                        <p className="text-xs text-muted-foreground">Uptime</p>
                        <p className="text-sm font-medium">{service.uptime}</p>
                    </div>
                )}
                <Badge variant="outline" className={cn("capitalize shadow-none", getStatusColor(status))}>
                {status}
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
  const { latestTopology, getServiceHistory, getServiceUptime } = useServiceHealth();

  const services: ServiceHealthView[] = useMemo(() => {
      if (!latestTopology || !latestTopology.core) return [];
      const map = new Map<string, ServiceHealthView>();

      const traverse = (nodes: any[]) => {
          nodes.forEach(node => {
              if (node.type === 'NODE_TYPE_SERVICE') {
                  const uptimeMs = getServiceUptime(node.id);
                  // Deduplicate by ID
                  if (!map.has(node.id)) {
                      map.set(node.id, {
                          id: node.id,
                          name: node.label,
                          status: node.status || 'NODE_STATUS_UNSPECIFIED',
                          latency: node.metrics?.latencyMs ? `${node.metrics.latencyMs.toFixed(0)}ms` : '--',
                          uptime: uptimeMs > 0 ? formatUptime(uptimeMs) : '--',
                          message: node.metrics?.errorRate > 0 ? `Error Rate: ${(node.metrics.errorRate * 100).toFixed(1)}%` : undefined
                      });
                  }
              }
              if (node.children) traverse(node.children);
          });
      }
      traverse(latestTopology.core ? [latestTopology.core] : []);
      if (latestTopology.core && latestTopology.core.children) traverse(latestTopology.core.children);

      // Sort by status priority
      return Array.from(map.values()).sort((a, b) => {
           const score = (s: string) => {
               const simple = mapStatus(s as NodeStatus);
               return simple === 'unhealthy' ? 0 : simple === 'degraded' ? 1 : simple === 'healthy' ? 2 : 3;
           }
           return score(a.status) - score(b.status);
      });
  }, [latestTopology, getServiceUptime]);

  const isLoading = !latestTopology;

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
        <div className="flex items-center justify-between">
            <div>
                <CardTitle>System Health</CardTitle>
                <CardDescription>
                Live health checks for {services.length} connected services.
                </CardDescription>
            </div>
             <div className="flex items-center gap-1 text-[10px] text-muted-foreground bg-muted/50 px-2 py-1 rounded">
                <Clock className="h-3 w-3" />
                <span>History (2.5m)</span>
            </div>
        </div>
      </CardHeader>
      <CardContent>
        <div className="space-y-1">
          {services.map((service) => (
            <ServiceHealthItem
                key={service.id}
                service={service}
                history={getServiceHistory(service.id)}
            />
          ))}
        </div>
      </CardContent>
    </Card>
  );
}
