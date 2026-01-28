/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useEffect, useMemo } from "react";
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
                const history = await apiClient.getHealthHistory();

                const points: HealthPoint[] = history.map((h: any) => {
                    const uptime = h.uptimePercentage !== undefined ? h.uptimePercentage : (h.uptime_percentage || 0);
                    let status: HealthPoint["status"] = "ok";
                    if (uptime < 90) status = "degraded";
                    if (uptime < 50) status = "error";

                    // Format timestamp to HH:mm
                    const date = new Date(h.timestamp);
                    const time = date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });

                    return {
                        time: time,
                        status: status,
                        uptime: Math.round(uptime)
                    };
                });
                setData(points);
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
        offline: "hsl(var(--muted))",
        unknown: "hsl(var(--muted))",
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
                                interval={3}
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
                        99.9% Overall Uptime
                    </div>
                </div>
            </CardContent>
        </Card>
    );
}
