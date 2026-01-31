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
import { RefreshCcw, RefreshCw, Activity, PlayCircle, StopCircle, Trash2, Box, Rocket, AlertTriangle } from "lucide-react";
import { use } from "react";
import { apiClient } from "@/lib/client";
import { toast } from "sonner";
import { useRouter } from "next/navigation";
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from "@/components/ui/alert-dialog";

/**
 * StackStatus component.
 * @param props - The component props.
 * @param props.stackId - The unique identifier for stack.
 * @returns The rendered component.
 */
function StackStatus({ stackId }: { stackId: string }) {
    const [services, setServices] = useState<any[]>([]);
    const [isLoading, setIsLoading] = useState(true);
    const [errorCount, setErrorCount] = useState(0);

    const fetchStatus = async () => {
        setIsLoading(true);
        try {
            const collection = await apiClient.getCollection(stackId);
            const svcList = collection.services || [];

            let errors = 0;
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
                    // Service might not be registered yet or error
                    status = "Error"; // Or "Not Deployed"
                }

                if (status !== "Active" && status !== "Running") {
                    errors++;
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
            setErrorCount(errors);
        } catch (error) {
            console.error("Failed to load stack status", error);
            toast.error("Failed to load stack status");
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
                        <CardTitle className="text-sm font-medium">Issues</CardTitle>
                    </CardHeader>
                    <CardContent>
                        <div className={`text-2xl font-bold ${errorCount > 0 ? "text-red-500" : "text-muted-foreground"}`}>
                             {errorCount}
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
                        <RefreshCw className={`h-4 w-4 ${isLoading ? 'animate-spin' : ''}`} />
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
                                    <Badge
                                        variant={(svc.status === "Running" || svc.status === "Active") ? "default" : "secondary"}
                                        className={
                                            (svc.status === "Running" || svc.status === "Active") ? "bg-green-500 hover:bg-green-600" :
                                            (svc.status === "Error" || svc.status === "Unknown") ? "bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-300" : ""
                                        }
                                    >
                                        {svc.status === "Error" && <AlertTriangle className="h-3 w-3 mr-1" />}
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
    const [isDeploying, setIsDeploying] = useState(false);
    const router = useRouter();

    const handleDeploy = async () => {
        setIsDeploying(true);
        try {
            await apiClient.applyCollection(resolvedParams.stackId);
            toast.success("Stack deployment initiated");
            setActiveTab("status");
        } catch (e) {
            console.error(e);
            toast.error("Failed to deploy stack");
        } finally {
            setIsDeploying(false);
        }
    };

    const handleDelete = async () => {
        try {
            await apiClient.deleteCollection(resolvedParams.stackId);
            toast.success("Stack deleted");
            router.push("/stacks");
        } catch (e) {
            console.error(e);
            toast.error("Failed to delete stack");
        }
    };

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
                    <Button variant="outline" size="sm" onClick={() => window.location.reload()}>
                        <RefreshCw className="mr-2 h-4 w-4" /> Refresh
                    </Button>
                    <Button
                        size="sm"
                        onClick={handleDeploy}
                        disabled={isDeploying}
                        className="bg-green-600 hover:bg-green-700 text-white"
                    >
                        {isDeploying ? <RefreshCw className="mr-2 h-4 w-4 animate-spin" /> : <Rocket className="mr-2 h-4 w-4" />}
                        Deploy Stack
                    </Button>

                    <AlertDialog>
                        <AlertDialogTrigger asChild>
                            <Button variant="ghost" size="icon" className="text-destructive hover:bg-destructive/10">
                                <Trash2 className="h-4 w-4" />
                            </Button>
                        </AlertDialogTrigger>
                        <AlertDialogContent>
                            <AlertDialogHeader>
                                <AlertDialogTitle>Delete Stack?</AlertDialogTitle>
                                <AlertDialogDescription>
                                    Are you sure you want to delete this stack? This will NOT delete the running services, only the stack definition from the registry.
                                </AlertDialogDescription>
                            </AlertDialogHeader>
                            <AlertDialogFooter>
                                <AlertDialogCancel>Cancel</AlertDialogCancel>
                                <AlertDialogAction onClick={handleDelete} className="bg-destructive hover:bg-destructive/90">
                                    Delete
                                </AlertDialogAction>
                            </AlertDialogFooter>
                        </AlertDialogContent>
                    </AlertDialog>
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
