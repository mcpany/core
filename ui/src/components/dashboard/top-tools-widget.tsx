/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useEffect, useState } from "react";
import { Bar, BarChart, ResponsiveContainer, Tooltip, XAxis, YAxis, CartesianGrid, Cell } from "recharts";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";

interface ToolUsageStats {
  name: string;
  serviceId: string;
  count: number;
}

export function TopToolsWidget() {
  const [data, setData] = useState<ToolUsageStats[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    async function fetchData() {
      try {
        const res = await fetch("/api/v1/dashboard/top-tools");
        if (res.ok) {
          const json = await res.json();
          setData(json || []);
        }
      } catch (error) {
        console.error("Failed to fetch top tools", error);
      } finally {
        setLoading(false);
      }
    }

    fetchData();
    // Refresh every 30s
    const interval = setInterval(fetchData, 30000);
    return () => clearInterval(interval);
  }, []);

  if (loading && data.length === 0) {
      return (
          <Card className="col-span-3 backdrop-blur-sm bg-background/50 h-full">
            <CardHeader>
                <CardTitle>Top Tools</CardTitle>
                <CardDescription>Most frequently executed tools.</CardDescription>
            </CardHeader>
            <CardContent className="h-[300px] flex items-center justify-center text-muted-foreground">
                Loading...
            </CardContent>
          </Card>
      )
  }

  // If no data (e.g. no tools used yet)
  if (data.length === 0) {
      return (
        <Card className="col-span-3 backdrop-blur-sm bg-background/50 h-full">
            <CardHeader>
                <CardTitle>Top Tools</CardTitle>
                <CardDescription>Most frequently executed tools.</CardDescription>
            </CardHeader>
            <CardContent className="h-[300px] flex items-center justify-center text-muted-foreground">
                No tool usage data yet.
            </CardContent>
        </Card>
      )
  }

  return (
    <Card className="col-span-3 backdrop-blur-sm bg-background/50 h-full">
      <CardHeader>
        <CardTitle>Top Tools</CardTitle>
        <CardDescription>
          Most frequently executed tools.
        </CardDescription>
      </CardHeader>
      <CardContent className="pl-2">
        <div className="h-[300px] w-full">
            <ResponsiveContainer width="100%" height="100%">
            <BarChart data={data} layout="vertical" margin={{ top: 5, right: 30, left: 20, bottom: 5 }}>
                <CartesianGrid strokeDasharray="3 3" horizontal={false} stroke="hsl(var(--muted))" />
                <XAxis type="number" hide />
                <YAxis
                    dataKey="name"
                    type="category"
                    stroke="#888888"
                    fontSize={12}
                    tickLine={false}
                    axisLine={false}
                    width={100}
                    tickFormatter={(value) => value.length > 15 ? value.substring(0, 15) + '...' : value}
                />
                <Tooltip
                    contentStyle={{ borderRadius: '8px', border: 'none', boxShadow: '0 4px 6px -1px rgb(0 0 0 / 0.1)', backgroundColor: 'hsl(var(--background))' }}
                    cursor={{fill: 'transparent'}}
                    formatter={(value: number) => [value, 'Executions']}
                />
                <Bar dataKey="count" fill="hsl(var(--primary))" radius={[0, 4, 4, 0]}>
                    {data.map((entry, index) => (
                        <Cell key={`cell-${index}`} fill="hsl(var(--primary))" fillOpacity={0.8 - (index * 0.05)} />
                    ))}
                </Bar>
            </BarChart>
            </ResponsiveContainer>
        </div>
      </CardContent>
    </Card>
  );
}
