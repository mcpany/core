/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useEffect, useState } from "react";
import { apiClient, ListServicesResponse } from "@/lib/client";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle, CardFooter } from "@/components/ui/card";
import { useToast } from "@/hooks/use-toast";
import { Plus, Power, Trash2, RefreshCw, AlertTriangle } from "lucide-react";
import Link from "next/link";
import { Badge } from "@/components/ui/badge";

/**
 * UpstreamServicesPage component.
 * @returns The rendered component.
 */
export default function UpstreamServicesPage() {
    const { toast } = useToast();
    const [services, setServices] = useState<any[]>([]);
    const [loading, setLoading] = useState(true);

    const loadServices = async () => {
        setLoading(true);
        try {
            const data = await apiClient.listServices();
            setServices(data || []);
        } catch (e) {
            console.error(e);
            toast({ title: "Failed to load services", variant: "destructive" });
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        loadServices();
    }, []);

    const toggleStatus = async (service: any) => {
        try {
            await apiClient.setServiceStatus(service.name, !service.disable);
            toast({ title: service.disable ? "Service Enabled" : "Service Disabled" });
            loadServices();
        } catch (e) {
             toast({ title: "Failed to update status", variant: "destructive" });
        }
    };

    const handleDelete = async (id: string) => {
        if (!confirm("Are you sure you want to unregister this service?")) return;
        try {
            await apiClient.unregisterService(id);
            toast({ title: "Service Unregistered" });
            loadServices();
        } catch (e) {
            toast({ title: "Failed to delete", variant: "destructive" });
        }
    };

    return (
        <div className="flex flex-col gap-8 p-8 h-[calc(100vh-4rem)] overflow-y-auto">
             <div className="flex items-center justify-between">
                <div>
                    <h1 className="text-3xl font-bold tracking-tight">Upstream Services</h1>
                    <p className="text-muted-foreground mt-2">
                        Manage active MCP/Upstream services.
                    </p>
                </div>
                <Button asChild>
                    <Link href="/marketplace?tab=local">
                        <Plus className="mr-2 h-4 w-4" />
                        Add Service
                    </Link>
                </Button>
            </div>

            <div className="grid grid-cols-1 gap-6">
                {services.map((service) => (
                    <Card key={service.id || service.name} className={`flex flex-col md:flex-row md:items-center justify-between p-2 ${service.disable ? 'opacity-70' : ''}`}>
                         <Link href={`/upstream-services/${service.name}`} className="flex-1 cursor-pointer">
                            <div className="p-4 hover:bg-muted/50 rounded-lg transition-colors">
                                 <div className="flex items-center gap-3 mb-2">
                                    <h3 className="font-semibold text-lg">{service.name}</h3>
                                    {service.disable ? (
                                        <Badge variant="outline" className="text-muted-foreground">Disabled</Badge>
                                    ) : (
                                        <Badge className="bg-green-500 hover:bg-green-600">Active</Badge>
                                    )}
                                    <span className="text-xs text-muted-foreground font-mono bg-muted px-2 py-0.5 rounded">
                                        {service.version}
                                    </span>
                                 </div>
                                 <p className="text-sm text-muted-foreground mb-2">
                                     Type: {service.commandLineService ? 'Command Line' : service.httpService ? 'HTTP' : 'MCP'}
                                 </p>
                                 {service.lastError && (
                                     <div className="text-sm text-destructive flex items-center gap-1 mt-2">
                                         <AlertTriangle className="h-3 w-3" />
                                         Error: {service.lastError}
                                     </div>
                                 )}
                            </div>
                         </Link>

                         <div className="flex items-center gap-2 p-4 md:border-l">
                             <Button size="sm" variant="ghost" onClick={() => toggleStatus(service)}>
                                 <Power className={`h-4 w-4 ${service.disable ? 'text-muted-foreground' : 'text-green-600'}`} />
                             </Button>
                             <Button size="sm" variant="ghost" onClick={loadServices}>
                                 <RefreshCw className="h-4 w-4" />
                             </Button>
                             <Button size="sm" variant="ghost" onClick={() => handleDelete(service.name)}>
                                 <Trash2 className="h-4 w-4 text-destructive" />
                             </Button>
                         </div>
                    </Card>
                ))}
                {services.length === 0 && !loading && (
                    <div className="text-center p-12 text-muted-foreground border rounded-lg border-dashed">
                        No active services. Go to Marketplace to add one.
                    </div>
                )}
            </div>
        </div>
    );
}
