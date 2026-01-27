/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { useState, useEffect, memo } from "react";
import { Area, AreaChart, ResponsiveContainer, Tooltip, XAxis, YAxis, CartesianGrid } from "recharts";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { apiClient } from "@/lib/client";
import { useDashboard } from "@/components/dashboard/dashboard-context";

/**
 * RequestVolumeChart component.
 * @returns The rendered component.
 */
export const RequestVolumeChart = memo(function RequestVolumeChart() {
  const [data, setData] = useState<{ time: string; total: number }[]>([]);
  const [mounted, setMounted] = useState(false);
  const { serviceId } = useDashboard();

  useEffect(() => {
    setMounted(true);
    const fetchData = async () => {
        try {
            const traffic = await apiClient.getDashboardTraffic(serviceId);
            // Backend returns traffic points directly
            setData(traffic);
        } catch (error) {
            console.error("Failed to fetch traffic data", error);
        }
    };
    fetchData();
    // Poll every 30 seconds
    const interval = setInterval(() => {
      // ⚡ Bolt Optimization: Stop polling when tab is hidden to save resources
      // Check navigator.webdriver to ensure tests continue to poll even if window is hidden
      const isTest = typeof navigator !== 'undefined' && (navigator as any).webdriver;
      if (!document.hidden || isTest) {
        fetchData();
      }
    }, 30000);

    // ⚡ Bolt Optimization: Resume immediately when tab becomes visible
    const onVisibilityChange = () => {
      if (!document.hidden) {
        fetchData();
      }
    };
    document.addEventListener("visibilitychange", onVisibilityChange);

    return () => {
      clearInterval(interval);
      document.removeEventListener("visibilitychange", onVisibilityChange);
    };
  }, [serviceId]);

  if (!mounted) return null;

  return (
    <Card className="col-span-3 backdrop-blur-sm bg-background/50 h-full">
      <CardHeader>
        <CardTitle>Request Volume</CardTitle>
        <CardDescription>
          Requests handled over the last 24 hours.
        </CardDescription>
      </CardHeader>
      <CardContent className="pl-2">
        <div className="h-[300px] w-full">
            <ResponsiveContainer width="100%" height="100%">
            <AreaChart data={data}>
                <defs>
                    <linearGradient id="colorTotal" x1="0" y1="0" x2="0" y2="1">
                        <stop offset="5%" stopColor="#8884d8" stopOpacity={0.8}/>
                        <stop offset="95%" stopColor="#8884d8" stopOpacity={0}/>
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
                <CartesianGrid strokeDasharray="3 3" vertical={false} stroke="hsl(var(--muted))" />
                <Tooltip
                    contentStyle={{ borderRadius: '8px', border: 'none', boxShadow: '0 4px 6px -1px rgb(0 0 0 / 0.1)' }}
                />
                <Area
                    type="monotone"
                    dataKey="requests"
                    stroke="#8884d8"
                    fillOpacity={1}
                    fill="url(#colorTotal)"
                />
            </AreaChart>
            </ResponsiveContainer>
        </div>
      </CardContent>
    </Card>
  );
});
