/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useEffect, useMemo } from "react";
import {
    Area,
    AreaChart,
    Bar,
    BarChart,
    CartesianGrid,
    Cell,
    Line,
    LineChart,
    Pie,
    PieChart,
    ResponsiveContainer,
    Tooltip,
    XAxis,
    YAxis,
    Legend
} from "recharts";
import {
    ArrowDownRight,
    ArrowUpRight,
    Activity,
    Clock,
    AlertTriangle,
    Calendar,
} from "lucide-react";
import {
    Card,
    CardContent,
    CardDescription,
    CardHeader,
    CardTitle,
} from "@/components/ui/card";
import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from "@/components/ui/select";
import { Button } from "@/components/ui/button";

import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { apiClient } from "@/lib/client";

// Tool usage colors
const COLORS = ['#3b82f6', '#10b981', '#f59e0b', '#8b5cf6', '#ef4444', '#ec4899', '#6366f1'];

/**
 * AnalyticsDashboard component.
 * @returns The rendered component.
 */
export function AnalyticsDashboard() {
    const [timeRange, setTimeRange] = useState("1h");
    const [activeTab, setActiveTab] = useState("overview");

    const [trafficData, setTrafficData] = useState<any[]>([]);
    const [toolUsageData, setToolUsageData] = useState<any[]>([]);
    const [isMounted, setIsMounted] = useState(false);

    useEffect(() => {
        setIsMounted(true);
        const fetchDashboardData = async () => {
            try {
                const [traffic, tools] = await Promise.all([
                    apiClient.getDashboardTraffic(),
                    apiClient.getTopTools()
                ]);
                setTrafficData(traffic || []);

                // Format tool usage data
                const formattedTools = (tools || []).map((t: any, index: number) => ({
                    name: t.name,
                    value: t.count,
                    color: COLORS[index % COLORS.length]
                }));
                setToolUsageData(formattedTools);
            } catch (error) {
                console.error("Failed to fetch dashboard data", error);
            }
        };

        fetchDashboardData();
        const interval = setInterval(fetchDashboardData, 30000);
        return () => clearInterval(interval);
    }, [timeRange]);

    // ⚡ Bolt Optimization: Memoize statistics calculations to prevent unnecessary re-computations
    // when switching tabs or when other unrelated state changes.
    const stats = useMemo(() => {
        const totalRequests = trafficData.reduce((acc, cur) => acc + (cur.requests || cur.total || 0), 0);
        const avgLatency = trafficData.length
            ? Math.floor(trafficData.reduce((acc, cur) => acc + (cur.latency || 0), 0) / trafficData.length)
            : 0;
        const errorCount = trafficData.reduce((acc, cur) => acc + (cur.errors || 0), 0);
        const errorRate = totalRequests ? ((errorCount / totalRequests) * 100).toFixed(2) : "0.00";
        // Assuming 1 minute per data point for "rps" calculation if we have enough points, otherwise just total
        const durationMinutes = trafficData.length;
        const avgRps = (durationMinutes && totalRequests) ? (totalRequests / (durationMinutes * 60)).toFixed(2) : "0.00";

        return { totalRequests, avgLatency, errorRate, avgRps };
    }, [trafficData]);

    const { totalRequests, avgLatency, errorRate, avgRps } = stats;

    if (!isMounted) return null;

    return (
        <div className="flex-1 space-y-4 p-8 pt-6 h-full overflow-y-auto">
            <div className="flex items-center justify-between space-y-2">
                <div>
                    <h2 className="text-3xl font-bold tracking-tight">Analytics & Stats</h2>
                    <p className="text-muted-foreground">Real-time insights into your MCP infrastructure.</p>
                </div>
                <div className="flex items-center space-x-2">
                    <Select value={timeRange} onValueChange={setTimeRange}>
                        <SelectTrigger className="w-[180px]">
                            <Calendar className="mr-2 h-4 w-4" />
                            <SelectValue placeholder="Select range" />
                        </SelectTrigger>
                        <SelectContent>
                            <SelectItem value="1h">Last 1 Hour</SelectItem>
                            <SelectItem value="24h">Last 24 Hours</SelectItem>
                        </SelectContent>
                    </Select>
                    <Button disabled>
                        <ArrowDownRight className="mr-2 h-4 w-4" /> Export Report
                    </Button>
                </div>
            </div>

            <Tabs defaultValue="overview" value={activeTab} onValueChange={setActiveTab} className="space-y-4">
                <TabsList>
                    <TabsTrigger value="overview">Overview</TabsTrigger>
                    <TabsTrigger value="performance">Performance</TabsTrigger>
                    <TabsTrigger value="errors">Errors</TabsTrigger>
                </TabsList>

                <TabsContent value="overview" className="space-y-4">
                    <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
                        <Card>
                            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                                <CardTitle className="text-sm font-medium">Total Requests</CardTitle>
                                <Activity className="h-4 w-4 text-muted-foreground" />
                            </CardHeader>
                            <CardContent>
                                <div className="text-2xl font-bold">{totalRequests.toLocaleString()}</div>
                                <p className="text-xs text-muted-foreground">
                                    <span className="text-emerald-500 flex items-center">
                                       <Activity className="h-3 w-3 mr-1" /> Live
                                    </span>
                                </p>
                            </CardContent>
                        </Card>
                        <Card>
                            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                                <CardTitle className="text-sm font-medium">Avg Throughput</CardTitle>
                                <Activity className="h-4 w-4 text-muted-foreground" />
                            </CardHeader>
                            <CardContent>
                                <div className="text-2xl font-bold">{avgRps} rps</div>
                            </CardContent>
                        </Card>
                        <Card>
                            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                                <CardTitle className="text-sm font-medium">Avg Latency</CardTitle>
                                <Clock className="h-4 w-4 text-muted-foreground" />
                            </CardHeader>
                            <CardContent>
                                <div className="text-2xl font-bold">{avgLatency}ms</div>
                            </CardContent>
                        </Card>
                        <Card>
                            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                                <CardTitle className="text-sm font-medium">Error Rate</CardTitle>
                                <AlertTriangle className="h-4 w-4 text-muted-foreground" />
                            </CardHeader>
                            <CardContent>
                                <div className="text-2xl font-bold">{errorRate}%</div>
                            </CardContent>
                        </Card>
                    </div>

                    <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-7">
                        <Card className="col-span-4">
                            <CardHeader>
                                <CardTitle>Request Volume</CardTitle>
                                <CardDescription>
                                    Traffic across all MCP services over the last period.
                                </CardDescription>
                            </CardHeader>
                            <CardContent className="pl-2">
                                <div className="h-[300px]">
                                    <ResponsiveContainer width="100%" height="100%">
                                        <AreaChart data={trafficData}>
                                            <defs>
                                                <linearGradient id="colorRequests" x1="0" y1="0" x2="0" y2="1">
                                                    <stop offset="5%" stopColor="hsl(var(--primary))" stopOpacity={0.3} />
                                                    <stop offset="95%" stopColor="hsl(var(--primary))" stopOpacity={0} />
                                                </linearGradient>
                                            </defs>
                                            <XAxis
                                                dataKey="time"
                                                stroke="hsl(var(--muted-foreground))"
                                                fontSize={12}
                                                tickLine={false}
                                                axisLine={false}
                                            />
                                            <YAxis
                                                stroke="hsl(var(--muted-foreground))"
                                                fontSize={12}
                                                tickLine={false}
                                                axisLine={false}
                                                tickFormatter={(value) => `${value}`}
                                            />
                                            <Tooltip
                                                contentStyle={{ backgroundColor: 'hsl(var(--card))', border: '1px solid hsl(var(--border))', color: 'hsl(var(--foreground))' }}
                                                labelStyle={{ color: 'hsl(var(--foreground))' }}
                                            />
                                            <CartesianGrid strokeDasharray="3 3" strokeOpacity={0.1} vertical={false} />
                                            <Area
                                                type="monotone"
                                                dataKey="requests"
                                                stroke="hsl(var(--primary))"
                                                fillOpacity={1}
                                                fill="url(#colorRequests)"
                                                isAnimationActive={false}
                                            />
                                        </AreaChart>
                                    </ResponsiveContainer>
                                </div>
                            </CardContent>
                        </Card>
                        <Card className="col-span-3">
                            <CardHeader>
                                <CardTitle>Tool Usage Distribution</CardTitle>
                                <CardDescription>
                                    Most frequently called tools.
                                </CardDescription>
                            </CardHeader>
                            <CardContent>
                                <div className="h-[300px]">
                                    <ResponsiveContainer width="100%" height="100%">
                                        <PieChart>
                                            <Pie
                                                data={toolUsageData}
                                                cx="50%"
                                                cy="50%"
                                                innerRadius={60}
                                                outerRadius={80}
                                                dataKey="value"
                                                isAnimationActive={false}
                                            >
                                                {toolUsageData.map((entry, index) => (
                                                    <Cell key={`cell-${index}`} fill={entry.color} />
                                                ))}
                                            </Pie>
                                            <Tooltip
                                                contentStyle={{ backgroundColor: 'hsl(var(--card))', border: '1px solid hsl(var(--border))' }}
                                                itemStyle={{ color: 'hsl(var(--foreground))' }}
                                            />
                                            <Legend />
                                        </PieChart>
                                    </ResponsiveContainer>
                                </div>
                            </CardContent>
                        </Card>
                    </div>
                </TabsContent>

                 <TabsContent value="performance" className="space-y-4">
                     <Card>
                        <CardHeader>
                            <CardTitle>Latency Trend</CardTitle>
                            <CardDescription>Average response time per request.</CardDescription>
                        </CardHeader>
                        <CardContent className="pl-2">
                             <div className="h-[350px]">
                                <ResponsiveContainer width="100%" height="100%">
                                    <BarChart data={trafficData}>
                                        <XAxis
                                            dataKey="time"
                                            stroke="hsl(var(--muted-foreground))"
                                            fontSize={12}
                                            tickLine={false}
                                            axisLine={false}
                                        />
                                        <YAxis
                                            stroke="hsl(var(--muted-foreground))"
                                            fontSize={12}
                                            tickLine={false}
                                            axisLine={false}
                                            unit="ms"
                                        />
                                        <Tooltip
                                            cursor={{fill: 'transparent'}}
                                            contentStyle={{ backgroundColor: 'hsl(var(--card))', border: '1px solid hsl(var(--border))' }}
                                            labelStyle={{ color: 'hsl(var(--foreground))' }}
                                        />
                                        <Bar dataKey="latency" fill="hsl(var(--primary))" radius={[4, 4, 0, 0]} isAnimationActive={false} />
                                    </BarChart>
                                </ResponsiveContainer>
                            </div>
                        </CardContent>
                     </Card>
                 </TabsContent>

                 <TabsContent value="errors" className="space-y-4">
                     <div className="grid gap-4 md:grid-cols-2">
                        <Card className="col-span-2">
                            <CardHeader>
                                <CardTitle>Error Trend</CardTitle>
                                <CardDescription>Number of failed requests over time.</CardDescription>
                            </CardHeader>
                            <CardContent className="pl-2">
                                 <div className="h-[300px]">
                                    <ResponsiveContainer width="100%" height="100%">
                                        <LineChart data={trafficData}>
                                            <XAxis
                                                dataKey="time"
                                                stroke="hsl(var(--muted-foreground))"
                                                fontSize={12}
                                                tickLine={false}
                                                axisLine={false}
                                            />
                                            <YAxis
                                                stroke="hsl(var(--muted-foreground))"
                                                fontSize={12}
                                                tickLine={false}
                                                axisLine={false}
                                            />
                                            <Tooltip
                                                contentStyle={{ backgroundColor: 'hsl(var(--card))', border: '1px solid hsl(var(--border))' }}
                                                labelStyle={{ color: 'hsl(var(--foreground))' }}
                                            />
                                            <Line type="monotone" dataKey="errors" stroke="hsl(var(--destructive))" strokeWidth={2} dot={false} isAnimationActive={false} />
                                        </LineChart>
                                    </ResponsiveContainer>
                                </div>
                            </CardContent>
                        </Card>
                     </div>
                 </TabsContent>
            </Tabs>
        </div>
    );
}
