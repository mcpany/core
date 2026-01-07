
"use client";

import { Bar, BarChart, ResponsiveContainer, XAxis, YAxis, Tooltip } from "recharts";
import { GlassCard } from "@/components/layout/glass-card";
import { CardHeader, CardTitle, CardDescription, CardContent } from "@/components/ui/card";

export interface RequestVolumeData {
    name: string;
    total: number;
}

export function RequestVolumeChart({ data }: { data: RequestVolumeData[] }) {
  return (
    <GlassCard className="col-span-4">
      <CardHeader>
        <CardTitle>Request Volume</CardTitle>
        <CardDescription>Total requests processed over the last 24 hours.</CardDescription>
      </CardHeader>
      <CardContent className="pl-2">
        <ResponsiveContainer width="100%" height={350}>
          <BarChart data={data}>
            <XAxis
              dataKey="name"
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
                contentStyle={{
                    backgroundColor: "hsl(var(--card))",
                    borderColor: "hsl(var(--border))",
                    borderRadius: "var(--radius)",
                }}
                cursor={{ fill: "hsl(var(--muted))" }}
            />
            <Bar
                dataKey="total"
                fill="hsl(var(--primary))"
                radius={[4, 4, 0, 0]}
            />
          </BarChart>
        </ResponsiveContainer>
      </CardContent>
    </GlassCard>
  );
}
