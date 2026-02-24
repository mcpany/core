/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useEffect } from "react";
import { apiClient, GetDiscoveryStatusResponse } from "@/lib/client";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { RefreshCw, Search, CheckCircle2, AlertTriangle, Clock } from "lucide-react";
import { formatDistanceToNow } from "date-fns";

export function DiscoveryStatus() {
    const [status, setStatus] = useState<GetDiscoveryStatusResponse | null>(null);
    const [loading, setLoading] = useState(true);

    const fetchStatus = async () => {
        setLoading(true);
        try {
            const res = await apiClient.getDiscoveryStatus();
            setStatus(res);
        } catch (e) {
            console.error("Failed to fetch discovery status", e);
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        fetchStatus();
    }, []);

    return (
        <Card>
            <CardHeader className="flex flex-row items-center justify-between pb-2">
                <CardTitle className="text-lg font-medium flex items-center gap-2">
                    <Search className="h-5 w-5" />
                    Discovery Status
                </CardTitle>
                <Button variant="ghost" size="sm" onClick={fetchStatus} disabled={loading}>
                    <RefreshCw className={`h-4 w-4 ${loading ? "animate-spin" : ""}`} />
                </Button>
            </CardHeader>
            <CardContent>
                <div className="space-y-4">
                    {status?.providers?.length === 0 && (
                        <div className="text-sm text-muted-foreground text-center py-4">
                            No discovery providers configured.
                        </div>
                    )}
                    {status?.providers?.map((provider) => (
                        <div key={provider.name} className="flex items-center justify-between p-3 border rounded-lg bg-muted/20">
                            <div className="flex items-center gap-3">
                                {provider.status === "OK" ? (
                                    <CheckCircle2 className="h-5 w-5 text-green-500" />
                                ) : (
                                    <AlertTriangle className="h-5 w-5 text-red-500" />
                                )}
                                <div>
                                    <div className="font-medium flex items-center gap-2">
                                        {provider.name}
                                        <Badge variant={provider.status === "OK" ? "outline" : "destructive"} className="text-xs">
                                            {provider.status}
                                        </Badge>
                                    </div>
                                    <div className="text-xs text-muted-foreground mt-1 flex items-center gap-2">
                                        {provider.lastRunAt && (
                                            <span className="flex items-center gap-1">
                                                <Clock className="h-3 w-3" />
                                                {formatDistanceToNow(new Date(provider.lastRunAt), { addSuffix: true })}
                                            </span>
                                        )}
                                        <span>• {provider.discoveredCount} services found</span>
                                    </div>
                                    {provider.lastError && (
                                        <div className="text-xs text-red-500 mt-1 font-mono break-all">
                                            Error: {provider.lastError}
                                        </div>
                                    )}
                                </div>
                            </div>
                        </div>
                    ))}
                </div>
            </CardContent>
        </Card>
    );
}
