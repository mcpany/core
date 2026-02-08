/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { UpstreamServiceConfig } from "@/lib/client";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Sparkline } from "@/components/charts/sparkline";
import { Activity, Server, Clock, Globe } from "lucide-react";
import { useMemo } from "react";

interface ServiceOverviewProps {
    service: UpstreamServiceConfig;
    status: any;
    trafficData?: any[];
}

/**
 * ServiceOverview displays a high-level summary of the service's health and metrics.
 * It includes status indicators, sparkline charts for traffic history, and key configuration details.
 */
export function ServiceOverview({ service, status, trafficData = [] }: ServiceOverviewProps) {
    const isHealthy = !service.lastError && (!status?.status || status.status === "OK");
    const metrics = status?.metrics || {};

    // Process traffic data for sparklines
    const latencyData = useMemo(() => {
        if (!trafficData.length) return [];
        return trafficData.map(d => d.latency || 0);
    }, [trafficData]);

    const rpsData = useMemo(() => {
        if (!trafficData.length) return [];
        return trafficData.map(d => d.requests || 0);
    }, [trafficData]);

    const errorData = useMemo(() => {
        if (!trafficData.length) return [];
        return trafficData.map(d => d.errors || 0);
    }, [trafficData]);

    // Helpers to format display values
    const displayLatency = metrics.latency !== undefined ? `${metrics.latency}ms` : "N/A";
    const displayRPS = metrics.rps !== undefined ? metrics.rps : "N/A";
    const displayErrorRate = metrics.error_rate !== undefined ? `${metrics.error_rate}%` : "N/A";

    return (
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
            <Card>
                <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                    <CardTitle className="text-sm font-medium">Health Status</CardTitle>
                    <Activity className="h-4 w-4 text-muted-foreground" />
                </CardHeader>
                <CardContent>
                    <div className="flex items-center gap-2">
                        <div className={`h-3 w-3 rounded-full ${isHealthy ? "bg-green-500" : "bg-red-500"} animate-pulse`} />
                        <div className="text-2xl font-bold">{isHealthy ? "Healthy" : "Unhealthy"}</div>
                    </div>
                    <p className="text-xs text-muted-foreground mt-1">
                        {service.lastError || "Service is operating normally"}
                    </p>
                </CardContent>
            </Card>

            <Card>
                <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                    <CardTitle className="text-sm font-medium">Latency (ms)</CardTitle>
                    <Clock className="h-4 w-4 text-muted-foreground" />
                </CardHeader>
                <CardContent>
                    <div className="text-2xl font-bold">{displayLatency}</div>
                    <div className="h-[40px] mt-2">
                        {latencyData.length > 0 ? (
                            <Sparkline data={latencyData} width={120} height={40} color="#3b82f6" />
                        ) : (
                            <div className="text-xs text-muted-foreground pt-2">No history</div>
                        )}
                    </div>
                </CardContent>
            </Card>

            <Card>
                <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                    <CardTitle className="text-sm font-medium">Requests / sec</CardTitle>
                    <Globe className="h-4 w-4 text-muted-foreground" />
                </CardHeader>
                <CardContent>
                    <div className="text-2xl font-bold">{displayRPS}</div>
                    <div className="h-[40px] mt-2">
                         {rpsData.length > 0 ? (
                             <Sparkline data={rpsData} width={120} height={40} color="#8b5cf6" />
                         ) : (
                             <div className="text-xs text-muted-foreground pt-2">No history</div>
                         )}
                    </div>
                </CardContent>
            </Card>

            <Card>
                <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                    <CardTitle className="text-sm font-medium">Error Rate</CardTitle>
                    <Server className="h-4 w-4 text-muted-foreground" />
                </CardHeader>
                <CardContent>
                    <div className="text-2xl font-bold">{displayErrorRate}</div>
                    <div className="h-[40px] mt-2">
                         {errorData.length > 0 ? (
                             <Sparkline data={errorData} width={120} height={40} color="#ef4444" />
                         ) : (
                             <div className="text-xs text-muted-foreground pt-2">No history</div>
                         )}
                    </div>
                </CardContent>
            </Card>

            <Card className="col-span-full">
                <CardHeader>
                    <CardTitle>Service Information</CardTitle>
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
