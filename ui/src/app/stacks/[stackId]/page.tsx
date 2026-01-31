/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useEffect, useCallback, use } from "react";
import { StackEditor } from "@/components/stacks/stack-editor";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { RefreshCcw, Activity, PlayCircle, StopCircle, Box, AlertTriangle, CheckCircle, XCircle } from "lucide-react";
import { apiClient, UpstreamServiceConfig } from "@/lib/client";
import { useToast } from "@/hooks/use-toast";
import {
    Table,
    TableBody,
    TableCell,
    TableHead,
    TableHeader,
    TableRow,
} from "@/components/ui/table";
import { Sparkline } from "@/components/charts/sparkline";
import { useServiceHealth } from "@/contexts/service-health-context";

// Helper component for rows to use hooks like useServiceHealth
function ServiceRow({ svc, onRestart, onToggle }: { svc: UpstreamServiceConfig, onRestart: (name: string) => void, onToggle: (name: string, disable: boolean) => void }) {
    const { getServiceHistory } = useServiceHealth();
    const history = getServiceHistory(svc.name);
    const latencies = history.map(h => h.latencyMs);
    const maxLatency = Math.max(...latencies, 50);

    const isRunning = !svc.disable;
    const hasError = !!svc.lastError || svc.status === 'ERROR';

    return (
        <TableRow>
            <TableCell className="font-medium">
                <div className="flex items-center gap-2">
                    <Box className="h-4 w-4 text-muted-foreground" />
                    {svc.name}
                </div>
            </TableCell>
            <TableCell>
                {isRunning ? (
                    hasError ? (
                        <Badge variant="destructive" className="flex w-fit items-center gap-1">
                            <AlertTriangle className="h-3 w-3" /> Error
                        </Badge>
                    ) : (
                        <Badge variant="default" className="bg-green-500 hover:bg-green-600 flex w-fit items-center gap-1">
                            <CheckCircle className="h-3 w-3" /> Running
                        </Badge>
                    )
                ) : (
                    <Badge variant="secondary" className="flex w-fit items-center gap-1">
                        <StopCircle className="h-3 w-3" /> Stopped
                    </Badge>
                )}
            </TableCell>
            <TableCell>
                 <div className="w-[80px] h-[24px]">
                    {isRunning && (
                        <Sparkline
                            data={latencies}
                            width={80}
                            height={24}
                            color={hasError ? "#ef4444" : "#22c55e"}
                            max={maxLatency}
                        />
                    )}
                </div>
            </TableCell>
            <TableCell className="font-mono text-xs text-muted-foreground">
                {svc.lastError ? (
                    <span className="text-destructive truncate max-w-[200px] block" title={svc.lastError}>
                        {svc.lastError}
                    </span>
                ) : (
                    <span>-</span>
                )}
            </TableCell>
            <TableCell className="text-right">
                <div className="flex justify-end gap-2">
                    {isRunning ? (
                        <Button variant="ghost" size="icon" className="h-8 w-8 text-destructive hover:text-destructive hover:bg-destructive/10" onClick={() => onToggle(svc.name, true)} title="Stop">
                            <StopCircle className="h-4 w-4" />
                        </Button>
                    ) : (
                        <Button variant="ghost" size="icon" className="h-8 w-8 text-green-500 hover:text-green-600 hover:bg-green-500/10" onClick={() => onToggle(svc.name, false)} title="Start">
                            <PlayCircle className="h-4 w-4" />
                        </Button>
                    )}
                    <Button variant="ghost" size="icon" className="h-8 w-8" onClick={() => onRestart(svc.name)} title="Restart">
                        <Activity className="h-4 w-4" />
                    </Button>
                </div>
            </TableCell>
        </TableRow>
    );
}

/**
 * StackStatus component.
 * @param props - The component props.
 * @param props.stackId - The unique identifier for stack.
 * @returns The rendered component.
 */
