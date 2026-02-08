/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { UpstreamServiceConfig, apiClient, GetServiceStatusResponse } from "@/lib/client";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Activity, Clock, Server } from "lucide-react";
import { useState, useEffect, useCallback } from "react";
import { Sparkline } from "@/components/charts/sparkline";
import { cn } from "@/lib/utils";

interface ServiceOverviewProps {
    service: UpstreamServiceConfig;
}

export function ServiceOverview({ service }: ServiceOverviewProps) {
    const [status, setStatus] = useState<GetServiceStatusResponse | null>(null);
    const [history, setHistory] = useState<number[]>([]);

    const fetchStatus = useCallback(async () => {
        try {
            const res = await apiClient.getServiceStatus(service.name);
            setStatus(res);
            // Simulate history update if needed or use real history if available
            // Assuming metrics contains current values, we might need a separate history endpoint or context
            // For now, we'll just use the current latency if available as a single point or keep a local history
            if (res.metrics && res.metrics['latency_ms']) {
                setHistory(prev => [...prev.slice(-19), res.metrics['latency_ms']]);
            }
        } catch (e) {
            console.error("Failed to fetch status", e);
        }
    }, [service.name]);

    useEffect(() => {
        fetchStatus();
        const interval = setInterval(fetchStatus, 5000);
        return () => clearInterval(interval);
    }, [fetchStatus]);

    const getType = () => {
        if (service.httpService) return 'HTTP';
        if (service.grpcService) return 'gRPC';
        if (service.commandLineService) return 'CLI';
        if (service.mcpService) return 'MCP';
        if (service.openapiService) return 'OpenAPI';
        return 'Unknown';
    };

    const getAddress = () => {
         return service.grpcService?.address ||
            service.httpService?.address ||
            service.commandLineService?.command ||
            service.mcpService?.httpConnection?.httpAddress ||
            service.mcpService?.stdioConnection?.command ||
            "-";
    };

    const isHealthy = !service.lastError && (!status || (status.metrics && status.metrics['uptime'] > 0));

    return (
        <div className="space-y-6">
            <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
                <Card>
                    <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                        <CardTitle className="text-sm font-medium">Status</CardTitle>
                        <Activity className={cn("h-4 w-4", isHealthy ? "text-green-500" : "text-red-500")} />
                    </CardHeader>
                    <CardContent>
                        <div className="text-2xl font-bold flex items-center gap-2">
                            {isHealthy ? "Healthy" : "Unhealthy"}
                        </div>
                        <p className="text-xs text-muted-foreground">
                            {service.lastError || "Service is running normally"}
                        </p>
                    </CardContent>
                </Card>
                <Card>
                    <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                        <CardTitle className="text-sm font-medium">Latency</CardTitle>
                        <Clock className="h-4 w-4 text-muted-foreground" />
                    </CardHeader>
                    <CardContent>
                        <div className="text-2xl font-bold">
                            {status?.metrics && status.metrics['latency_ms'] ? `${status.metrics['latency_ms']}ms` : "-"}
                        </div>
                        <div className="h-[20px] mt-2">
                             <Sparkline data={history} width={120} height={20} color={isHealthy ? "#22c55e" : "#ef4444"} />
                        </div>
                    </CardContent>
                </Card>
                <Card>
                    <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                        <CardTitle className="text-sm font-medium">Requests</CardTitle>
                        <Server className="h-4 w-4 text-muted-foreground" />
                    </CardHeader>
                    <CardContent>
                        <div className="text-2xl font-bold">
                             {(status?.metrics && status.metrics['requests']) || 0}
                        </div>
                        <p className="text-xs text-muted-foreground">
                            Total requests processed
                        </p>
                    </CardContent>
                </Card>
                <Card>
                    <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                        <CardTitle className="text-sm font-medium">Uptime</CardTitle>
                        <Clock className="h-4 w-4 text-muted-foreground" />
                    </CardHeader>
                    <CardContent>
                        <div className="text-2xl font-bold">
                            {status?.metrics && status.metrics['uptime'] ? `${Math.floor(status.metrics['uptime'] / 60)}m` : "-"}
                        </div>
                        <p className="text-xs text-muted-foreground">
                            Since last restart
                        </p>
                    </CardContent>
                </Card>
            </div>

            <Card>
                <CardHeader>
                    <CardTitle>Configuration Details</CardTitle>
                </CardHeader>
                <CardContent className="space-y-4">
                     <div className="grid grid-cols-2 gap-4">
                        <div>
                            <span className="text-sm font-medium text-muted-foreground">Service ID</span>
                            <div className="font-mono text-sm">{service.id}</div>
                        </div>
                         <div>
                            <span className="text-sm font-medium text-muted-foreground">Type</span>
                            <div><Badge variant="outline">{getType()}</Badge></div>
                        </div>
                         <div>
                            <span className="text-sm font-medium text-muted-foreground">Endpoint / Command</span>
                            <div className="font-mono text-sm truncate" title={getAddress()}>{getAddress()}</div>
                        </div>
                         <div>
                            <span className="text-sm font-medium text-muted-foreground">Version</span>
                            <div className="text-sm">{service.version}</div>
                        </div>
                     </div>
                </CardContent>
            </Card>
        </div>
    );
}
