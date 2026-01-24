/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useEffect } from "react";
import { StackEditor } from "@/components/stacks/stack-editor";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { RefreshCcw, Activity, PlayCircle, StopCircle, Trash2, Box } from "lucide-react";
import { use } from "react";
import { apiClient } from "@/lib/client";

// Placeholder for StackStatus if we want a separate component
/**
 * StackStatus component.
 * @param props - The component props.
 * @param props.stackId - The unique identifier for stack.
 * @returns The rendered component.
 */
function StackStatus({ stackId }: { stackId: string }) {
    const [services, setServices] = useState<any[]>([]);
    const [isLoading, setIsLoading] = useState(true);

    const fetchStatus = async () => {
        setIsLoading(true);
        try {
            const collection = await apiClient.getCollection(stackId);
            const svcList = collection.services || [];

            const servicesWithStatus = await Promise.all(svcList.map(async (svc: any) => {
                // Default values if status fetch fails
                let status = "Unknown";
                let metrics = { uptime: "-", cpu: "-", mem: "-" };

                try {
                    const statusData = await apiClient.getServiceStatus(svc.name);
                    status = statusData.status || "Unknown";
                    if (statusData.metrics) {
                         metrics = { ...metrics, ...statusData.metrics };
                    }
                } catch (e) {
                    // ignore error, keep defaults
                }

                return {
                    name: svc.name,
                    status: status,
                    uptime: metrics.uptime || "-",
                    cpu: metrics.cpu || "-",
                    mem: metrics.mem || "-"
                };
            }));
            setServices(servicesWithStatus);
        } catch (error) {
            console.error("Failed to load stack status", error);
        } finally {
            setIsLoading(false);
        }
    };

    useEffect(() => {
        fetchStatus();
    }, [stackId]);

    return (
        <div className="space-y-4">
             <div className="flex items-center gap-4">
                 <Card className="flex-1">
                    <CardHeader className="pb-2">
                        <CardTitle className="text-sm font-medium">Total Services</CardTitle>
                    </CardHeader>
                    <CardContent>
                        <div className="text-2xl font-bold">{services.length}</div>
                    </CardContent>
                 </Card>
                 <Card className="flex-1">
                    <CardHeader className="pb-2">
                        <CardTitle className="text-sm font-medium">Running</CardTitle>
                    </CardHeader>
                    <CardContent>
                        <div className="text-2xl font-bold text-green-500">
                            {services.filter(s => s.status === "Active" || s.status === "Running").length}
                        </div>
                    </CardContent>
                 </Card>
                 <Card className="flex-1">
                    <CardHeader className="pb-2">
                        <CardTitle className="text-sm font-medium">Errors</CardTitle>
                    </CardHeader>
                    <CardContent>
                        <div className="text-2xl font-bold text-muted-foreground">
                             {/* Placeholder logic for errors */}
                             0
                        </div>
                    </CardContent>
                 </Card>
             </div>

             <Card>
                <CardHeader className="flex flex-row items-center justify-between">
                    <div>
                        <CardTitle className="text-lg">Runtime Status</CardTitle>
                        <CardDescription>Live status of services defined in this stack.</CardDescription>
                    </div>
                    <Button variant="ghost" size="sm" onClick={fetchStatus} disabled={isLoading}>
                        <RefreshCcw className={`h-4 w-4 ${isLoading ? 'animate-spin' : ''}`} />
                    </Button>
                </CardHeader>
                <CardContent>
                    <div className="rounded-md border">
                        <div className="grid grid-cols-6 gap-4 p-4 border-b font-medium text-sm bg-muted/50">
                            <div className="col-span-2">Service Name</div>
                            <div>Status</div>
                            <div>Uptime</div>
                            <div>CPU</div>
                            <div className="text-right">Actions</div>
                        </div>
                        {services.length === 0 && !isLoading && (
                            <div className="p-4 text-center text-sm text-muted-foreground">No services in this stack.</div>
                        )}
                        {services.map((svc) => (
                            <div key={svc.name} className="grid grid-cols-6 gap-4 p-4 items-center text-sm border-b last:border-0 hover:bg-muted/10 transition-colors">
                                <div className="col-span-2 font-mono flex items-center gap-2">
                                    <Box className="h-4 w-4 text-muted-foreground" />
                                    {svc.name}
                                </div>
                                <div>
                                    <Badge variant={(svc.status === "Running" || svc.status === "Active") ? "default" : "secondary"} className={(svc.status === "Running" || svc.status === "Active") ? "bg-green-500 hover:bg-green-600" : ""}>
                                        {svc.status}
                                    </Badge>
                                </div>
                                <div className="text-muted-foreground">{svc.uptime}</div>
                                <div className="text-muted-foreground font-mono text-xs">{svc.cpu} / {svc.mem}</div>
                                <div className="flex justify-end gap-2">
                                    {(svc.status === "Running" || svc.status === "Active") ? (
                                        <Button variant="ghost" size="icon" className="h-8 w-8 text-destructive" title="Stop">
                                            <StopCircle className="h-4 w-4" />
                                        </Button>
                                    ) : (
                                        <Button variant="ghost" size="icon" className="h-8 w-8 text-green-500" title="Start">
                                            <PlayCircle className="h-4 w-4" />
                                        </Button>
                                    )}
                                    <Button variant="ghost" size="icon" className="h-8 w-8" title="Logs">
                                        <Activity className="h-4 w-4" />
                                    </Button>
                                </div>
                            </div>
                        ))}
                    </div>
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
    const [activeTab, setActiveTab] = useState("editor");

    return (
        <div className="flex-1 space-y-4 p-8 pt-6 h-[calc(100vh-4rem)] flex flex-col">
            <div className="flex items-center justify-between">
                <div className="flex flex-col gap-1">
                     <h2 className="text-3xl font-bold tracking-tight flex items-center gap-2">
                        {resolvedParams.stackId}
                        <Badge variant="outline" className="text-xs font-normal">Stack</Badge>
                     </h2>
                </div>
                <div className="flex items-center gap-2">
                    {/* Refresh button here refreshes the page/tabs? Or specific component? */}
                    {/* For now let's leave it generic or connected to context. But component has internal refresh */}
                    <Button variant="outline" size="sm" onClick={() => window.location.reload()}>
                        <RefreshCcw className="mr-2 h-4 w-4" /> Refresh
                    </Button>
                    {activeTab === "editor" && (
                         <Button size="sm" onClick={() => {
                             // This button duplicates the Save inside Editor,
                             // maybe just let Editor handle it or use a global context/ref
                         }} className="hidden">
                            Deploy Stack
                        </Button>
                    )}
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