function StackStatus({ stackId }: { stackId: string }) {
    const [services, setServices] = useState<UpstreamServiceConfig[]>([]);
    const [isLoading, setIsLoading] = useState(true);
    const { toast } = useToast();

    const fetchStatus = useCallback(async () => {
        setIsLoading(true);
        try {
            const [collection, allServices] = await Promise.all([
                apiClient.getCollection(stackId),
                apiClient.listServices()
            ]);

            const stackSvcList = collection.services || [];
            // Normalize stack services to a set of names for easy lookup
            // Note: collection.services returns configs, which have names.
            const stackSvcNames = new Set(stackSvcList.map((s: any) => s.name));

            // Filter allServices to find those in this stack
            // This ensures we get the runtime status (UpstreamServiceConfig with status/lastError)
            const relevantServices = allServices.filter((s: any) => stackSvcNames.has(s.name));

            // If some services in the stack definition are not yet registered/returned by listServices,
            // we should probably include them as "Unknown" or "Not Created".
            // But for now, let's show what we found.
            setServices(relevantServices);
        } catch (error) {
            console.error("Failed to load stack status", error);
            toast({
                variant: "destructive",
                title: "Error",
                description: "Failed to load stack status. Ensure backend is reachable."
            });
        } finally {
            setIsLoading(false);
        }
    }, [stackId, toast]);

    useEffect(() => {
        fetchStatus();
    }, [fetchStatus]);

    const handleBulkToggle = async (enable: boolean) => {
        const action = enable ? "start" : "stop";
        if (!confirm(`Are you sure you want to ${action} all services in this stack?`)) return;

        // Optimistic update
        setServices(prev => prev.map(s => ({ ...s, disable: !enable })));

        try {
            await Promise.all(services.map(svc => apiClient.setServiceStatus(svc.name, !enable)));
            toast({ title: "Success", description: `All services ${enable ? 'started' : 'stopped'}.` });
            fetchStatus();
        } catch (e) {
             toast({ variant: "destructive", title: "Error", description: "Failed to update some services." });
             fetchStatus();
        }
    };

    const handleRestartAll = async () => {
        if (!confirm(`Are you sure you want to restart all running services?`)) return;

        try {
            const running = services.filter(s => !s.disable);
            await Promise.all(running.map(svc => apiClient.restartService(svc.name)));
            toast({ title: "Success", description: "Restart signal sent to all running services." });
        } catch (e) {
            toast({ variant: "destructive", title: "Error", description: "Failed to restart some services." });
        }
    };

    const handleToggle = async (name: string, disable: boolean) => {
        try {
            await apiClient.setServiceStatus(name, disable);
            setServices(prev => prev.map(s => s.name === name ? { ...s, disable } : s));
            toast({ title: disable ? "Service Stopped" : "Service Started", description: `Service ${name} updated.` });
        } catch (e) {
             toast({ variant: "destructive", title: "Error", description: "Failed to update service." });
        }
    }

    const handleRestart = async (name: string) => {
        try {
            await apiClient.restartService(name);
            toast({ title: "Service Restarted", description: `Service ${name} restarted.` });
        } catch (e) {
             toast({ variant: "destructive", title: "Error", description: "Failed to restart service." });
        }
    }

    const runningCount = services.filter(s => !s.disable).length;
    const errorCount = services.filter(s => s.lastError || s.status === 'ERROR').length;

    return (
        <div className="space-y-6">
             {/* Summary Cards */}
             <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                 <Card>
                    <CardHeader className="pb-2">
                        <CardTitle className="text-sm font-medium text-muted-foreground">Total Services</CardTitle>
                    </CardHeader>
                    <CardContent>
                        <div className="text-3xl font-bold">{services.length}</div>
                    </CardContent>
                 </Card>
                 <Card>
                    <CardHeader className="pb-2">
                        <CardTitle className="text-sm font-medium text-muted-foreground">Healthy / Running</CardTitle>
                    </CardHeader>
                    <CardContent>
                        <div className="flex items-end gap-2">
                             <div className="text-3xl font-bold text-green-500">
                                {runningCount}
                            </div>
                            <div className="mb-1 text-sm text-muted-foreground">
                                / {services.length}
                            </div>
                        </div>
                    </CardContent>
                 </Card>
                 <Card>
                    <CardHeader className="pb-2">
                        <CardTitle className="text-sm font-medium text-muted-foreground">Issues</CardTitle>
                    </CardHeader>
                    <CardContent>
                        <div className={`text-3xl font-bold ${errorCount > 0 ? 'text-destructive' : 'text-muted-foreground'}`}>
                             {errorCount}
                        </div>
                    </CardContent>
                 </Card>
             </div>

             <Card className="overflow-hidden border-muted/50 shadow-sm">
                <CardHeader className="flex flex-col md:flex-row items-start md:items-center justify-between gap-4 bg-muted/20 pb-4">
                    <div>
                        <CardTitle className="text-lg">Services</CardTitle>
                        <CardDescription>Live status of services defined in this stack.</CardDescription>
                    </div>
                    <div className="flex items-center gap-2">
                         <Button variant="outline" size="sm" onClick={fetchStatus} disabled={isLoading}>
                            <RefreshCcw className={`mr-2 h-4 w-4 ${isLoading ? 'animate-spin' : ''}`} /> Refresh
                        </Button>
                        <Button variant="outline" size="sm" onClick={() => handleBulkToggle(true)}>
                            <PlayCircle className="mr-2 h-4 w-4 text-green-600" /> Start All
                        </Button>
                        <Button variant="outline" size="sm" onClick={() => handleBulkToggle(false)}>
                            <StopCircle className="mr-2 h-4 w-4 text-muted-foreground" /> Stop All
                        </Button>
                        <Button variant="outline" size="sm" onClick={handleRestartAll}>
                            <Activity className="mr-2 h-4 w-4" /> Restart All
                        </Button>
                    </div>
                </CardHeader>
                <CardContent className="p-0">
                    <Table>
                        <TableHeader>
                            <TableRow className="bg-muted/10">
                                <TableHead>Service Name</TableHead>
                                <TableHead>Status</TableHead>
                                <TableHead>Latency (24h)</TableHead>
                                <TableHead>Last Error</TableHead>
                                <TableHead className="text-right">Actions</TableHead>
                            </TableRow>
                        </TableHeader>
                        <TableBody>
                            {services.length === 0 && !isLoading && (
                                <TableRow>
                                    <TableCell colSpan={5} className="h-24 text-center text-muted-foreground">
                                        No services found in this stack.
                                    </TableCell>
                                </TableRow>
                            )}
                            {services.map((svc) => (
                                <ServiceRow key={svc.name} svc={svc} onRestart={handleRestart} onToggle={handleToggle} />
                            ))}
                        </TableBody>
                    </Table>
                </CardContent>
             </Card>
        </div>
    );
}

