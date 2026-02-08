/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useMemo } from "react";
import { UpstreamServiceConfig, GetServiceStatusResponse } from "@/lib/client";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Activity, Clock, AlertTriangle, CheckCircle, Server, Zap, Shield, Hash } from "lucide-react";
import { Sparkline } from "@/components/charts/sparkline";
import { useServiceHealth } from "@/contexts/service-health-context";

interface ServiceOverviewProps {
    service: UpstreamServiceConfig;
    status: GetServiceStatusResponse;
}

/**
 * ServiceOverview component.
 * Displays key metrics and health status for a service.
 * @param props - The component props.
 * @param props.service - The service configuration.
 * @param props.status - The service runtime status.
 * @returns The rendered component.
 */
export function ServiceOverview({ service, status }: ServiceOverviewProps) {
    const { getServiceHistory } = useServiceHealth();
    const history = getServiceHistory(service.name);

    const latencies = useMemo(() => history.map(h => h.latencyMs), [history]);
    const maxLatency = useMemo(() => Math.max(...latencies, 50), [latencies]);
    const currentLatency = latencies.length > 0 ? latencies[latencies.length - 1] : 0;

    const errorRates = useMemo(() => history.map(h => h.errorRate * 100), [history]); // Convert to percentage
    const maxError = useMemo(() => Math.max(...errorRates, 10), [errorRates]);
    const currentErrorRate = errorRates.length > 0 ? errorRates[errorRates.length - 1] : 0;

    const qpsHistory = useMemo(() => history.map(h => h.qps), [history]);
    const maxQps = useMemo(() => Math.max(...qpsHistory, 5), [qpsHistory]);
    const currentQps = qpsHistory.length > 0 ? qpsHistory[qpsHistory.length - 1] : 0;

    // Determine health status
    const isHealthy = !service.lastError && currentErrorRate < 10;
    const statusColor = isHealthy ? "text-green-500" : "text-destructive";
    const StatusIcon = isHealthy ? CheckCircle : AlertTriangle;

    // Calculate uptime (mocked or derived from metrics if available)
    const uptime = status.metrics?.uptime ? formatUptime(Number(status.metrics.uptime)) : "Unknown";

    return (
        <div className="space-y-6">
            {/* Health & Status Cards */}
            <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
                <Card>
                    <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                        <CardTitle className="text-sm font-medium">Health Status</CardTitle>
                        <Activity className={`h-4 w-4 ${statusColor}`} />
                    </CardHeader>
                    <CardContent>
                        <div className={`text-2xl font-bold flex items-center gap-2 ${statusColor}`}>
                            <StatusIcon className="h-6 w-6" />
                            {isHealthy ? "Healthy" : "Unhealthy"}
                        </div>
                        <p className="text-xs text-muted-foreground mt-1">
                            {service.lastError || "Service is operating normally."}
                        </p>
                    </CardContent>
                </Card>

                <Card>
                    <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                        <CardTitle className="text-sm font-medium">Latency</CardTitle>
                        <Zap className="h-4 w-4 text-muted-foreground" />
                    </CardHeader>
                    <CardContent>
                        <div className="text-2xl font-bold">{currentLatency.toFixed(0)}ms</div>
                        <div className="h-[40px] mt-2">
                            <Sparkline
                                data={latencies}
                                width={120}
                                height={40}
                                color={currentLatency > 500 ? "#eab308" : "#22c55e"}
                                max={maxLatency}
                            />
                        </div>
                    </CardContent>
                </Card>

                <Card>
                    <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                        <CardTitle className="text-sm font-medium">Requests / Sec</CardTitle>
                        <Activity className="h-4 w-4 text-muted-foreground" />
                    </CardHeader>
                    <CardContent>
                        <div className="text-2xl font-bold">{currentQps.toFixed(1)}</div>
                        <div className="h-[40px] mt-2">
                            <Sparkline
                                data={qpsHistory}
                                width={120}
                                height={40}
                                color="#3b82f6"
                                max={maxQps}
                            />
                        </div>
                    </CardContent>
                </Card>

                <Card>
                    <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                        <CardTitle className="text-sm font-medium">Error Rate</CardTitle>
                        <AlertTriangle className="h-4 w-4 text-muted-foreground" />
                    </CardHeader>
                    <CardContent>
                        <div className="text-2xl font-bold">{currentErrorRate.toFixed(1)}%</div>
                        <div className="h-[40px] mt-2">
                            <Sparkline
                                data={errorRates}
                                width={120}
                                height={40}
                                color="#ef4444"
                                max={maxError}
                            />
                        </div>
                    </CardContent>
                </Card>
            </div>

            {/* Service Details */}
            <Card>
                <CardHeader>
                    <CardTitle>Service Information</CardTitle>
                </CardHeader>
                <CardContent>
                    <dl className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6 text-sm">
                        <div>
                            <dt className="font-medium text-muted-foreground mb-1">Service ID</dt>
                            <dd className="font-mono bg-muted px-2 py-1 rounded inline-block">{service.id}</dd>
                        </div>
                        <div>
                            <dt className="font-medium text-muted-foreground mb-1">Base URL / Command</dt>
                            <dd className="font-mono truncate" title={getAddress(service)}>
                                {getAddress(service)}
                            </dd>
                        </div>
                        <div>
                            <dt className="font-medium text-muted-foreground mb-1">Version</dt>
                            <dd>{service.version || "latest"}</dd>
                        </div>
                        <div>
                            <dt className="font-medium text-muted-foreground mb-1">Uptime</dt>
                            <dd className="flex items-center gap-1">
                                <Clock className="h-3 w-3" /> {uptime}
                            </dd>
                        </div>
                        <div>
                            <dt className="font-medium text-muted-foreground mb-1">Security</dt>
                            <dd>
                                {isSecure(service) ? (
                                    <Badge variant="outline" className="text-green-600 border-green-200">
                                        <Shield className="h-3 w-3 mr-1" /> TLS Enabled
                                    </Badge>
                                ) : (
                                    <Badge variant="secondary">Insecure</Badge>
                                )}
                            </dd>
                        </div>
                        <div>
                            <dt className="font-medium text-muted-foreground mb-1">Tags</dt>
                            <dd className="flex gap-1 flex-wrap">
                                {service.tags?.map(tag => (
                                    <Badge key={tag} variant="secondary" className="text-xs">{tag}</Badge>
                                )) || "-"}
                            </dd>
                        </div>
                    </dl>
                </CardContent>
            </Card>
        </div>
    );
}

function getAddress(service: UpstreamServiceConfig): string {
    return service.grpcService?.address ||
        service.httpService?.address ||
        service.commandLineService?.command ||
        service.mcpService?.httpConnection?.httpAddress ||
        service.mcpService?.stdioConnection?.command ||
        "-";
}

function isSecure(service: UpstreamServiceConfig): boolean {
    return !!(service.grpcService?.tlsConfig || service.httpService?.tlsConfig || service.mcpService?.httpConnection?.tlsConfig);
}

function formatUptime(seconds: number): string {
    if (!seconds) return "0s";
    const d = Math.floor(seconds / (3600 * 24));
    const h = Math.floor((seconds % (3600 * 24)) / 3600);
    const m = Math.floor((seconds % 3600) / 60);
    const s = Math.floor(seconds % 60);

    const parts = [];
    if (d > 0) parts.push(`${d}d`);
    if (h > 0) parts.push(`${h}h`);
    if (m > 0) parts.push(`${m}m`);
    if (s > 0 || parts.length === 0) parts.push(`${s}s`);
    return parts.join(" ");
}
