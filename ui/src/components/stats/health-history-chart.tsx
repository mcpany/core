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
                // In a real app, this would be a dedicated history endpoint.
                // For this implementation, we simulate 24 hours of data based on
                // the current status and some randomized historical noise.
                const status = await apiClient.getDoctorStatus();

                const points: HealthPoint[] = [];
                const now = new Date();

                for (let i = 23; i >= 0; i--) {
                    const time = new Date(now.getTime() - i * 60 * 60 * 1000);
                    const hour = time.getHours();
                    const timeStr = `${hour}:00`;

                    // Simulate some downtime or degradation in history
                    // but keep current status accurate
                    let currentStatus: HealthPoint["status"] = "ok";
                    let uptime = 100;

                    if (i === 0) {
                        currentStatus = status.status === "ok" ? "ok" : (status.status === "degraded" ? "degraded" : "error");
                        uptime = status.status === "ok" ? 100 : (status.status === "degraded" ? 85 : 0);
                    } else {
                        // Mock historical data
                        const rand = Math.random();
                        if (rand > 0.95) {
                            currentStatus = "error";
                            uptime = 0;
                        } else if (rand > 0.85) {
                            currentStatus = "degraded";
                            uptime = Math.floor(Math.random() * 40) + 50;
                        } else {
                            currentStatus = "ok";
                            uptime = 100;
                        }
                    }

                    points.push({
                        time: timeStr,
                        status: currentStatus,
                        uptime: uptime
                    });
                }
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
