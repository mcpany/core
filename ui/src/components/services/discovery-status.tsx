"use client";

import { useEffect, useState } from "react";
import { apiClient } from "@/lib/client";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { CheckCircle, XCircle, Clock, RefreshCw } from "lucide-react";
import { formatDistanceToNow } from "date-fns";
import { Button } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";

interface DiscoveryProviderStatus {
    name: string;
    status: string;
    lastError?: string;
    lastRunAt?: string;
    discoveredCount: number;
}

export function DiscoveryStatus() {
    const [providers, setProviders] = useState<DiscoveryProviderStatus[]>([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);

    const fetchStatus = async () => {
        setLoading(true);
        setError(null);
        try {
            const res = await apiClient.getDiscoveryStatus();
            setProviders(res.providers || []);
        } catch (e: any) {
            console.error(e);
            setError(e.message || "Failed to fetch discovery status");
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        fetchStatus();
    }, []);

    if (loading && providers.length === 0) {
        return (
             <Card>
                <CardHeader>
                    <CardTitle>Auto-Discovery Status</CardTitle>
                    <CardDescription>Status of service discovery providers.</CardDescription>
                </CardHeader>
                <CardContent className="space-y-4">
                     <Skeleton className="h-12 w-full" />
                     <Skeleton className="h-12 w-full" />
                </CardContent>
            </Card>
        );
    }

    return (
        <Card className="backdrop-blur-sm bg-background/50">
            <CardHeader className="flex flex-row items-center justify-between">
                <div>
                    <CardTitle>Auto-Discovery Status</CardTitle>
                    <CardDescription>Status of service discovery providers.</CardDescription>
                </div>
                <Button variant="ghost" size="icon" onClick={fetchStatus} disabled={loading}>
                    <RefreshCw className={`h-4 w-4 ${loading ? "animate-spin" : ""}`} />
                </Button>
            </CardHeader>
            <CardContent>
                {error && (
                    <div className="text-sm text-red-500 mb-4">
                        {error}
                    </div>
                )}
                {providers.length === 0 && !error ? (
                    <div className="text-sm text-muted-foreground text-center py-4">
                        No discovery providers configured.
                    </div>
                ) : (
                    <div className="space-y-4">
                        {providers.map((provider) => (
                            <div key={provider.name} className="flex items-center justify-between p-4 border rounded-lg bg-card/50">
                                <div className="space-y-1 w-full">
                                    <div className="font-medium flex items-center justify-between">
                                        <span className="flex items-center gap-2 text-lg">
                                             {provider.name}
                                        </span>
                                        {provider.status === "OK" ? (
                                            <Badge variant="outline" className="text-green-500 border-green-500 bg-green-500/10">
                                                <CheckCircle className="w-3 h-3 mr-1" /> OK
                                            </Badge>
                                        ) : (
                                             <Badge variant="destructive">
                                                <XCircle className="w-3 h-3 mr-1" /> Error
                                            </Badge>
                                        )}
                                    </div>
                                    <div className="text-sm text-muted-foreground flex items-center gap-4 mt-2">
                                         <div className="flex items-center gap-1">
                                            <Clock className="w-3 h-3" />
                                            {provider.lastRunAt ? formatDistanceToNow(new Date(provider.lastRunAt), { addSuffix: true }) : "Never run"}
                                         </div>
                                         <span className="text-xs">•</span>
                                         <span>{provider.discoveredCount} services discovered</span>
                                    </div>
                                    {provider.lastError && (
                                        <div className="text-xs text-red-500 mt-2 bg-red-500/10 p-2 rounded border border-red-500/20">
                                            Error: {provider.lastError}
                                        </div>
                                    )}
                                </div>
                            </div>
                        ))}
                    </div>
                )}
            </CardContent>
        </Card>
    );
}
