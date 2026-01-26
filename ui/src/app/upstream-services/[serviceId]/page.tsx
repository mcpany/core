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
import { Loader2, ArrowLeft, Power, Trash2, Save, Play, Activity, CheckCircle2, XCircle, AlertTriangle } from "lucide-react";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Badge } from "@/components/ui/badge";
import { Separator } from "@/components/ui/separator";

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
    const [saving, setSaving] = useState(false);
    const [testing, setTesting] = useState(false);
    const [diagnosticResult, setDiagnosticResult] = useState<{ valid: boolean; error?: string; details?: string } | null>(null);

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

    const handleTestConnection = async () => {
        if (!service) return;
        setTesting(true);
        setDiagnosticResult(null);
        try {
            // validateService sends the current config state to the backend for validation
            const result = await apiClient.validateService(service);
            setDiagnosticResult(result);
            if (result.valid) {
                toast({ title: "Connection Successful", description: "Service is reachable and configured correctly." });
            } else {
                toast({
                    title: "Connection Failed",
                    description: result.error ? `${result.error}${result.details ? `: ${result.details}` : ''}` : "Unknown error",
                    variant: "destructive"
                });
            }
        } catch (e) {
            const errStr = String(e);
            setDiagnosticResult({ valid: false, error: "Validation Error", details: errStr });
            toast({ title: "Validation Error", description: errStr, variant: "destructive" });
        } finally {
            setTesting(false);
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
                    <Button variant="outline" onClick={handleTestConnection} disabled={testing}>
                        {testing ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : <Play className="mr-2 h-4 w-4" />}
                        Test Connection
                    </Button>
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
                    <TabsTrigger value="diagnostics">Diagnostics</TabsTrigger>
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

                {/* DIAGNOSTICS TAB */}
                <TabsContent value="diagnostics" className="mt-6 space-y-6">
                    <Card>
                        <CardHeader>
                            <CardTitle>Diagnostics & Health</CardTitle>
                            <CardDescription>Check connectivity and validate configuration.</CardDescription>
                        </CardHeader>
                        <CardContent className="space-y-6">
                            {/* Status Card */}
                            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                                <div className="p-4 border rounded-lg flex items-center gap-4 bg-card">
                                    <div className={`p-2 rounded-full ${service.lastError ? 'bg-red-100 text-red-600 dark:bg-red-900/30 dark:text-red-400' : 'bg-green-100 text-green-600 dark:bg-green-900/30 dark:text-green-400'}`}>
                                        {service.lastError ? <AlertTriangle className="h-6 w-6" /> : <Activity className="h-6 w-6" />}
                                    </div>
                                    <div>
                                        <div className="font-semibold">Service Status</div>
                                        <div className="text-sm text-muted-foreground">
                                            {service.lastError ? "Error Detected" : "Healthy"}
                                        </div>
                                    </div>
                                </div>
                            </div>

                            {service.lastError && (
                                <div className="p-4 bg-destructive/10 text-destructive rounded-md border border-destructive/20">
                                    <div className="font-semibold mb-1 flex items-center gap-2">
                                        <XCircle className="h-4 w-4" />
                                        Last Error
                                    </div>
                                    <pre className="text-xs whitespace-pre-wrap mt-2 font-mono bg-background/50 p-2 rounded">{service.lastError}</pre>
                                </div>
                            )}

                            <Separator />

                            <div className="space-y-4">
                                <div className="flex items-center justify-between">
                                    <div>
                                        <h3 className="text-lg font-medium">Connectivity Check</h3>
                                        <p className="text-sm text-muted-foreground">
                                            Verify that the MCP Any server can reach this upstream service.
                                        </p>
                                    </div>
                                    <Button onClick={handleTestConnection} disabled={testing}>
                                        {testing && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
                                        Run Diagnostics
                                    </Button>
                                </div>

                                {diagnosticResult && (
                                    <div className={`p-4 rounded-md border ${diagnosticResult.valid ? 'bg-green-50/50 border-green-200 dark:bg-green-900/10 dark:border-green-900' : 'bg-red-50/50 border-red-200 dark:bg-red-900/10 dark:border-red-900'}`}>
                                        <div className="flex items-start gap-3">
                                            {diagnosticResult.valid ? (
                                                <CheckCircle2 className="h-5 w-5 text-green-600 dark:text-green-400 mt-0.5" />
                                            ) : (
                                                <XCircle className="h-5 w-5 text-red-600 dark:text-red-400 mt-0.5" />
                                            )}
                                            <div className="space-y-1">
                                                <div className={`font-semibold ${diagnosticResult.valid ? 'text-green-700 dark:text-green-300' : 'text-red-700 dark:text-red-300'}`}>
                                                    {diagnosticResult.valid ? "Configuration Valid & Reachable" : "Check Failed"}
                                                </div>
                                                {diagnosticResult.error && (
                                                    <p className="text-sm text-muted-foreground">
                                                        {diagnosticResult.error}
                                                    </p>
                                                )}
                                                {diagnosticResult.details && (
                                                    <pre className="text-xs mt-2 p-2 bg-background/50 rounded border text-muted-foreground whitespace-pre-wrap">
                                                        {diagnosticResult.details}
                                                    </pre>
                                                )}
                                            </div>
                                        </div>
                                    </div>
                                )}
                            </div>
                        </CardContent>
                    </Card>
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
                        <CardFooter className="gap-2">
                            <Button variant="outline" onClick={handleTestConnection} disabled={testing}>
                                {testing ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : <CheckCircle2 className="mr-2 h-4 w-4" />}
                                Validate
                            </Button>
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
