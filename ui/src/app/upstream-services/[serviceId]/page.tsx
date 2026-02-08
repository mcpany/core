/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useEffect, useState, useMemo } from "react";
import { useParams, useRouter } from "next/navigation";
import { apiClient, UpstreamServiceConfig, GetServiceStatusResponse, ResourceDefinition, ToolDefinition } from "@/lib/client";
import { Button } from "@/components/ui/button";
import { useToast } from "@/hooks/use-toast";
import { Loader2, ArrowLeft, Trash2, Settings, Activity, Wrench, FileText, Terminal, RefreshCw } from "lucide-react";
import { Badge } from "@/components/ui/badge";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";

import { ServiceOverview } from "@/components/services/details/service-overview";
import { ServiceTools } from "@/components/services/details/service-tools";
import { ServiceResources } from "@/components/services/details/service-resources";
import { LogStream } from "@/components/logs/log-stream";
import { ServiceEditor } from "@/components/services/editor/service-editor";

/**
 * UpstreamServiceDetailPage component.
 * @returns The rendered component.
 */
export default function UpstreamServiceDetailPage() {
    const params = useParams();
    const router = useRouter();
    const { toast } = useToast();
    const serviceId = params.serviceId as string;

    const [service, setService] = useState<UpstreamServiceConfig | null>(null);
    const [status, setStatus] = useState<GetServiceStatusResponse | null>(null);
    const [loading, setLoading] = useState(true);
    const [refreshing, setRefreshing] = useState(false);

    const fetchServiceData = async () => {
        if (!serviceId) return;
        try {
            // Fetch Service Config
            const data = await apiClient.getService(serviceId);
            const fetchedService = data.service || data;
            setService(fetchedService);

            // Fetch Service Status (Runtime info)
            try {
                // Use name or ID? Status usually keyed by name.
                const statusData = await apiClient.getServiceStatus(fetchedService.name);
                setStatus(statusData);
            } catch (e) {
                console.warn("Failed to fetch service status", e);
                // Fallback status if offline or error
                setStatus({ tools: [], metrics: {} });
            }
        } catch (e) {
            console.error(e);
            toast({ title: "Failed to load service", description: "Service not found or error occurred.", variant: "destructive" });
        } finally {
            setLoading(false);
            setRefreshing(false);
        }
    };

    useEffect(() => {
        fetchServiceData();
    }, [serviceId, toast]);

    const handleRefresh = () => {
        setRefreshing(true);
        fetchServiceData();
    };

    const handleSave = async () => {
        if (!service) return;
        try {
            await apiClient.updateService(service);
            toast({ title: "Service Updated", description: "Configuration saved successfully." });
            fetchServiceData(); // Refresh to ensure sync
        } catch (e) {
            toast({ title: "Update Failed", description: String(e), variant: "destructive" });
        }
    };

    const handleUnregister = async () => {
        if (!confirm("Are you sure you want to unregister this service? This action cannot be undone.")) return;
        try {
            await apiClient.unregisterService(serviceId);
            toast({ title: "Service Unregistered" });
            router.push("/upstream-services");
        } catch (e) {
            toast({ title: "Unregister Failed", description: String(e), variant: "destructive" });
        }
    };

    // Extract resources from config
    const resources = useMemo(() => {
        if (!service) return [];
        return (
            service.httpService?.resources ||
            service.grpcService?.resources ||
            service.commandLineService?.resources ||
            service.mcpService?.resources ||
            service.openapiService?.resources ||
            []
        );
    }, [service]);

    // Combine tools from config and status (status usually has the active/discovered ones)
    const tools = useMemo(() => {
        if (status?.tools && status.tools.length > 0) return status.tools;
        // Fallback to config if status empty (e.g. offline)
        return (
             service?.httpService?.tools ||
             service?.grpcService?.tools ||
             service?.commandLineService?.tools ||
             service?.mcpService?.tools ||
             service?.openapiService?.tools ||
             []
        );
    }, [service, status]);

    if (loading) {
        return <div className="flex h-screen items-center justify-center"><Loader2 className="h-8 w-8 animate-spin" /></div>;
    }

    if (!service) {
        return (
            <div className="p-8 text-center">
                <h1 className="text-2xl font-bold">Service Not Found</h1>
                <Button variant="link" onClick={() => router.push("/upstream-services")}>Back to Services</Button>
            </div>
        );
    }

    return (
        <div className="flex flex-col h-screen overflow-hidden bg-background">
             {/* Header */}
            <div className="flex-none border-b p-4 flex items-center justify-between bg-muted/20">
                <div className="flex items-center gap-4">
                    <Button variant="ghost" size="icon" onClick={() => router.push("/upstream-services")}>
                        <ArrowLeft className="h-5 w-5" />
                    </Button>
                    <div>
                        <h1 className="text-2xl font-bold tracking-tight flex items-center gap-3">
                            {service.name}
                            <Badge variant={service.disable ? "secondary" : "default"} className={service.disable ? "bg-muted text-muted-foreground" : "bg-green-500 hover:bg-green-600"}>
                                {service.disable ? "Disabled" : "Active"}
                            </Badge>
                        </h1>
                         <p className="text-muted-foreground text-xs font-mono">{service.id || "ID not assigned"}</p>
                    </div>
                </div>
                <div className="flex gap-2">
                     <Button variant="outline" size="sm" onClick={handleRefresh} disabled={refreshing}>
                        <RefreshCw className={`mr-2 h-4 w-4 ${refreshing ? "animate-spin" : ""}`} />
                        Refresh
                    </Button>
                     <Button variant="destructive" size="sm" onClick={handleUnregister}>
                        <Trash2 className="mr-2 h-4 w-4" />
                        Unregister
                    </Button>
                </div>
            </div>

            {/* Content Tabs */}
            <div className="flex-1 overflow-hidden p-6">
                <Tabs defaultValue="overview" className="h-full flex flex-col">
                    <div className="flex-none mb-6">
                        <TabsList>
                            <TabsTrigger value="overview"><Activity className="h-4 w-4 mr-2"/> Overview</TabsTrigger>
                            <TabsTrigger value="tools"><Wrench className="h-4 w-4 mr-2"/> Tools ({tools.length})</TabsTrigger>
                            <TabsTrigger value="resources"><FileText className="h-4 w-4 mr-2"/> Resources ({resources.length})</TabsTrigger>
                            <TabsTrigger value="logs"><Terminal className="h-4 w-4 mr-2"/> Logs</TabsTrigger>
                            <TabsTrigger value="settings"><Settings className="h-4 w-4 mr-2"/> Settings</TabsTrigger>
                        </TabsList>
                    </div>

                    <TabsContent value="overview" className="flex-1 overflow-auto mt-0">
                        {status && <ServiceOverview service={service} status={status} />}
                    </TabsContent>

                    <TabsContent value="tools" className="flex-1 overflow-auto mt-0">
                        <ServiceTools tools={tools} />
                    </TabsContent>

                    <TabsContent value="resources" className="flex-1 overflow-auto mt-0">
                        <ServiceResources resources={resources} />
                    </TabsContent>

                    <TabsContent value="logs" className="flex-1 overflow-hidden mt-0 h-full">
                        <LogStream fixedSource={service.name} />
                    </TabsContent>

                    <TabsContent value="settings" className="flex-1 overflow-hidden mt-0">
                        <ServiceEditor
                            service={service}
                            onChange={setService}
                            onSave={handleSave}
                            onCancel={() => router.push("/upstream-services")}
                        />
                    </TabsContent>
                </Tabs>
            </div>
        </div>
    );
}
