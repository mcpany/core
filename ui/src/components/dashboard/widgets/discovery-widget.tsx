/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useEffect, useState } from "react";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { apiClient } from "@/lib/client";
import { DiscoveryProviderStatus } from "@proto/admin/v1/admin";
import { Badge } from "@/components/ui/badge";
import { Loader2, RefreshCw, AlertTriangle, CheckCircle2, Search } from "lucide-react";
import { Button } from "@/components/ui/button";
import { formatDistanceToNow } from "date-fns";
import { ScrollArea } from "@/components/ui/scroll-area";

/**
 * Widget that displays the status of auto-discovery providers.
 * @returns The rendered component.
 */
export function DiscoveryWidget() {
    const [statuses, setStatuses] = useState<DiscoveryProviderStatus[]>([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);

    const fetchData = async () => {
        setLoading(true);
        setError(null);
        try {
            const res = await apiClient.getDiscoveryStatus();
            setStatuses(res.providers || []);
        } catch (e) {
            console.error("Failed to fetch discovery status", e);
            setError("Failed to load discovery status.");
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        fetchData();
        // Poll every 30 seconds
        const interval = setInterval(fetchData, 30000);
        return () => clearInterval(interval);
    }, []);

    const getStatusColor = (status: string) => {
        if (status === "OK") return "bg-green-500/10 text-green-500 hover:bg-green-500/20 border-green-500/20";
        if (status === "ERROR") return "bg-destructive/10 text-destructive hover:bg-destructive/20 border-destructive/20";
        return "bg-muted text-muted-foreground";
    };

    return (
        <Card className="h-full flex flex-col">
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <div className="space-y-1">
                    <CardTitle className="text-base font-medium">Auto-Discovery</CardTitle>
                    <CardDescription>Status of local service discovery.</CardDescription>
                </div>
                <Button variant="ghost" size="icon" className="h-8 w-8" onClick={fetchData} disabled={loading}>
                     <RefreshCw className={`h-4 w-4 ${loading ? 'animate-spin' : ''}`} />
                </Button>
            </CardHeader>
            <CardContent className="flex-1 overflow-hidden p-0">
                {loading && statuses.length === 0 ? (
                    <div className="flex items-center justify-center h-full text-muted-foreground">
                        <Loader2 className="h-6 w-6 animate-spin mr-2" />
                        <span className="text-sm">Scanning...</span>
                    </div>
                ) : error ? (
                    <div className="flex flex-col items-center justify-center h-full text-destructive p-4 text-center">
                        <AlertTriangle className="h-8 w-8 mb-2 opacity-50" />
                        <p className="text-sm font-medium">{error}</p>
                        <Button variant="outline" size="sm" className="mt-4" onClick={fetchData}>Retry</Button>
                    </div>
                ) : statuses.length === 0 ? (
                    <div className="flex flex-col items-center justify-center h-full text-muted-foreground p-4 text-center">
                        <Search className="h-8 w-8 mb-2 opacity-20" />
                        <p className="text-sm">No discovery providers active.</p>
                    </div>
                ) : (
                    <ScrollArea className="h-full">
                        <div className="space-y-1 p-4 pt-0">
                            {statuses.map((provider) => (
                                <div key={provider.name} className="flex items-start justify-between p-3 rounded-lg border bg-card/50 hover:bg-accent/5 transition-colors">
                                    <div className="space-y-1">
                                        <div className="flex items-center gap-2">
                                            <span className="font-semibold text-sm">{provider.name}</span>
                                            <Badge variant="outline" className={getStatusColor(provider.status)}>
                                                {provider.status}
                                            </Badge>
                                        </div>
                                        {provider.lastError ? (
                                             <p className="text-xs text-destructive mt-1 break-all">
                                                {provider.lastError}
                                             </p>
                                        ) : (
                                            <div className="flex items-center text-xs text-muted-foreground gap-1">
                                                <CheckCircle2 className="h-3 w-3 text-green-500" />
                                                <span>Found {provider.discoveredCount} services</span>
                                            </div>
                                        )}
                                    </div>
                                    <div className="text-[10px] text-muted-foreground whitespace-nowrap ml-2">
                                        {provider.lastRunAt ? formatDistanceToNow(new Date(provider.lastRunAt), { addSuffix: true }) : 'Never'}
                                    </div>
                                </div>
                            ))}
                        </div>
                    </ScrollArea>
                )}
            </CardContent>
        </Card>
    );
}
