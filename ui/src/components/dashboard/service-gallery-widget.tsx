/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useEffect, useState } from "react";
import { apiClient, UpstreamServiceConfig } from "@/lib/client";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";
import { Server, Zap, Globe, Terminal, Settings } from "lucide-react";
import Link from "next/link";
import { cn } from "@/lib/utils";

/**
 * ServiceGalleryWidget displays a grid of connected services as sleek cards.
 * @returns The rendered widget.
 */
export function ServiceGalleryWidget() {
    const [services, setServices] = useState<UpstreamServiceConfig[]>([]);
    const [loading, setLoading] = useState(true);

    useEffect(() => {
        const loadServices = async () => {
            try {
                const res = await apiClient.listServices();
                if (Array.isArray(res)) {
                    setServices(res);
                } else if (res && Array.isArray(res.services)) {
                    setServices(res.services);
                }
            } catch (error) {
                console.error("Failed to fetch services:", error);
            } finally {
                setLoading(false);
            }
        };

        loadServices();
    }, []);

    const getIcon = (type: string | undefined) => {
        switch (type?.toLowerCase()) {
            case 'http': return <Globe className="h-5 w-5 text-blue-500" />;
            case 'grpc': return <Zap className="h-5 w-5 text-amber-500" />;
            case 'command': return <Terminal className="h-5 w-5 text-green-500" />;
            default: return <Server className="h-5 w-5 text-muted-foreground" />;
        }
    };

    if (loading) {
        return (
            <Card className="h-full flex flex-col min-h-[300px] border-none shadow-none bg-transparent">
                <CardHeader className="pb-2">
                    <CardTitle className="text-lg flex items-center gap-2">
                        <Server className="h-5 w-5" /> Active Services
                    </CardTitle>
                </CardHeader>
                <CardContent className="flex-1 overflow-auto">
                    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4">
                        {[1, 2, 3].map(i => (
                            <Skeleton key={i} className="h-32 w-full rounded-xl" />
                        ))}
                    </div>
                </CardContent>
            </Card>
        );
    }

    if (services.length === 0) {
        return (
            <Card className="h-full flex flex-col min-h-[300px] backdrop-blur-md bg-background/60 border-muted/50 shadow-sm transition-all hover:shadow-md">
                <CardHeader className="pb-2 border-b border-muted/20">
                    <CardTitle className="text-lg flex items-center gap-2 font-medium tracking-tight">
                        <Server className="h-5 w-5 text-primary" /> Active Services
                    </CardTitle>
                </CardHeader>
                <CardContent className="flex-1 flex flex-col items-center justify-center p-8 text-center space-y-4">
                    <div className="p-4 bg-muted/30 rounded-full border border-muted/50 shadow-inner">
                         <Server className="h-8 w-8 text-muted-foreground/50" />
                    </div>
                    <div className="space-y-1">
                         <p className="text-lg font-medium tracking-tight">No Services Connected</p>
                         <p className="text-sm text-muted-foreground max-w-sm">
                             Connect your first upstream service to unlock AI capabilities.
                         </p>
                    </div>
                    <Button asChild variant="default" className="mt-4 shadow-sm">
                        <Link href="/upstream-services">Add Service</Link>
                    </Button>
                </CardContent>
            </Card>
        );
    }

    return (
        <Card className="h-full flex flex-col min-h-[300px] backdrop-blur-md bg-background/40 border-muted/50 shadow-sm">
            <CardHeader className="pb-4 border-b border-muted/20">
                <div className="flex items-center justify-between">
                    <CardTitle className="text-lg flex items-center gap-2 font-medium tracking-tight">
                        <Server className="h-5 w-5 text-primary" /> Connected Services
                    </CardTitle>
                    <Badge variant="secondary" className="font-mono text-xs rounded-full px-2">
                        {services.length} Total
                    </Badge>
                </div>
            </CardHeader>
            <CardContent className="flex-1 overflow-auto p-4 md:p-6 bg-gradient-to-b from-transparent to-muted/10">
                <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4">
                    {services.map(service => {
                        const isUp = !service.disable;
                        const type = service.httpService ? "HTTP" :
                                     service.grpcService ? "gRPC" :
                                     service.commandLineService ? "Command" : "Unknown";

                        return (
                            <div
                                key={service.id}
                                className={cn(
                                    "group relative flex flex-col p-5 rounded-2xl border transition-all duration-300",
                                    "hover:shadow-lg hover:-translate-y-1 hover:border-primary/30",
                                    isUp ? "bg-card/80 backdrop-blur-xl border-muted" : "bg-muted/30 border-muted opacity-75"
                                )}
                            >
                                <div className="flex justify-between items-start mb-4">
                                    <div className="p-2.5 rounded-xl bg-background shadow-sm border border-muted/50">
                                        {getIcon(type)}
                                    </div>
                                    <Badge
                                        variant={isUp ? "default" : "secondary"}
                                        className={cn(
                                            "text-[10px] uppercase font-semibold px-2 py-0.5 rounded-full",
                                            isUp ? "bg-green-500/10 text-green-600 hover:bg-green-500/20 border-green-500/20" : "bg-muted text-muted-foreground"
                                        )}
                                    >
                                        {isUp ? "Active" : "Disabled"}
                                    </Badge>
                                </div>

                                <div className="mb-4">
                                    <h3 className="font-semibold text-base tracking-tight truncate" title={service.name}>
                                        {service.name}
                                    </h3>
                                    <p className="text-xs text-muted-foreground mt-1 flex items-center gap-1.5">
                                        <span className="inline-block w-1.5 h-1.5 rounded-full bg-primary/50"></span>
                                        {type}
                                    </p>
                                </div>

                                <div className="mt-auto pt-4 border-t border-muted/50 flex items-center justify-between">
                                    <span className="text-xs font-mono text-muted-foreground bg-muted/50 px-2 py-1 rounded-md truncate max-w-[120px]" title={service.id}>
                                        {service.id}
                                    </span>
                                    <Button
                                        asChild
                                        variant="ghost"
                                        size="icon"
                                        className="h-8 w-8 rounded-full opacity-0 group-hover:opacity-100 transition-opacity bg-background shadow-sm border hover:bg-muted"
                                    >
                                        <Link href={`/upstream-services?service=${service.id}`}>
                                            <Settings className="h-4 w-4" />
                                        </Link>
                                    </Button>
                                </div>
                            </div>
                        );
                    })}
                </div>
            </CardContent>
        </Card>
    );
}
