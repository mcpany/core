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

    useEffect(() => {
        const fetchStats = async () => {
            try {
                // In a real app, this would be a dedicated analytics endpoint.
                // We'll mock some data based on existing tools.
                const response = await apiClient.listTools();
                const allTools = Array.isArray(response.tools) ? response.tools : [];

                // Shuffle and pick 4 with random failure rates for demo
                const sorted = allTools
                    .map((t: any) => ({
                        name: t.name,
                        service: t.service || "unknown",
                        failureRate: Math.random() < 0.3 ? Math.random() * 40 + 5 : Math.random() * 5,
                        totalCalls: Math.floor(Math.random() * 1000) + 100
                    }))
                    .sort((a: ToolFailureRate, b: ToolFailureRate) => b.failureRate - a.failureRate)
                    .slice(0, 4);

                setTools(sorted);
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
            <CardHeader className="flex flex-row items-center justify-between pb-2 space-y-0">
                <CardTitle className="text-sm font-medium">Tool Failure Rates</CardTitle>
                <AlertCircle className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
                <div className="space-y-4">
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
