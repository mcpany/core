/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import React, { useState, useEffect } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Badge } from "@/components/ui/badge";
import { apiClient, Metric, ServiceHealthResponse, ToolAnalytics, ToolFailureStats } from "@/lib/client";
import { Loader2, Activity, Server, Zap, AlertCircle, Clock, CheckCircle2, XCircle, FileText, MessageSquare, Hash } from "lucide-react";
import { ServiceHealthSparkline } from "@/components/services/service-health-sparkline";
import { cn } from "@/lib/utils";

// Map icon names to lucide components
const IconMap: Record<string, React.ElementType> = {
    Activity: Activity,
    Server: Server,
    Zap: Zap,
    AlertCircle: AlertCircle,
    Clock: Clock,
    FileText: FileText,
    MessageSquare: MessageSquare,
    Hash: Hash,
};

export function DashboardLayout() {
    const [metrics, setMetrics] = useState<Metric[]>([]);
    const [health, setHealth] = useState<ServiceHealthResponse | null>(null);
    const [topTools, setTopTools] = useState<ToolAnalytics[]>([]);
    const [toolFailures, setToolFailures] = useState<ToolFailureStats[]>([]);
    const [loading, setLoading] = useState(true);

    useEffect(() => {
        const loadData = async () => {
            try {
                const [metricsData, healthData, topToolsData, toolFailuresData] = await Promise.all([
                    apiClient.getDashboardMetrics(),
                    apiClient.getDashboardHealth(),
                    apiClient.getTopTools(),
                    apiClient.getToolFailures()
                ]);
                setMetrics(metricsData || []);
                setHealth(healthData);
                // getTopTools returns any, map it properly if needed. Usually returns ToolAnalytics[] or similar
                setTopTools((topToolsData as any) || []);
                setToolFailures(toolFailuresData || []);
            } catch (err) {
                console.error("Failed to load dashboard data", err);
            } finally {
                setLoading(false);
            }
        };

        loadData();

        // Refresh every 30s
        const interval = setInterval(loadData, 30000);
        return () => clearInterval(interval);
    }, []);

    if (loading) {
        return (
            <div className="flex h-64 items-center justify-center">
                <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
            </div>
        );
    }

    return (
        <div className="space-y-6 animate-in fade-in duration-500">
            {/* Top Row: Metric Cards */}
            <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
                {metrics.slice(0, 8).map((metric, i) => {
                    const Icon = IconMap[metric.icon] || Activity;
                    return (
                        <Card key={i} className="backdrop-blur-sm bg-background/50 border-muted/50 hover:bg-muted/10 transition-colors shadow-sm">
                            <CardHeader className="flex flex-row items-center justify-between pb-2 space-y-0">
                                <CardTitle className="text-sm font-medium text-muted-foreground">
                                    {metric.label}
                                </CardTitle>
                                <Icon className="h-4 w-4 text-muted-foreground opacity-70" />
                            </CardHeader>
                            <CardContent>
                                <div className="text-2xl font-bold">{metric.value}</div>
                                {(metric.change && metric.change !== "--") && (
                                    <p className="text-xs mt-1 flex items-center gap-1">
                                        <span className={cn(
                                            "font-medium",
                                            metric.trend === "up" ? "text-red-500" : metric.trend === "down" ? "text-green-500" : "text-muted-foreground"
                                        )}>
                                            {metric.change}
                                        </span>
                                        <span className="text-muted-foreground">vs last period</span>
                                    </p>
                                )}
                            </CardContent>
                        </Card>
                    );
                })}
            </div>

            {/* Middle Row: Service Health */}
            <Card className="backdrop-blur-sm bg-background/50 shadow-sm border-muted/50">
                <CardHeader>
                    <CardTitle className="flex items-center gap-2">
                        <Server className="h-5 w-5 text-primary" />
                        Service Health
                    </CardTitle>
                </CardHeader>
                <CardContent>
                    <Table>
                        <TableHeader>
                            <TableRow className="hover:bg-transparent border-b-muted/50">
                                <TableHead>Service</TableHead>
                                <TableHead>Status</TableHead>
                                <TableHead>Uptime</TableHead>
                                <TableHead>Latency</TableHead>
                                <TableHead className="w-[100px]">Activity</TableHead>
                            </TableRow>
                        </TableHeader>
                        <TableBody>
                            {(!health || !health.services || health.services.length === 0) ? (
                                <TableRow>
                                    <TableCell colSpan={5} className="text-center text-muted-foreground h-24">
                                        No services configured.
                                    </TableCell>
                                </TableRow>
                            ) : (
                                health.services.map((svc) => (
                                    <TableRow key={svc.id} className="group hover:bg-muted/20 border-b-muted/20">
                                        <TableCell className="font-medium">
                                            {svc.name}
                                            {svc.message && (
                                                <div className="text-[10px] text-destructive truncate max-w-[200px]" title={svc.message}>
                                                    {svc.message}
                                                </div>
                                            )}
                                        </TableCell>
                                        <TableCell>
                                            <Badge variant={
                                                svc.status === 'healthy' ? 'default' :
                                                svc.status === 'degraded' ? 'secondary' :
                                                svc.status === 'inactive' ? 'outline' : 'destructive'
                                            } className={cn(
                                                "capitalize text-[10px]",
                                                svc.status === 'healthy' && "bg-green-500/15 text-green-700 dark:text-green-400 hover:bg-green-500/25",
                                                svc.status === 'degraded' && "bg-yellow-500/15 text-yellow-700 dark:text-yellow-400 hover:bg-yellow-500/25"
                                            )}>
                                                {svc.status === 'healthy' && <CheckCircle2 className="mr-1 h-3 w-3" />}
                                                {svc.status === 'unhealthy' && <XCircle className="mr-1 h-3 w-3" />}
                                                {svc.status}
                                            </Badge>
                                        </TableCell>
                                        <TableCell className="text-sm font-mono text-muted-foreground">
                                            {svc.uptime}
                                        </TableCell>
                                        <TableCell className="text-sm font-mono text-muted-foreground">
                                            {svc.latency}
                                        </TableCell>
                                        <TableCell>
                                            <div className="h-6 w-full opacity-70 group-hover:opacity-100 transition-opacity">
                                                <ServiceHealthSparkline serviceName={svc.name} disabled={svc.status === 'inactive'} />
                                            </div>
                                        </TableCell>
                                    </TableRow>
                                ))
                            )}
                        </TableBody>
                    </Table>
                </CardContent>
            </Card>

            {/* Bottom Row: Top Tools & Failures */}
            <div className="grid gap-6 md:grid-cols-2">
                <Card className="backdrop-blur-sm bg-background/50 shadow-sm border-muted/50">
                    <CardHeader>
                        <CardTitle className="flex items-center gap-2">
                            <Zap className="h-5 w-5 text-amber-500" />
                            Top Tools
                        </CardTitle>
                    </CardHeader>
                    <CardContent>
                        <Table>
                            <TableHeader>
                                <TableRow className="hover:bg-transparent border-b-muted/50">
                                    <TableHead>Tool</TableHead>
                                    <TableHead>Service</TableHead>
                                    <TableHead className="text-right">Calls</TableHead>
                                </TableRow>
                            </TableHeader>
                            <TableBody>
                                {topTools.length === 0 ? (
                                    <TableRow>
                                        <TableCell colSpan={3} className="text-center text-muted-foreground h-24">
                                            No tool usage data.
                                        </TableCell>
                                    </TableRow>
                                ) : (
                                    topTools.slice(0, 5).map((tool, i) => (
                                        <TableRow key={i} className="hover:bg-muted/20 border-b-muted/20">
                                            <TableCell className="font-medium text-sm">{tool.name || (tool as any).toolName}</TableCell>
                                            <TableCell className="text-xs text-muted-foreground">{tool.serviceId || (tool as any).service}</TableCell>
                                            <TableCell className="text-right font-mono text-sm">{(tool.totalCalls || (tool as any).count || 0).toLocaleString()}</TableCell>
                                        </TableRow>
                                    ))
                                )}
                            </TableBody>
                        </Table>
                    </CardContent>
                </Card>

                <Card className="backdrop-blur-sm bg-background/50 shadow-sm border-muted/50">
                    <CardHeader>
                        <CardTitle className="flex items-center gap-2 text-destructive">
                            <AlertCircle className="h-5 w-5" />
                            Recent Failures
                        </CardTitle>
                    </CardHeader>
                    <CardContent>
                        <Table>
                            <TableHeader>
                                <TableRow className="hover:bg-transparent border-b-muted/50">
                                    <TableHead>Tool</TableHead>
                                    <TableHead>Service</TableHead>
                                    <TableHead className="text-right">Failure Rate</TableHead>
                                </TableRow>
                            </TableHeader>
                            <TableBody>
                                {toolFailures.length === 0 ? (
                                    <TableRow>
                                        <TableCell colSpan={3} className="text-center text-muted-foreground h-24">
                                            No recent failures.
                                        </TableCell>
                                    </TableRow>
                                ) : (
                                    toolFailures.slice(0, 5).map((failure, i) => (
                                        <TableRow key={i} className="hover:bg-muted/20 border-b-muted/20">
                                            <TableCell className="font-medium text-sm">{failure.name}</TableCell>
                                            <TableCell className="text-xs text-muted-foreground">{failure.serviceId}</TableCell>
                                            <TableCell className="text-right">
                                                <Badge variant="destructive" className="font-mono text-[10px]">
                                                    {failure.failureRate.toFixed(1)}%
                                                </Badge>
                                            </TableCell>
                                        </TableRow>
                                    ))
                                )}
                            </TableBody>
                        </Table>
                    </CardContent>
                </Card>
            </div>
        </div>
    );
}
