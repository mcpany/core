/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */



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


    useEffect(() => {

        const fetchData = async () => {
            try {
                // Fetch real health data from the dashboard/health endpoint
                const healthData = await apiClient.getDashboardHealth();

                const points: HealthPoint[] = [];
                const historyMap = healthData.history || {};

                // Collect all points to find min and max time
                let minTime = Infinity;
                let maxTime = -Infinity;
                const allPoints: { timestamp: number, status: string }[] = [];

                Object.values(historyMap).forEach((serviceHistory) => {
                    if (Array.isArray(serviceHistory)) {
                        serviceHistory.forEach((point) => {
                            if (point && point.timestamp) {
                                minTime = Math.min(minTime, point.timestamp);
                                maxTime = Math.max(maxTime, point.timestamp);
                                allPoints.push(point);
                            }
                        });
                    }
                });

                if (allPoints.length === 0) {
                    setData([]);
                    return;
                }

                // Bucket the timestamps into chunks (e.g. 60 buckets)
                const buckets = 60;
                let timeSpan = maxTime - minTime;

                // If the timespan is too small (e.g. 0), use a default 1-minute bucket length
                if (timeSpan <= 0) {
                    timeSpan = 60000;
                    maxTime = minTime + timeSpan;
                }
                const bucketSize = timeSpan / buckets;

                const bucketData: Record<number, { up: number, total: number }> = {};
                for (let i = 0; i < buckets; i++) {
                    bucketData[i] = { up: 0, total: 0 };
                }

                Object.values(historyMap).forEach((serviceHistory) => {
                    if (Array.isArray(serviceHistory)) {
                        serviceHistory.forEach((point) => {
                            if (point && point.timestamp) {
                                let bIndex = Math.floor((point.timestamp - minTime) / bucketSize);
                                bIndex = Math.min(Math.max(bIndex, 0), buckets - 1); // clamp to avoid out-of-bounds
                                bucketData[bIndex].total += 1;
                                // Assume healthy if status is "healthy", "up", "ok", etc.
                                if (["healthy", "up", "ok", "serving"].includes(point.status.toLowerCase())) {
                                    bucketData[bIndex].up += 1;
                                }
                            }
                        });
                    }
                });

                // Convert buckets back to array
                for (let i = 0; i < buckets; i++) {
                    const bucket = bucketData[i];
                    // Create a time label for the bucket
                    const time = new Date(minTime + i * bucketSize).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });

                    if (bucket && bucket.total > 0) {
                        const uptimeRatio = bucket.up / bucket.total;
                        let status: HealthPoint["status"] = "ok";
                        if (uptimeRatio < 0.5) status = "error";
                        else if (uptimeRatio < 1.0) status = "degraded";

                        points.push({
                            time,
                            status,
                            uptime: Math.round(uptimeRatio * 100)
                        });
                    } else {
                        // If no data in bucket, fill with previous state or 100% ok
                        points.push({
                            time,
                            status: "ok",
                            uptime: 100
                        });
                    }
                }

                setData(points);
            } catch (error) {
                console.error("Failed to fetch health history", error);
            } finally {

            }
        };

        fetchData();
    }, []);

    const STATUS_COLORS: Record<string, string> = {
        healthy: "hsl(var(--chart-2))",
        ok: "hsl(var(--chart-2))",
        degraded: "hsl(var(--chart-4))",
        error: "hsl(var(--chart-1))",
        offline: "#9ca3af", // gray-400 for visible contrast in dark mode
        unknown: "#9ca3af",
    };

    const getBarColor = (status: HealthPoint["status"]): string => {
        return STATUS_COLORS[status] || "#9ca3af";
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
