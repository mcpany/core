
"use client";

import { useEffect, useState } from "react";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Area, AreaChart, ResponsiveContainer, Tooltip, XAxis, YAxis, CartesianGrid } from "recharts";
import { Skeleton } from "@/components/ui/skeleton";

export function RequestVolumeChart() {
  const [data, setData] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetch("/api/dashboard/metrics")
      .then(res => res.json())
      .then(d => {
          setData(d.requestVolume);
          setLoading(false);
      });
  }, []);

  return (
    <Card className="col-span-4 backdrop-blur-xl bg-background/60 border-muted/20 shadow-sm">
      <CardHeader>
        <CardTitle>Request Volume</CardTitle>
        <CardDescription>Requests per second over the last 30 minutes.</CardDescription>
      </CardHeader>
      <CardContent className="pl-2">
        {loading ? (
            <Skeleton className="h-[300px] w-full" />
        ) : (
            <ResponsiveContainer width="100%" height={300}>
            <AreaChart data={data}>
                <defs>
                <linearGradient id="colorReqs" x1="0" y1="0" x2="0" y2="1">
                    <stop offset="5%" stopColor="hsl(var(--primary))" stopOpacity={0.3} />
                    <stop offset="95%" stopColor="hsl(var(--primary))" stopOpacity={0} />
                </linearGradient>
                </defs>
                <CartesianGrid strokeDasharray="3 3" className="stroke-muted/20" vertical={false} />
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
                    contentStyle={{ backgroundColor: "hsl(var(--background))", borderRadius: "8px", border: "1px solid hsl(var(--border))" }}
                    itemStyle={{ color: "hsl(var(--foreground))" }}
                />
                <Area
                type="monotone"
                dataKey="reqs"
                stroke="hsl(var(--primary))"
                strokeWidth={2}
                fillOpacity={1}
                fill="url(#colorReqs)"
                />
            </AreaChart>
            </ResponsiveContainer>
        )}
      </CardContent>
    </Card>
  );
}
