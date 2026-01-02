/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useMemo } from "react";
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
    CheckCircle2,
    Calendar,
    Download
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

// Mock Data Generators

const generateTimeData = (points: number) => {
    const data = [];
    const now = new Date();
    for (let i = 0; i < points; i++) {
        const time = new Date(now.getTime() - (points - i) * 60000);
        data.push({
            time: time.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' }),
            requests: Math.floor(Math.random() * 100) + 50,
            errors: Math.floor(Math.random() * 10),
            latency: Math.floor(Math.random() * 200) + 50
        });
    }
    return data;
};

const TOOL_USAGE_DATA = [
    { name: 'read_file', value: 400, color: '#3b82f6' },
    { name: 'list_files', value: 300, color: '#10b981' },
    { name: 'web_search', value: 300, color: '#f59e0b' },
    { name: 'execute_command', value: 200, color: '#8b5cf6' },
    { name: 'git_commit', value: 100, color: '#ef4444' },
];

const ERROR_DISTRIBUTION = [
    { name: 'Timeout', value: 45 },
    { name: 'Auth Failed', value: 25 },
    { name: 'Invalid Args', value: 20 },
    { name: 'Internal Error', value: 10 },
];

export function AnalyticsDashboard() {
    const [timeRange, setTimeRange] = useState("1h");
    const [activeTab, setActiveTab] = useState("overview");

    const trafficData = useMemo(() => generateTimeData(20), [timeRange]);

    const totalRequests = trafficData.reduce((acc, cur) => acc + cur.requests, 0);
    const avgLatency = Math.floor(trafficData.reduce((acc, cur) => acc + cur.latency, 0) / trafficData.length);
    const errorRate = (trafficData.reduce((acc, cur) => acc + cur.errors, 0) / totalRequests * 100).toFixed(2);

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
                            <SelectItem value="7d">Last 7 Days</SelectItem>
                            <SelectItem value="30d">Last 30 Days</SelectItem>
                        </SelectContent>
                    </Select>
                    <Button>
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
                                        +20.1% <ArrowUpRight className="h-3 w-3 ml-1" />
                                    </span>
                                    from last period
                                </p>
                            </CardContent>
                        </Card>
                        <Card>
                            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                                <CardTitle className="text-sm font-medium">Avg Latency</CardTitle>
                                <Clock className="h-4 w-4 text-muted-foreground" />
                            </CardHeader>
                            <CardContent>
                                <div className="text-2xl font-bold">{avgLatency}ms</div>
                                <p className="text-xs text-muted-foreground">
                                    <span className="text-rose-500 flex items-center">
                                        +4.5% <ArrowUpRight className="h-3 w-3 ml-1" />
                                    </span>
                                    slower than usual
                                </p>
                            </CardContent>
                        </Card>
                        <Card>
                            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                                <CardTitle className="text-sm font-medium">Error Rate</CardTitle>
                                <AlertTriangle className="h-4 w-4 text-muted-foreground" />
                            </CardHeader>
                            <CardContent>
                                <div className="text-2xl font-bold">{errorRate}%</div>
                                <p className="text-xs text-muted-foreground">
                                    <span className="text-emerald-500 flex items-center">
                                        -1.2% <ArrowDownRight className="h-3 w-3 ml-1" />
                                    </span>
                                    improvement
                                </p>
                            </CardContent>
                        </Card>
                        <Card>
                            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                                <CardTitle className="text-sm font-medium">Active Services</CardTitle>
                                <CheckCircle2 className="h-4 w-4 text-muted-foreground" />
                            </CardHeader>
                            <CardContent>
                                <div className="text-2xl font-bold">12</div>
                                <p className="text-xs text-muted-foreground">
                                    All systems operational
                                </p>
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
                                                    <stop offset="5%" stopColor="#3b82f6" stopOpacity={0.3} />
                                                    <stop offset="95%" stopColor="#3b82f6" stopOpacity={0} />
                                                </linearGradient>
                                            </defs>
                                            <XAxis
                                                dataKey="time"
                                                stroke="#888888"
                                                fontSize={12}
                                                tickLine={false}
                                                axisLine={false}
                                            />
                                            <YAxis
                                                stroke="#888888"
                                                fontSize={12}
                                                tickLine={false}
                                                axisLine={false}
                                                tickFormatter={(value) => `${value}`}
                                            />
                                            <Tooltip
                                                contentStyle={{ backgroundColor: 'hsl(var(--background))', border: '1px solid hsl(var(--border))' }}
                                                labelStyle={{ color: 'hsl(var(--foreground))' }}
                                            />
                                            <CartesianGrid strokeDasharray="3 3" strokeOpacity={0.1} vertical={false} />
                                            <Area
                                                type="monotone"
                                                dataKey="requests"
                                                stroke="#3b82f6"
                                                strokeWidth={2}
                                                fillOpacity={1}
                                                fill="url(#colorRequests)"
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
                                                data={TOOL_USAGE_DATA}
                                                cx="50%"
                                                cy="50%"
                                                innerRadius={60}
                                                outerRadius={80}
                                                paddingAngle={5}
                                                dataKey="value"
                                            >
                                                {TOOL_USAGE_DATA.map((entry, index) => (
                                                    <Cell key={`cell-${index}`} fill={entry.color} />
                                                ))}
                                            </Pie>
                                            <Tooltip
                                                contentStyle={{ backgroundColor: 'hsl(var(--background))', border: '1px solid hsl(var(--border))' }}
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
                                            stroke="#888888"
                                            fontSize={12}
                                            tickLine={false}
                                            axisLine={false}
                                        />
                                        <YAxis
                                            stroke="#888888"
                                            fontSize={12}
                                            tickLine={false}
                                            axisLine={false}
                                            unit="ms"
                                        />
                                        <Tooltip
                                            cursor={{fill: 'transparent'}}
                                            contentStyle={{ backgroundColor: 'hsl(var(--background))', border: '1px solid hsl(var(--border))' }}
                                        />
                                        <Bar dataKey="latency" fill="#f59e0b" radius={[4, 4, 0, 0]} />
                                    </BarChart>
                                </ResponsiveContainer>
                            </div>
                        </CardContent>
                     </Card>
                 </TabsContent>

                 <TabsContent value="errors" className="space-y-4">
                     <div className="grid gap-4 md:grid-cols-2">
                        <Card>
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
                                                stroke="#888888"
                                                fontSize={12}
                                                tickLine={false}
                                                axisLine={false}
                                            />
                                            <YAxis
                                                stroke="#888888"
                                                fontSize={12}
                                                tickLine={false}
                                                axisLine={false}
                                            />
                                            <Tooltip
                                                contentStyle={{ backgroundColor: 'hsl(var(--background))', border: '1px solid hsl(var(--border))' }}
                                            />
                                            <Line type="monotone" dataKey="errors" stroke="#ef4444" strokeWidth={2} dot={false} />
                                        </LineChart>
                                    </ResponsiveContainer>
                                </div>
                            </CardContent>
                        </Card>
                        <Card>
                            <CardHeader>
                                <CardTitle>Error Types</CardTitle>
                                <CardDescription>Categorization of system errors.</CardDescription>
                            </CardHeader>
                            <CardContent>
                                <div className="space-y-4">
                                    {ERROR_DISTRIBUTION.map((err) => (
                                        <div key={err.name} className="flex items-center">
                                            <div className="w-[100px] text-sm font-medium">{err.name}</div>
                                            <div className="flex-1 h-2 bg-muted rounded-full overflow-hidden">
                                                <div
                                                    className="h-full bg-destructive"
                                                    style={{ width: `${err.value}%` }}
                                                />
                                            </div>
                                            <div className="w-[50px] text-right text-sm text-muted-foreground">{err.value}%</div>
                                        </div>
                                    ))}
                                </div>
                            </CardContent>
                        </Card>
                     </div>
                 </TabsContent>
            </Tabs>
        </div>
    );
}
