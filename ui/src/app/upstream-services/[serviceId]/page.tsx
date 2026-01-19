/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useEffect, useState } from "react";
import { useParams, useRouter } from "next/navigation";
import { apiClient, UpstreamServiceConfig } from "@/lib/client";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from "@/components/ui/card";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { useToast } from "@/hooks/use-toast";
import { Loader2, ArrowLeft, Power, Trash2, Save } from "lucide-react";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Badge } from "@/components/ui/badge";
import { Separator } from "@/components/ui/separator";

export default function UpstreamServiceDetailPage() {
    const params = useParams();
    const router = useRouter();
    const { toast } = useToast();
    const serviceId = params.serviceId as string;

    const [service, setService] = useState<UpstreamServiceConfig | null>(null);
    const [loading, setLoading] = useState(true);
    const [saving, setSaving] = useState(false);

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
        setSaving(true);
        try {
            await apiClient.updateService(service);
            toast({ title: "Service Updated", description: "Configuration saved successfully." });
        } catch (e) {
            toast({ title: "Update Failed", description: String(e), variant: "destructive" });
        } finally {
            setSaving(false);
        }
    };

    const handleToggleStatus = async () => {
        if (!service) return;
        try {
            const newStatus = !service.disable;
            await apiClient.setServiceStatus(service.name, newStatus);
            setService({ ...service, disable: newStatus });
            toast({ title: newStatus ? "Service Disabled" : "Service Enabled" });
        } catch (e) {
            toast({ title: "Action Failed", description: String(e), variant: "destructive" });
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
        <div className="flex flex-col gap-6 p-8 h-[calc(100vh-4rem)] overflow-y-auto">
            {/* Header */}
            <div className="flex items-center justify-between">
                <div className="flex items-center gap-4">
                    <Button variant="ghost" size="icon" onClick={() => router.push("/upstream-services")}>
                        <ArrowLeft className="h-5 w-5" />
                    </Button>
                    <div>
                        <h1 className="text-3xl font-bold tracking-tight flex items-center gap-3">
                            {service.name}
                            <Badge variant={service.disable ? "secondary" : "default"} className={service.disable ? "bg-muted text-muted-foreground" : "bg-green-500 hover:bg-green-600"}>
                                {service.disable ? "Disabled" : "Active"}
                            </Badge>
                        </h1>
                        <p className="text-muted-foreground mt-1 text-sm font-mono">{service.id}</p>
                    </div>
                </div>
                <div className="flex gap-2">
                    <Button variant="outline" onClick={handleToggleStatus}>
                        <Power className="mr-2 h-4 w-4" />
                        {service.disable ? "Enable" : "Disable"}
                    </Button>
                    <Button variant="destructive" onClick={handleUnregister}>
                        <Trash2 className="mr-2 h-4 w-4" />
                        Unregister
                    </Button>
                </div>
            </div>

            <Tabs defaultValue="overview" className="w-full">
                <TabsList>
                    <TabsTrigger value="overview">Overview</TabsTrigger>
                    <TabsTrigger value="config">Configuration</TabsTrigger>
                    <TabsTrigger value="auth">Authentication</TabsTrigger>
                    <TabsTrigger value="webhooks">Webhooks</TabsTrigger>
                </TabsList>

                {/* OVERVIEW TAB */}
                <TabsContent value="overview" className="mt-6 space-y-6">
                    <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                        <Card>
                            <CardHeader>
                                <CardTitle>Service Details</CardTitle>
                            </CardHeader>
                            <CardContent className="space-y-4">
                                <div className="grid grid-cols-2 gap-2 text-sm">
                                    <span className="font-semibold">Type:</span>
                                    <span>
                                        {service.commandLineService ? "Command Line" :
                                         service.httpService ? "HTTP" :
                                         service.mcpService ? "MCP (Remote)" : "Unknown"}
                                    </span>
                                    <span className="font-semibold">Version:</span>
                                    <span>{service.version || "latest"}</span>
                                    <span className="font-semibold">Priority:</span>
                                    <span>{service.priority || 0}</span>
                                </div>
                            </CardContent>
                        </Card>
                         <Card>
                            <CardHeader>
                                <CardTitle>Stats</CardTitle>
                            </CardHeader>
                            <CardContent>
                                <p className="text-muted-foreground text-sm">Real-time stats coming soon.</p>
                            </CardContent>
                        </Card>
                    </div>
                </TabsContent>

                {/* CONFIGURATION TAB */}
                <TabsContent value="config" className="mt-6">
                    <Card>
                        <CardHeader>
                            <CardTitle>Configuration</CardTitle>
                            <CardDescription>Update connection parameters and environment variables.</CardDescription>
                        </CardHeader>
                        <CardContent className="space-y-6">
                            {service.commandLineService && (
                                <div className="space-y-4">
                                    <div className="grid gap-2">
                                        <Label>Command</Label>
                                        <Input
                                            value={service.commandLineService.command}
                                            onChange={(e) => setService({
                                                ...service,
                                                commandLineService: { ...service.commandLineService!, command: e.target.value }
                                            })}
                                        />
                                    </div>
                                    <div className="grid gap-2">
                                        <Label>Working Directory</Label>
                                        <Input
                                            value={service.commandLineService.workingDirectory || ""}
                                            onChange={(e) => setService({
                                                ...service,
                                                commandLineService: { ...service.commandLineService!, workingDirectory: e.target.value }
                                            })}
                                        />
                                    </div>
                                </div>
                            )}
                            {/* Add other service type configs here */}
                        </CardContent>
                        <CardFooter>
                            <Button onClick={handleSave} disabled={saving}>
                                {saving && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
                                <Save className="mr-2 h-4 w-4" />
                                Save Changes
                            </Button>
                        </CardFooter>
                    </Card>
                </TabsContent>

                {/* AUTHENTICATION TAB */}
                <TabsContent value="auth" className="mt-6">
                    <Card>
                        <CardHeader>
                            <CardTitle>Authentication Binding</CardTitle>
                            <CardDescription>
                                Bind a stored credential to this service. This credential will be used when establishing connections to the upstream service.
                            </CardDescription>
                        </CardHeader>
                        <CardContent>
                            <div className="p-4 border rounded-lg bg-muted/20 text-center text-muted-foreground">
                                Authentication Binding UI is under construction.
                                <br/>
                                <Button variant="link">Manage Credentials</Button>
                            </div>
                        </CardContent>
                    </Card>
                </TabsContent>

                 {/* WEBHOOKS TAB */}
                <TabsContent value="webhooks" className="mt-6">
                    <Card>
                         <CardHeader>
                            <CardTitle>Webhooks</CardTitle>
                            <CardDescription>Manage Pre-Call and Post-Call webhooks.</CardDescription>
                        </CardHeader>
                        <CardContent>
                             <div className="p-4 border rounded-lg bg-muted/20 text-center text-muted-foreground">
                                Webhook configuration list will appear here.
                            </div>
                        </CardContent>
                    </Card>
                </TabsContent>
            </Tabs>
        </div>
    );
}
