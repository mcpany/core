/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useEffect, useState } from "react";
import { apiClient, Subscription, ServiceEntry, UpstreamServiceConfig } from "@/lib/client";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Switch } from "@/components/ui/switch";
import { useToast } from "@/hooks/use-toast";
import { Loader2, Plus, RefreshCw, Trash2, Check, Download } from "lucide-react";

export default function MarketplacePage() {
    const [subscriptions, setSubscriptions] = useState<Subscription[]>([]);
    const [installedServices, setInstalledServices] = useState<UpstreamServiceConfig[]>([]);
    const [isLoading, setIsLoading] = useState(true);
    const { toast } = useToast();

    const fetchData = async () => {
        setIsLoading(true);
        try {
            const [subs, services] = await Promise.all([
                apiClient.listSubscriptions(),
                apiClient.listServices()
            ]);
            setSubscriptions(subs);
            setInstalledServices(services);
        } catch (error) {
            toast({
                title: "Error fetching data",
                description: String(error),
                variant: "destructive"
            });
        } finally {
            setIsLoading(false);
        }
    };

    useEffect(() => {
        fetchData();
    }, []);

    const handleSync = async (id: string) => {
        try {
            await apiClient.syncSubscription(id);
            toast({ title: "Synced subscription" });
            fetchData();
        } catch (error) {
            toast({
                title: "Failed to sync",
                description: String(error),
                variant: "destructive"
            });
        }
    };

    const isInstalled = (serviceName: string) => {
        return installedServices.some(s => s.name === serviceName);
    };

    const handleInstall = async (service: ServiceEntry) => {
        try {
            // Check if already installed
            if (isInstalled(service.name)) {
                toast({ title: "Already installed" });
                return;
            }

            // Map ServiceEntry to UpstreamServiceConfig
            // Note: ServiceEntry.config is a map, UpstreamServiceConfig needs specific structure
            // We'll assume the 'type' in ServiceEntry helps map it, but for now we map basic stdio/http
            // based on the config present.

            const config: UpstreamServiceConfig = {
                name: service.name,
                // Defaulting to stdio if type matches, but we need to parse config
                // The DefaultPopularServices have "command", "args", "env" in config map.
            };

            if (service.type === "stdio") {
                config.command_line_service = {
                    command: service.config["command"] || "",
                    args: service.config["args"] ? service.config["args"].split(" ") : [],
                    env: service.config["env"] ? parseEnv(service.config["env"]) : undefined
                };
            } else if (service.type === "http") {
               config.http_service = {
                   address: service.config["address"] || ""
               };
            }
            // Add more types if needed

            await apiClient.registerService(config);
            toast({ title: "Service installed", description: service.name });
            fetchData();
        } catch (error) {
            toast({
                title: "Failed to install service",
                description: String(error),
                variant: "destructive"
            });
        }
    };

    const handleUninstall = async (serviceName: string) => {
        try {
            await apiClient.unregisterService(serviceName); // Backend delete usually takes name or ID
            toast({ title: "Service uninstalled", description: serviceName });
            fetchData();
        } catch (error) {
             toast({
                title: "Failed to uninstall service",
                description: String(error),
                variant: "destructive"
            });
        }
    };

    return (
        <div className="space-y-6">
            <div className="flex items-center justify-between">
                <h1 className="text-3xl font-bold tracking-tight">Marketplace</h1>
                <Button onClick={fetchData} variant="outline" size="sm">
                    <RefreshCw className="mr-2 h-4 w-4" />
                    Refresh
                </Button>
            </div>

            <Tabs defaultValue="explore" className="space-y-4">
                <TabsList>
                    <TabsTrigger value="explore">Explore</TabsTrigger>
                    <TabsTrigger value="subscriptions">Manage Subscriptions</TabsTrigger>
                </TabsList>

                <TabsContent value="explore" className="space-y-4">
                     {isLoading ? (
                        <div className="flex justify-center p-8"><Loader2 className="h-8 w-8 animate-spin" /></div>
                     ) : (
                         <div className="space-y-8">
                             {subscriptions.map(sub => (
                                 <div key={sub.id} className="space-y-3">
                                     <h3 className="text-lg font-semibold tracking-tight">{sub.name}</h3>
                                     <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
                                         {(sub.services || []).map((service, idx) => (
                                             <ServiceCard
                                                 key={`${sub.id}-${idx}`}
                                                 service={service}
                                                 installed={isInstalled(service.name)}
                                                 onInstall={() => handleInstall(service)}
                                                 onUninstall={() => handleUninstall(service.name)}
                                             />
                                         ))}
                                     </div>
                                 </div>
                             ))}
                             {subscriptions.length === 0 && <div className="text-center text-muted-foreground">No subscriptions found.</div>}
                         </div>
                     )}
                </TabsContent>

                <TabsContent value="subscriptions" className="space-y-4">
                    {/* Implementation for managing subscriptions (add/remove) */}
                     <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
                        {subscriptions.map(sub => (
                            <Card key={sub.id}>
                                <CardHeader>
                                    <div className="flex justify-between items-start">
                                        <CardTitle>{sub.name}</CardTitle>
                                        <Badge variant={sub.is_active ? "default" : "secondary"}>
                                            {sub.is_active ? "Active" : "Inactive"}
                                        </Badge>
                                    </div>
                                    <CardDescription>{sub.description}</CardDescription>
                                </CardHeader>
                                <CardContent>
                                    <div className="text-xs text-muted-foreground mb-4">
                                        Source: {sub.source_url}
                                    </div>
                                    <div className="flex gap-2">
                                        <Button size="sm" variant="outline" onClick={() => handleSync(sub.id!)}>
                                            <RefreshCw className="mr-2 h-4 w-4" /> Sync
                                        </Button>
                                        {/* Add Delete/Edit button if needed */}
                                    </div>
                                </CardContent>
                            </Card>
                        ))}
                     </div>
                </TabsContent>
            </Tabs>
        </div>
    );
}

function ServiceCard({ service, installed, onInstall, onUninstall }: { service: ServiceEntry, installed: boolean, onInstall: () => void, onUninstall: () => void }) {
    return (
        <Card>
            <CardHeader>
                <CardTitle className="flex justify-between items-center">
                    {service.name}
                    {installed && <Badge variant="secondary"><Check className="h-3 w-3 mr-1" /> Installed</Badge>}
                </CardTitle>
                <CardDescription>{service.description}</CardDescription>
            </CardHeader>
            <CardContent>
                <div className="flex justify-end">
                    {installed ? (
                         <Button variant="outline" size="sm" onClick={onUninstall}>
                            <Trash2 className="mr-2 h-4 w-4" /> Uninstall
                        </Button>
                    ) : (
                        <Button size="sm" onClick={onInstall}>
                            <Download className="mr-2 h-4 w-4" /> Install
                        </Button>
                    )}
                </div>
            </CardContent>
        </Card>
    );
}

function parseEnv(envStr: string): Record<string, string> {
    const env: Record<string, string> = {};
    if (!envStr) return env;
    // Simple parsing for "KEY=VAL,KEY2=VAL2" or just "KEY=VAL"
    // The defaults I set have simple strings.
    // Ideally use real parsing if complex.
    // For "DefaultPopularServices", config["env"] is "BRAVE_API_KEY=<YOUR_KEY>"
    const parts = envStr.split(",");
    for (const part of parts) {
        const [k, v] = part.split("=");
        if (k && v) env[k.trim()] = v.trim();
    }
    return env;
}
