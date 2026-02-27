/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useEffect } from "react";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Progress } from "@/components/ui/progress";
import { AlertCircle } from "lucide-react";
import { useDashboard } from "@/components/dashboard/dashboard-context";

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
    const { serviceId } = useDashboard();

    useEffect(() => {
        const fetchStats = async () => {
            try {
                let url = '/api/dashboard/tool-failures';
                if (serviceId) {
                    url += `?serviceId=${encodeURIComponent(serviceId)}`;
                }
                const res = await fetch(url);
                if (res.ok) {
                    const stats = await res.json();
                    const mapped: ToolFailureRate[] = stats.map((s: { name: string; serviceId: string; failureRate: number; totalCalls: number }) => ({
                        name: s.name,
                        service: s.serviceId, // Map serviceId from API to service for UI
                        failureRate: s.failureRate,
                        totalCalls: s.totalCalls
                    }));
                    setTools(mapped);
                } else {
                    console.error("Failed to fetch tool failure rates");
                }
            } catch (error) {
                console.error("Failed to fetch tool failure rates", error);
            } finally {
                setLoading(false);
            }
        };

        fetchStats();
    }, [serviceId]);

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
