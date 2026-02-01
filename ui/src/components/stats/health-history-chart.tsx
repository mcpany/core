/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useEffect } from "react";
import { Bar, BarChart, ResponsiveContainer, Tooltip, XAxis, YAxis, Cell } from "recharts";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { apiClient } from "@/lib/client";

interface HealthPoint {
    time: string;
    status: "ok" | "degraded" | "error" | "offline";
    uptime: number; // 0 to 100
}

/**
 * HealthHistoryChart component.
 * Displays server uptime history over the last 24 hours.
 * @returns The rendered component.
 */
export function HealthHistoryChart() {
    const [data, setData] = useState<HealthPoint[]>([]);
    const [loading, setLoading] = useState(true);

    useEffect(() => {
        const fetchData = async () => {
            try {
                const response = await apiClient.getDashboardHealth();

                // Try to get "system" history, or fallback to aggregating all services
                let history = response.history["system"];

                if (!history && Object.keys(response.history).length > 0) {
                     // Fallback: Use the history of the first service? Or aggregate?
                     // For now, let's just use the first available one to show *something*
                     const firstKey = Object.keys(response.history)[0];
                     history = response.history[firstKey];
                }

                if (history && history.length > 0) {
                     const points: HealthPoint[] = history.map(h => {
                        const date = new Date(h.timestamp);
                        // Format time as HH:MM
                        const timeStr = date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });

                        let uptime = 100;
                        let status: HealthPoint["status"] = "ok";

                        // Map backend status to UI status/uptime
                        // backend: "up", "down", "ok", "error", etc.
                        const s = h.status.toLowerCase();
                        if (s === 'down' || s === 'error' || s === 'offline') {
                            status = "error";
                            uptime = 0;
                        } else if (s === 'degraded' || s === 'unhealthy') {
                            status = "degraded";
                            uptime = 80;
                        }

                        return {
                            time: timeStr,
                            status: status,
                            uptime: uptime
                        };
                     });
                     setData(points);
                } else {
                     setData([]);
                }
            } catch (error) {
                console.error("Failed to fetch health history", error);
            } finally {
                setLoading(false);
            }
        };

        fetchData();
    }, []);

    const STATUS_COLORS = {
        healthy: "hsl(var(--chart-2))",
        ok: "hsl(var(--chart-2))",
        degraded: "hsl(var(--chart-4))",
        error: "hsl(var(--chart-1))",
        offline: "hsl(var(--muted-foreground))",
        unknown: "hsl(var(--muted-foreground))",
    };

    const getBarColor = (status: HealthPoint["status"]) => {
        return STATUS_COLORS[status] || "hsl(var(--muted))";
    };

    return (
        <Card className="col-span-4 backdrop-blur-sm bg-background/50">
            <CardHeader>
                <CardTitle>System Uptime</CardTitle>
                <CardDescription>
                    Availability and health status over the last 24 hours.
                </CardDescription>
            </CardHeader>
            <CardContent>
                <div className="h-[200px] w-full">
                    <ResponsiveContainer width="100%" height="100%">
                        <BarChart data={data}>
                            <XAxis
                                dataKey="time"
                                stroke="#888888"
                                fontSize={10}
                                tickLine={false}
                                axisLine={false}
                                interval="preserveStartEnd"
                            />
                            <YAxis hide domain={[0, 100]} />
                            <Tooltip
                                content={({ active, payload }) => {
                                    if (active && payload && payload.length) {
                                        const d = payload[0].payload as HealthPoint;
                                        return (
                                            <div className="rounded-lg border bg-background p-2 shadow-sm">
                                                <div className="grid grid-cols-2 gap-2">
                                                    <div className="flex flex-col">
                                                        <span className="text-[0.70rem] uppercase text-muted-foreground">
                                                            Time
                                                        </span>
                                                        <span className="font-bold text-muted-foreground">
                                                            {d.time}
                                                        </span>
                                                    </div>
                                                    <div className="flex flex-col">
                                                        <span className="text-[0.70rem] uppercase text-muted-foreground">
                                                            Uptime
                                                        </span>
                                                        <span className="font-bold" style={{ color: getBarColor(d.status) }}>
                                                            {d.uptime}%
                                                        </span>
                                                    </div>
                                                </div>
                                            </div>
                                        );
                                    }
                                    return null;
                                }}
                            />
                            <Bar dataKey="uptime" radius={[2, 2, 0, 0]}>
                                {data.map((entry, index) => (
                                    <Cell key={`cell-${index}`} fill={getBarColor(entry.status)} />
                                ))}
                            </Bar>
                        </BarChart>
                    </ResponsiveContainer>
                </div>
                <div className="mt-4 flex items-center justify-between text-xs text-muted-foreground">
                    <div className="flex items-center gap-2">
                        <div className="h-2 w-2 rounded-full bg-[hsl(var(--chart-2))]" />
                        <span>Operational</span>
                    </div>
                    <div className="flex items-center gap-2">
                        <div className="h-2 w-2 rounded-full bg-[hsl(var(--chart-4))]" />
                        <span>Degraded</span>
                    </div>
                    <div className="flex items-center gap-2">
                        <div className="h-2 w-2 rounded-full bg-[hsl(var(--chart-1))]" />
                        <span>Down</span>
                    </div>
                    <div className="font-medium text-foreground">
                        {data.length > 0 ? (
                             (data.filter(d => d.status === 'ok').length / data.length * 100).toFixed(1)
                        ) : '0.0'}% Overall Uptime
                    </div>
                </div>
            </CardContent>
        </Card>
    );
}
