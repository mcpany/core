/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useEffect, useState } from "react";
import { useParams, useRouter } from "next/navigation";
import { apiClient, UpstreamServiceConfig } from "@/lib/client";
import { Button } from "@/components/ui/button";
import { useToast } from "@/hooks/use-toast";
import { Loader2, ArrowLeft, Trash2, Settings, Monitor, Wrench, FileText, Terminal } from "lucide-react";
import { Badge } from "@/components/ui/badge";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { ServiceEditor } from "@/components/services/editor/service-editor";
import { ServiceOverview } from "@/components/services/service-overview";
import { ServiceTools } from "@/components/services/service-tools";
import { ServiceResources } from "@/components/services/service-resources";
import { LogStream } from "@/components/logs/log-stream";

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
    const [loading, setLoading] = useState(true);
    const [activeTab, setActiveTab] = useState("overview");

    // Fetch Service
    useEffect(() => {
        const fetchService = async () => {
            if (!serviceId) return;
            try {
                // Try fetching by ID (which might be the name based on our Instantiate logic)
                const data = await apiClient.getService(serviceId);
                setService(data.service || data); // Adjust based on actual response structure
            } catch (e) {
                console.error(e);
                toast({ title: "Failed to load service", description: "Service not found or error occurred.", variant: "destructive" });
                // Fallback: redirects or show error state
            } finally {
                setLoading(false);
            }
        };
        fetchService();
    }, [serviceId, toast]);

    const handleSave = async () => {
        if (!service) return;
        try {
            await apiClient.updateService(service);
            toast({ title: "Service Updated", description: "Configuration saved successfully." });
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
                     <Button variant="destructive" size="sm" onClick={handleUnregister}>
                        <Trash2 className="mr-2 h-4 w-4" />
                        Unregister
                    </Button>
                </div>
            </div>

            {/* Content */}
            <div className="flex-1 overflow-hidden flex flex-col">
                <Tabs value={activeTab} onValueChange={setActiveTab} className="flex-1 flex flex-col overflow-hidden">
                    <div className="border-b px-4 bg-background">
                        <TabsList className="bg-transparent h-12">
                            <TabsTrigger value="overview" className="data-[state=active]:bg-muted/50 data-[state=active]:shadow-none border-b-2 border-transparent data-[state=active]:border-primary rounded-none h-12 px-4">
                                <Monitor className="mr-2 h-4 w-4" /> Overview
                            </TabsTrigger>
                            <TabsTrigger value="tools" className="data-[state=active]:bg-muted/50 data-[state=active]:shadow-none border-b-2 border-transparent data-[state=active]:border-primary rounded-none h-12 px-4">
                                <Wrench className="mr-2 h-4 w-4" /> Tools
                            </TabsTrigger>
                            <TabsTrigger value="resources" className="data-[state=active]:bg-muted/50 data-[state=active]:shadow-none border-b-2 border-transparent data-[state=active]:border-primary rounded-none h-12 px-4">
                                <FileText className="mr-2 h-4 w-4" /> Resources
                            </TabsTrigger>
                            <TabsTrigger value="logs" className="data-[state=active]:bg-muted/50 data-[state=active]:shadow-none border-b-2 border-transparent data-[state=active]:border-primary rounded-none h-12 px-4">
                                <Terminal className="mr-2 h-4 w-4" /> Logs
                            </TabsTrigger>
                            <TabsTrigger value="settings" className="data-[state=active]:bg-muted/50 data-[state=active]:shadow-none border-b-2 border-transparent data-[state=active]:border-primary rounded-none h-12 px-4">
                                <Settings className="mr-2 h-4 w-4" /> Settings
                            </TabsTrigger>
                        </TabsList>
                    </div>

                    <div className="flex-1 overflow-y-auto bg-muted/5">
                        <TabsContent value="overview" className="p-6 m-0 h-full">
                            <ServiceOverview service={service} />
                        </TabsContent>
                        <TabsContent value="tools" className="p-6 m-0 h-full">
                            <ServiceTools service={service} />
                        </TabsContent>
                        <TabsContent value="resources" className="p-6 m-0 h-full">
                            <ServiceResources service={service} />
                        </TabsContent>
                        <TabsContent value="logs" className="p-0 m-0 h-full flex flex-col">
                            <div className="p-4 h-full">
                                <LogStream source={service.name} />
                            </div>
                        </TabsContent>
                        <TabsContent value="settings" className="p-0 m-0 h-full">
                             <ServiceEditor
                                service={service}
                                onChange={setService}
                                onSave={handleSave}
                                onCancel={() => setActiveTab("overview")}
                            />
                        </TabsContent>
                    </div>
                </Tabs>
            </div>
        </div>
    );
}