/**
 * StackDetailPage component.
 * @param props - The component props.
 * @param props.params - The params property.
 * @returns The rendered component.
 */
export default function StackDetailPage({ params }: { params: Promise<{ stackId: string }> }) {
    const resolvedParams = use(params);
    const [activeTab, setActiveTab] = useState("status");

    return (
        <div className="flex-1 space-y-4 p-8 pt-6 h-[calc(100vh-4rem)] flex flex-col">
            <div className="flex items-center justify-between">
                <div className="flex flex-col gap-1">
                     <h2 className="text-3xl font-bold tracking-tight flex items-center gap-2">
                        {resolvedParams.stackId}
                        <Badge variant="outline" className="text-xs font-normal">Stack</Badge>
                     </h2>
                </div>
            </div>

            <Tabs value={activeTab} onValueChange={setActiveTab} className="flex-1 flex flex-col space-y-4">
                <TabsList>
                    <TabsTrigger value="status">Overview & Status</TabsTrigger>
                    <TabsTrigger value="editor">Editor</TabsTrigger>
                </TabsList>

                <TabsContent value="status" className="flex-1">
                     <StackStatus stackId={resolvedParams.stackId} />
                </TabsContent>

                <TabsContent value="editor" className="flex-1 flex flex-col h-full min-h-0">
                    <StackEditor stackId={resolvedParams.stackId} />
                </TabsContent>
            </Tabs>
        </div>
    );
}
