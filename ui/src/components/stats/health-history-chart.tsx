/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useEffect } from "react";
import { Bar, BarChart, ResponsiveContainer, Tooltip, XAxis, YAxis, Cell } from "recharts";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { apiClient } from "@/lib/client";

export interface TrafficPoint {
    time: string;
    requests: number;
    errors: number;
    total?: number; // fallback
}

interface HealthPoint {
    time: string;
    status: "ok" | "degraded" | "error" | "offline";
    uptime: number; // Availability %
    requests: number;
    errors: number;
}

/**
 * HealthHistoryChart component.
 * Displays server traffic and health status over the last hour.
 * @returns The rendered component.
 */
export function HealthHistoryChart() {
    const [data, setData] = useState<HealthPoint[]>([]);
    const [loading, setLoading] = useState(true);

    const fetchData = async () => {
        try {
            const traffic: TrafficPoint[] = await apiClient.getDashboardTraffic();

            const points: HealthPoint[] = [];

            if (traffic && traffic.length > 0) {
                 for (const t of traffic) {
                    let pointStatus: HealthPoint["status"] = "ok";

                    const reqs = t.requests || t.total || 0;
                    const errs = t.errors || 0;

                    let availability = 100;
                    if (reqs > 0) {
                        availability = Math.max(0, ((reqs - errs) / reqs) * 100);
                    } else {
                        // No traffic = 100% availability (technically not down) or special status?
                        // For visualization, if no traffic, we might want to show empty or gray.
                        // Let's keep it 100% but maybe use status 'offline' or 'idle' for color.
                        pointStatus = "offline";
                    }

                    if (reqs > 0) {
                        if (availability < 90) {
                            pointStatus = "error";
                        } else if (availability < 99) {
                            pointStatus = "degraded";
                        }
                    }

                    points.push({
                        time: t.time,
                        status: pointStatus,
                        uptime: availability,
                        requests: reqs,
                        errors: errs
                    });
                 }
            }
            setData(points);
        } catch (error) {
            console.error("Failed to fetch health history", error);
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        fetchData();
        // Poll every 30 seconds for real-time updates
        const interval = setInterval(fetchData, 30000);
        return () => clearInterval(interval);
    }, []);

    const STATUS_COLORS = {
        healthy: "hsl(var(--chart-2))",
        ok: "hsl(var(--chart-2))", // Green
        degraded: "hsl(var(--chart-4))", // Yellow/Orange
        error: "hsl(var(--chart-1))", // Red
        offline: "hsl(var(--muted))", // Gray
        unknown: "hsl(var(--muted))",
    };

    const getBarColor = (status: HealthPoint["status"]) => {
        return STATUS_COLORS[status] || "hsl(var(--muted))";
    };

    return (
        <Card className="col-span-4 backdrop-blur-sm bg-background/50">
            <CardHeader>
                <CardTitle>Traffic & Health (Last Hour)</CardTitle>
                <CardDescription>
                    Availability based on request success rate.
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
                                                            Availability
                                                        </span>
                                                        <span className="font-bold" style={{ color: getBarColor(d.status) }}>
                                                            {d.uptime.toFixed(1)}%
                                                        </span>
                                                    </div>
                                                    <div className="flex flex-col">
                                                        <span className="text-[0.70rem] uppercase text-muted-foreground">
                                                            Requests
                                                        </span>
                                                        <span className="font-bold text-foreground">
                                                            {d.requests}
                                                        </span>
                                                    </div>
                                                    <div className="flex flex-col">
                                                        <span className="text-[0.70rem] uppercase text-muted-foreground">
                                                            Errors
                                                        </span>
                                                        <span className="font-bold text-destructive">
                                                            {d.errors}
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
                        <span>Healthy (&gt;99%)</span>
                    </div>
                    <div className="flex items-center gap-2">
                        <div className="h-2 w-2 rounded-full bg-[hsl(var(--chart-4))]" />
                        <span>Degraded (&gt;90%)</span>
                    </div>
                    <div className="flex items-center gap-2">
                        <div className="h-2 w-2 rounded-full bg-[hsl(var(--chart-1))]" />
                        <span>Critical (&lt;90%)</span>
                    </div>
                    {/* Calculate overall uptime from displayed data */}
                    <div className="font-medium text-foreground">
                         {(() => {
                             if (data.length === 0) return "No Data";
                             // Weighted average? Or just average of averages?
                             // Weighted by requests is better for "Success Rate".
                             const totalReqs = data.reduce((acc, p) => acc + p.requests, 0);
                             const totalErrs = data.reduce((acc, p) => acc + p.errors, 0);
                             if (totalReqs === 0) return "100% Overall";
                             const overall = ((totalReqs - totalErrs) / totalReqs) * 100;
                             return `${overall.toFixed(1)}% Success Rate`;
                         })()}
                    </div>
                </div>
            </CardContent>
        </Card>
    );
}
