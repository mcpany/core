/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useEffect } from "react";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Progress } from "@/components/ui/progress";
import { AlertCircle, ArrowUpRight } from "lucide-react";
import { apiClient } from "@/lib/client";
import { useDashboardDensity } from "@/contexts/dashboard-density-context";
import { cn } from "@/lib/utils";

interface ToolFailureRate {
    name: string;
    service: string;
    failureRate: number;
    totalCalls: number;
}

/**
 * ToolFailureRateWidget component.
 * Displays tools with the highest error rates.
 * @returns The rendered component.
 */
export function ToolFailureRateWidget() {
    const [tools, setTools] = useState<ToolFailureRate[]>([]);
    const [loading, setLoading] = useState(true);
    const { density } = useDashboardDensity();
    const isCompact = density === "compact";

    useEffect(() => {
        const fetchStats = async () => {
            try {
                const stats = await apiClient.getToolFailures();
                const mapped = stats.map(s => ({
                    name: s.name,
                    service: s.serviceId,
                    failureRate: s.failureRate,
                    totalCalls: s.totalCalls
                }));
                setTools(mapped);
            } catch (error) {
                console.error("Failed to fetch tool failure rates", error);
            } finally {
                setLoading(false);
            }
        };

        fetchStats();
    }, []);

    return (
        <Card className="col-span-3 backdrop-blur-sm bg-background/50">
            <CardHeader className={cn(
                "flex flex-row items-center justify-between space-y-0",
                isCompact ? "p-3 pb-1" : "pb-2"
            )}>
                <CardTitle className="text-sm font-medium">Tool Failure Rates</CardTitle>
                <AlertCircle className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent className={isCompact ? "p-3 pt-0" : ""}>
                <div className={isCompact ? "space-y-2" : "space-y-4"}>
                    {tools.map((tool) => (
                        <div key={tool.name} className="space-y-1">
                            <div className="flex items-center justify-between text-xs">
                                <div className="flex items-center gap-2">
                                    <span className="font-medium truncate max-w-[120px]">{tool.name}</span>
                                    <Badge variant="outline" className="text-[10px] px-1 h-4">{tool.service}</Badge>
                                </div>
                                <span className={tool.failureRate > 15 ? "text-destructive font-bold" : "text-muted-foreground"}>
                                    {tool.failureRate.toFixed(1)}%
                                </span>
                            </div>
                            <Progress
                                value={tool.failureRate}
                                className="h-1"
                                // progress-color depending on rate? (Custom CSS or shadcn logic)
                            />
                        </div>
                    ))}
                    {tools.length === 0 && !loading && (
                        <div className="text-center py-4 text-xs text-muted-foreground">
                            No tool call data available.
                        </div>
                    )}
                </div>
            </CardContent>
        </Card>
    );
}
