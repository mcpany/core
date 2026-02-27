/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Database, Clock, Layers } from "lucide-react";
import { IndexStatus } from "@proto/api/v1/discovery_service";
import { Skeleton } from "@/components/ui/skeleton";

interface IndexStatsProps {
    stats: IndexStatus | null;
    loading: boolean;
}

/**
 * Renders the top-level index statistics.
 *
 * @param props - The component props.
 * @param props.stats - The current status of the index.
 * @param props.loading - Whether the stats are currently loading.
 * @returns The rendered IndexStats component.
 */
export function IndexStats({ stats, loading }: IndexStatsProps) {
    if (loading) {
        return (
            <div className="grid gap-4 md:grid-cols-3">
                <Skeleton className="h-32 rounded-xl" />
                <Skeleton className="h-32 rounded-xl" />
                <Skeleton className="h-32 rounded-xl" />
            </div>
        );
    }

    return (
        <div className="grid gap-4 md:grid-cols-3">
            <Card>
                <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                    <CardTitle className="text-sm font-medium">Total Tools</CardTitle>
                    <Database className="h-4 w-4 text-muted-foreground" />
                </CardHeader>
                <CardContent>
                    <div className="text-2xl font-bold">{stats?.totalTools || 0}</div>
                    <p className="text-xs text-muted-foreground">
                        Available across all services
                    </p>
                </CardContent>
            </Card>
            <Card>
                <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                    <CardTitle className="text-sm font-medium">Indexed Tools</CardTitle>
                    <Layers className="h-4 w-4 text-muted-foreground" />
                </CardHeader>
                <CardContent>
                    <div className="text-2xl font-bold">{stats?.indexedTools || 0}</div>
                    <p className="text-xs text-muted-foreground">
                        Searchable in vector index
                    </p>
                </CardContent>
            </Card>
            <Card>
                <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                    <CardTitle className="text-sm font-medium">Last Updated</CardTitle>
                    <Clock className="h-4 w-4 text-muted-foreground" />
                </CardHeader>
                <CardContent>
                    <div className="text-2xl font-bold">
                        {stats?.lastUpdated ? new Date(stats.lastUpdated).toLocaleTimeString() : "Never"}
                    </div>
                    <p className="text-xs text-muted-foreground">
                        Index refresh timestamp
                    </p>
                </CardContent>
            </Card>
        </div>
    );
}
