/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useEffect, useState } from "react";
import { IndexStats } from "@/components/discovery/index-stats";
import { SearchConsole } from "@/components/discovery/search-console";
import { discoveryApi } from "@/lib/api-discovery";
import { IndexStatus } from "@proto/api/v1/discovery_service";
import { Activity } from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";

/**
 * The main discovery and search page.
 *
 * @returns The rendered DiscoveryPage component.
 */
export default function DiscoveryPage() {
    const [stats, setStats] = useState<IndexStatus | null>(null);
    const [loading, setLoading] = useState(true);

    useEffect(() => {
        const fetchStats = async () => {
            try {
                const data = await discoveryApi.getIndexStatus();
                setStats(data);
            } catch (error) {
                console.error("Failed to fetch discovery stats", error);
            } finally {
                setLoading(false);
            }
        };

        fetchStats();
    }, []);

    return (
        <div className="flex-1 space-y-4 p-8 pt-6">
            <div className="flex items-center justify-between space-y-2">
                <h1 className="text-3xl font-bold tracking-tight">Discovery & Search</h1>
                <div className="flex items-center space-x-2">
                    <span className="text-sm text-muted-foreground">
                        Manage the Semantic Tool Index
                    </span>
                </div>
            </div>

            <IndexStats stats={stats} loading={loading} />

            <div className="grid gap-4 md:grid-cols-7 lg:grid-cols-8">
                <div className="col-span-4 lg:col-span-5">
                    <SearchConsole />
                </div>
                <div className="col-span-3 lg:col-span-3">
                    <Card className="h-full">
                        <CardHeader>
                            <CardTitle className="text-lg flex items-center gap-2">
                                <Activity className="h-4 w-4" /> Recent Queries
                            </CardTitle>
                        </CardHeader>
                        <CardContent>
                            <div className="text-sm text-muted-foreground text-center py-10">
                                No recent agent queries recorded.
                            </div>
                        </CardContent>
                    </Card>
                </div>
            </div>
        </div>
    );
}
