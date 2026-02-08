/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useEffect } from "react";
import { Area, AreaChart, Line, LineChart, Bar, BarChart, ResponsiveContainer, Tooltip, XAxis, YAxis, CartesianGrid, Legend } from "recharts";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { apiClient } from "@/lib/client";
import { Skeleton } from "@/components/ui/skeleton";
import { AlertCircle, Activity, Clock } from "lucide-react";

interface TrafficPoint {
    time: string;
    requests: number;
    errors: number;
    latency: number;
}

/**
 * ServiceMetrics component.
 * Displays detailed metrics for a specific service including traffic, errors, and latency.
 * @param props - The component props.
 * @param props.serviceId - The ID of the service to display metrics for.
 * @returns The rendered component.
 */
export function ServiceMetrics({ serviceId }: { serviceId: string }) {
  const [data, setData] = useState<TrafficPoint[]>([]);
  const [loading, setLoading] = useState(true);
  const [mounted, setMounted] = useState(false);

  useEffect(() => {
    setMounted(true);
    const fetchData = async () => {
        try {
            const traffic = await apiClient.getDashboardTraffic(serviceId);
            setData(traffic);
        } catch (error) {
            console.error("Failed to fetch service metrics", error);
        } finally {
            setLoading(false);
        }
    };
    fetchData();
    // Poll every 30 seconds
    const interval = setInterval(fetchData, 30000);
    return () => clearInterval(interval);
  }, [serviceId]);

  if (!mounted) return null;

  if (loading && data.length === 0) {
      return (
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <Skeleton className="h-[300px] w-full" />
              <Skeleton className="h-[300px] w-full" />
              <Skeleton className="h-[300px] w-full col-span-1 md:col-span-2" />
          </div>
      );
  }

  if (data.length === 0) {
      return (
          <div className="flex flex-col items-center justify-center p-12 text-muted-foreground bg-muted/10 rounded-lg border border-dashed">
              <Activity className="h-12 w-12 opacity-20 mb-4" />
              <p>No metrics data available for this service yet.</p>
          </div>
      );
  }

  return (
    <div className="space-y-6">
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            {/* Request Volume */}
            <Card className="backdrop-blur-sm bg-background/50">
                <CardHeader>
                    <CardTitle className="text-base flex items-center gap-2">
                        <Activity className="h-4 w-4 text-primary" />
                        Request Traffic
                    </CardTitle>
                    <CardDescription>Requests per minute over the last hour.</CardDescription>
                </CardHeader>
                <CardContent className="pl-0">
                    <div className="h-[250px] w-full">
                        <ResponsiveContainer width="100%" height="100%">
                            <AreaChart data={data}>
                                <defs>
                                    <linearGradient id="colorRequests" x1="0" y1="0" x2="0" y2="1">
                                        <stop offset="5%" stopColor="#8884d8" stopOpacity={0.8}/>
                                        <stop offset="95%" stopColor="#8884d8" stopOpacity={0}/>
                                    </linearGradient>
                                </defs>
                                <XAxis dataKey="time" stroke="#888888" fontSize={10} tickLine={false} axisLine={false} />
                                <YAxis stroke="#888888" fontSize={10} tickLine={false} axisLine={false} />
                                <CartesianGrid strokeDasharray="3 3" vertical={false} stroke="hsl(var(--muted))" />
                                <Tooltip
                                    contentStyle={{ borderRadius: '8px', border: '1px solid hsl(var(--border))', boxShadow: '0 4px 6px -1px rgb(0 0 0 / 0.1)' }}
                                    labelStyle={{ color: 'hsl(var(--foreground))' }}
                                />
                                <Area type="monotone" dataKey="requests" stroke="#8884d8" fillOpacity={1} fill="url(#colorRequests)" name="Requests" />
                            </AreaChart>
                        </ResponsiveContainer>
                    </div>
                </CardContent>
            </Card>

            {/* Latency */}
            <Card className="backdrop-blur-sm bg-background/50">
                <CardHeader>
                    <CardTitle className="text-base flex items-center gap-2">
                        <Clock className="h-4 w-4 text-blue-500" />
                        Avg Latency
                    </CardTitle>
                    <CardDescription>Average response time (ms).</CardDescription>
                </CardHeader>
                <CardContent className="pl-0">
                    <div className="h-[250px] w-full">
                        <ResponsiveContainer width="100%" height="100%">
                            <LineChart data={data}>
                                <XAxis dataKey="time" stroke="#888888" fontSize={10} tickLine={false} axisLine={false} />
                                <YAxis stroke="#888888" fontSize={10} tickLine={false} axisLine={false} />
                                <CartesianGrid strokeDasharray="3 3" vertical={false} stroke="hsl(var(--muted))" />
                                <Tooltip
                                    contentStyle={{ borderRadius: '8px', border: '1px solid hsl(var(--border))', boxShadow: '0 4px 6px -1px rgb(0 0 0 / 0.1)' }}
                                    labelStyle={{ color: 'hsl(var(--foreground))' }}
                                />
                                <Line type="monotone" dataKey="latency" stroke="#3b82f6" strokeWidth={2} dot={false} name="Latency (ms)" />
                            </LineChart>
                        </ResponsiveContainer>
                    </div>
                </CardContent>
            </Card>

            {/* Error Rate */}
            <Card className="col-span-1 md:col-span-2 backdrop-blur-sm bg-background/50">
                <CardHeader>
                    <CardTitle className="text-base flex items-center gap-2">
                        <AlertCircle className="h-4 w-4 text-destructive" />
                        Success vs Errors
                    </CardTitle>
                    <CardDescription>Breakdown of successful requests and errors.</CardDescription>
                </CardHeader>
                <CardContent className="pl-0">
                    <div className="h-[250px] w-full">
                        <ResponsiveContainer width="100%" height="100%">
                            <BarChart data={data}>
                                <XAxis dataKey="time" stroke="#888888" fontSize={10} tickLine={false} axisLine={false} />
                                <YAxis stroke="#888888" fontSize={10} tickLine={false} axisLine={false} />
                                <CartesianGrid strokeDasharray="3 3" vertical={false} stroke="hsl(var(--muted))" />
                                <Tooltip
                                    contentStyle={{ borderRadius: '8px', border: '1px solid hsl(var(--border))', boxShadow: '0 4px 6px -1px rgb(0 0 0 / 0.1)' }}
                                    labelStyle={{ color: 'hsl(var(--foreground))' }}
                                    cursor={{fill: 'hsl(var(--muted)/0.2)'}}
                                />
                                <Legend />
                                <Bar dataKey="requests" name="Success" fill="#22c55e" stackId="a" />
                                <Bar dataKey="errors" name="Errors" fill="#ef4444" stackId="a" />
                            </BarChart>
                        </ResponsiveContainer>
                    </div>
                </CardContent>
            </Card>
        </div>
    </div>
  );
}
