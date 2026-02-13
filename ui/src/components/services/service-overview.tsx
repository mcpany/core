/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { UpstreamServiceConfig, apiClient } from "@/lib/client";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Activity, Server, Clock, Globe, RotateCw, Terminal, ArrowRight } from "lucide-react";
import { useMemo } from "react";
import { useServiceHealth } from "@/contexts/service-health-context";
import { Button } from "@/components/ui/button";
import { useToast } from "@/hooks/use-toast";
import { useRouter } from "next/navigation";
import { Area, AreaChart, ResponsiveContainer, Tooltip, Area as RechartsArea } from "recharts";
import { LogStream } from "@/components/logs/log-stream";

interface ServiceOverviewProps {
    service: UpstreamServiceConfig;
    status: any;
    trafficData?: any[];
}

/**
 * ServiceOverview displays a high-level summary of the service's health and metrics.
 * It connects to the real-time ServiceHealthContext for live updates.
 */
export function ServiceOverview({ service, status, trafficData = [] }: ServiceOverviewProps) {
    const { getServiceCurrentHealth, getServiceHistory } = useServiceHealth();
    const router = useRouter();
    const { toast } = useToast();

    // Get real-time data
    const currentHealth = getServiceCurrentHealth(service.name) || (service.id ? getServiceCurrentHealth(service.id) : null);
    const historyFromName = getServiceHistory(service.name);
    const historyFromId = service.id ? getServiceHistory(service.id) : [];
    const healthHistory = historyFromName.length > 0 ? historyFromName : historyFromId;

    // Derived state
    const isLive = !!currentHealth;

    // Fallback to props if no live data
    const metrics = currentHealth ? {
        latency: currentHealth.latencyMs,
        rps: currentHealth.qps,
        error_rate: currentHealth.errorRate * 100,
        status: currentHealth.status === 'NODE_STATUS_ACTIVE' ? "OK" : "ERROR"
    } : (status?.metrics || {});

    const isHealthy = !service.lastError && (
        currentHealth ? currentHealth.status === 'NODE_STATUS_ACTIVE' : (!status?.status || status.status === "OK")
    );

    const chartData = useMemo(() => {
        if (healthHistory.length > 0) {
            return healthHistory.map(h => ({
                time: new Date(h.timestamp).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', second: '2-digit' }),
                latency: h.latencyMs,
                rps: h.qps,
                errorRate: h.errorRate * 100
            }));
        }
        if (trafficData.length > 0) {
             return trafficData.map(d => ({
                time: d.time || "-",
                latency: d.latency || 0,
                // Backend API returns 'total' for requests count and 'errorRate' (0-1)
                rps: d.total || d.requests || 0,
                errorRate: (d.errorRate || d.errors || 0) * 100
            }));
        }
        return [];
    }, [healthHistory, trafficData]);

    const displayLatency = metrics.latency !== undefined ? `${Math.round(metrics.latency)}ms` : "N/A";
    const displayRPS = metrics.rps !== undefined ? metrics.rps.toFixed(1) : "N/A";
    const displayErrorRate = metrics.error_rate !== undefined ? `${metrics.error_rate.toFixed(2)}%` : "N/A";

    // Quick Actions
    const handleRestart = async () => {
        try {
            await apiClient.restartService(service.name);
            toast({ title: "Restart Initiated", description: `Service ${service.name} is restarting...` });
        } catch (e) {
            toast({ variant: "destructive", title: "Restart Failed", description: String(e) });
        }
    };

    const renderChart = (dataKey: string, color: string, fillId: string) => (
        <div className="h-[60px] w-full mt-2">
            <ResponsiveContainer width="100%" height="100%">
                <AreaChart data={chartData}>
                    <defs>
                        <linearGradient id={fillId} x1="0" y1="0" x2="0" y2="1">
                            <stop offset="5%" stopColor={color} stopOpacity={0.3} />
                            <stop offset="95%" stopColor={color} stopOpacity={0} />
                        </linearGradient>
                    </defs>
                    <Tooltip
                        contentStyle={{ borderRadius: '8px', border: 'none', boxShadow: '0 4px 6px -1px rgb(0 0 0 / 0.1)', fontSize: '12px', padding: '4px 8px' }}
                        cursor={{ stroke: 'rgba(255,255,255,0.1)' }}
                        labelStyle={{ display: 'none' }}
                        formatter={(value: any) => [typeof value === 'number' ? value.toFixed(2) : value, '']}
                    />
                    <RechartsArea
                        type="monotone"
                        dataKey={dataKey}
                        stroke={color}
                        strokeWidth={2}
                        fillOpacity={1}
                        fill={`url(#${fillId})`}
                        isAnimationActive={false} // Performance optimization for rapid updates
                    />
                </AreaChart>
            </ResponsiveContainer>
        </div>
    );

    return (
        <div className="space-y-6">
            {/* Status Card & Actions */}
            <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
                <Card className="col-span-full bg-muted/20 border-none shadow-none">
                    <CardContent className="p-0 flex items-center justify-between">
                        <div className="flex items-center gap-4 p-4">
                            <div className="relative">
                                <div className={`h-4 w-4 rounded-full ${isHealthy ? "bg-green-500" : "bg-red-500"} ${isLive ? "animate-pulse" : ""}`} />
                                {isLive && <div className={`absolute -inset-1 rounded-full ${isHealthy ? "bg-green-500" : "bg-red-500"} opacity-20 animate-ping`} />}
                            </div>
                            <div>
                                <h3 className="text-lg font-semibold flex items-center gap-2">
                                    {isHealthy ? "Operational" : "Degraded"}
                                    {!isLive && <Badge variant="outline" className="text-[10px] h-4">Offline / No Data</Badge>}
                                </h3>
                                <p className="text-sm text-muted-foreground">
                                    {service.lastError || "All systems functional"}
                                </p>
                            </div>
                        </div>
                        <div className="flex gap-2 p-4">
                             <Button variant="outline" size="sm" onClick={handleRestart}>
                                <RotateCw className="mr-2 h-3 w-3" /> Restart
                            </Button>
                            <Button variant="outline" size="sm" onClick={() => router.push(`/logs?source=${service.name}`)}>
                                <Terminal className="mr-2 h-3 w-3" /> Logs
                            </Button>
                        </div>
                    </CardContent>
                </Card>

                {/* Metrics Cards */}
                <Card>
                    <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                        <CardTitle className="text-sm font-medium">Latency</CardTitle>
                        <Clock className="h-4 w-4 text-muted-foreground" />
                    </CardHeader>
                    <CardContent>
                        <div className="text-2xl font-bold">{displayLatency}</div>
                        {chartData.length > 0 ? renderChart("latency", "#3b82f6", "colorLatency") : <div className="h-[60px] text-xs text-muted-foreground pt-4">No data</div>}
                    </CardContent>
                </Card>

                <Card>
                    <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                        <CardTitle className="text-sm font-medium">Throughput</CardTitle>
                        <Globe className="h-4 w-4 text-muted-foreground" />
                    </CardHeader>
                    <CardContent>
                        <div className="text-2xl font-bold">{displayRPS} <span className="text-xs font-normal text-muted-foreground">req/s</span></div>
                        {chartData.length > 0 ? renderChart("rps", "#8b5cf6", "colorRPS") : <div className="h-[60px] text-xs text-muted-foreground pt-4">No data</div>}
                    </CardContent>
                </Card>

                <Card>
                    <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                        <CardTitle className="text-sm font-medium">Error Rate</CardTitle>
                        <Server className="h-4 w-4 text-muted-foreground" />
                    </CardHeader>
                    <CardContent>
                        <div className="text-2xl font-bold">{displayErrorRate}</div>
                        {chartData.length > 0 ? renderChart("errorRate", "#ef4444", "colorError") : <div className="h-[60px] text-xs text-muted-foreground pt-4">No data</div>}
                    </CardContent>
                </Card>

                <Card>
                     <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                        <CardTitle className="text-sm font-medium">Uptime</CardTitle>
                        <Activity className="h-4 w-4 text-muted-foreground" />
                    </CardHeader>
                    <CardContent>
                        {/* Mock uptime calculation for now, or derive from history gaps */}
                        <div className="text-2xl font-bold">99.9%</div>
                        <div className="text-xs text-muted-foreground mt-2">Last 30 days</div>
                        <div className="h-[44px]"></div>
                    </CardContent>
                </Card>
            </div>

            {/* Logs Preview */}
            <Card className="col-span-full h-[400px] flex flex-col overflow-hidden">
                <CardHeader className="flex flex-row items-center justify-between pb-2 bg-muted/10">
                    <CardTitle className="text-lg font-medium flex items-center gap-2">
                        <Terminal className="h-5 w-5" /> Recent Logs
                    </CardTitle>
                    <Button variant="ghost" size="sm" onClick={() => router.push(`/logs?source=${service.name}`)}>
                        View Full Logs <ArrowRight className="ml-2 h-4 w-4" />
                    </Button>
                </CardHeader>
                <CardContent className="flex-1 p-0 min-h-0 relative">
                     <LogStream source={service.name} minimal />
                </CardContent>
            </Card>

            <Card className="col-span-full">
                <CardHeader>
                    <CardTitle>Configuration</CardTitle>
                </CardHeader>
                <CardContent>
                    <dl className="grid grid-cols-1 md:grid-cols-3 gap-4 text-sm">
                        <div>
                            <dt className="font-medium text-muted-foreground">Service Name</dt>
                            <dd>{service.name}</dd>
                        </div>
                        <div>
                            <dt className="font-medium text-muted-foreground">ID</dt>
                            <dd className="font-mono">{service.id}</dd>
                        </div>
                        <div>
                            <dt className="font-medium text-muted-foreground">Version</dt>
                            <dd>{service.version || "latest"}</dd>
                        </div>
                         <div>
                            <dt className="font-medium text-muted-foreground">Type</dt>
                            <dd>
                                <Badge variant="outline">
                                    {service.httpService ? "HTTP" :
                                     service.grpcService ? "gRPC" :
                                     service.commandLineService ? "CLI" :
                                     service.mcpService ? "MCP" : "Unknown"}
                                </Badge>
                            </dd>
                        </div>
                        <div className="col-span-2">
                            <dt className="font-medium text-muted-foreground">Address / Command</dt>
                            <dd className="font-mono bg-muted/50 p-1 rounded mt-1 truncate">
                                {service.httpService?.address ||
                                 service.grpcService?.address ||
                                 service.commandLineService?.command ||
                                 service.mcpService?.httpConnection?.httpAddress || "-"}
                            </dd>
                        </div>
                    </dl>
                </CardContent>
            </Card>
        </div>
    );
}
