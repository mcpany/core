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
 * Displays server uptime history over the last 24 hours based on real health checks.
 * @returns The rendered component.
 */
export function HealthHistoryChart() {
    const [data, setData] = useState<HealthPoint[]>([]);
    const [loading, setLoading] = useState(true);

    useEffect(() => {
        const fetchData = async () => {
            try {
                // Fetch real health history from the backend
                const response = await apiClient.getHealthHistory();

                // Aggregate all points from all services into a single timeline
                // response is { services: { "svc1": { points: [...] } } }
                const allPoints: { timestamp: number, status: string }[] = [];
                Object.values(response.services || {}).forEach((svc: any) => {
                    if (svc.points) {
                        allPoints.push(...svc.points.map((p: any) => ({
                            timestamp: Number(p.timestamp),
                            status: p.status
                        })));
                    }
                });

                if (allPoints.length === 0) {
                    // Fallback to empty state
                    setData([]);
                    return;
                }

                // Bucketize into 24 hourly bars
                const now = Date.now();
                const start = now - 24 * 60 * 60 * 1000;
                const bucketSize = 60 * 60 * 1000; // 1 hour
                const buckets: Record<number, { ok: number, total: number }> = {};

                // Initialize buckets
                // We use floor to snap to hour boundaries for cleaner UX?
                // Or relative to now? Relative to now ensures we cover exactly "last 24h".
                // Let's use relative to now.
                for (let i = 0; i < 24; i++) {
                    const bucketStart = start + i * bucketSize;
                    buckets[bucketStart] = { ok: 0, total: 0 };
                }

                allPoints.forEach(p => {
                    if (p.timestamp < start) return;

                    // Find bucket
                    // We iterate to find the bucket index
                    const offset = p.timestamp - start;
                    const bucketIndex = Math.floor(offset / bucketSize);

                    // Clamp to [0, 23]
                    if (bucketIndex >= 0 && bucketIndex < 24) {
                        const bucketStart = start + bucketIndex * bucketSize;
                        if (buckets[bucketStart]) {
                            buckets[bucketStart].total++;
                            // Case-insensitive check
                            if (p.status.toUpperCase() === "OK" || p.status.toUpperCase() === "HEALTHY") {
                                buckets[bucketStart].ok++;
                            }
                        }
                    }
                });

                const chartData = Object.entries(buckets).map(([timeStr, counts]) => {
                    const time = parseInt(timeStr);
                    // If no data points in that hour, assume it was fine (100%) or gap?
                    // GitHub style: gaps are empty.
                    // But for "uptime bar", gap usually means "no outage recorded".
                    // Let's assume 100 if seeded data should be there.
                    // If total is 0, we can skip or show 100?
                    // Let's show 100 to keep the bar full if it's "System Uptime".
                    // Wait, if no data, maybe the system was down/off?
                    // Let's show 0 or "unknown" if total is 0?
                    // Recharts needs a value.
                    // Let's go with: if total > 0 calculate, else 100 (optimistic) or 0 (pessimistic).
                    // Given we seed data, we should have data. If we don't, it might be before seeding.

                    const hasData = counts.total > 0;
                    const uptime = hasData ? (counts.ok / counts.total) * 100 : 100;

                    let status: "ok" | "degraded" | "error" = "ok";
                    if (uptime < 100) status = "degraded";
                    if (uptime < 90) status = "error";

                    // If no data, render as gray/offline?
                    // Let's allow "offline" status.
                    if (!hasData) status = "ok"; // Default to OK for visual continuity or use "offline"

                    return {
                        time: new Date(time).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' }),
                        uptime: Math.round(uptime),
                        status
                    };
                });

                // Sort by time (Object.entries might not guarantee order)
                chartData.sort((a, b) => {
                    // Re-parsing time string is hard.
                    // Better to rely on the fact that buckets keys are timestamps.
                    // But mapping lost the timestamp.
                    return 0; // We iterated in order, hopefully fine.
                    // Actually, Object.entries order is not guaranteed for numbers.
                    // Let's map directly from sorted keys.
                });

                // Proper sorting
                const sortedKeys = Object.keys(buckets).map(Number).sort((a, b) => a - b);
                const sortedChartData = sortedKeys.map(time => {
                     const counts = buckets[time];
                     const hasData = counts.total > 0;
                     const uptime = hasData ? (counts.ok / counts.total) * 100 : 100;
                     let status: HealthPoint["status"] = "ok";
                     if (uptime < 100) status = "degraded";
                     if (uptime < 90) status = "error";
                     if (!hasData) status = "offline"; // Use offline for gaps

                     // Map offline to 100 for height but gray color?
                     // Or 0 height?
                     // If 0 height, it's invisible.
                     // Let's use 100 height but gray color.
                     const displayUptime = hasData ? uptime : 100;

                     return {
                        time: new Date(time).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' }),
                        uptime: Math.round(displayUptime),
                        status
                     };
                });

                setData(sortedChartData);
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
        offline: "hsl(var(--muted))", // Gray for no data
        unknown: "hsl(var(--muted-foreground))",
    };

    const getBarColor = (status: HealthPoint["status"]) => {
        return STATUS_COLORS[status] || "hsl(var(--muted))";
    };

    if (loading) {
        return (
             <Card className="col-span-4 backdrop-blur-sm bg-background/50 h-[300px] animate-pulse">
                <CardHeader>
                    <div className="h-6 w-1/3 bg-muted rounded"></div>
                    <div className="h-4 w-2/3 bg-muted rounded mt-2"></div>
                </CardHeader>
                <CardContent className="flex items-center justify-center">
                    <div className="text-muted-foreground text-sm">Loading history...</div>
                </CardContent>
             </Card>
        );
    }

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
                                cursor={{ fill: 'var(--muted)', opacity: 0.1 }}
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
                                                        <span className="font-bold text-foreground">
                                                            {d.time}
                                                        </span>
                                                    </div>
                                                    <div className="flex flex-col">
                                                        <span className="text-[0.70rem] uppercase text-muted-foreground">
                                                            Status
                                                        </span>
                                                        <span className="font-bold capitalize" style={{ color: getBarColor(d.status) }}>
                                                            {d.status === 'offline' ? 'No Data' : `${d.uptime}%`}
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
                        24h History
                    </div>
                </div>
            </CardContent>
        </Card>
    );
}
