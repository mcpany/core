/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client"

import {
  Bar,
  BarChart,
  ResponsiveContainer,
  XAxis,
  YAxis,
  Tooltip,
} from "recharts"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"

interface ActivityChartProps {
  data: any[]
  loading?: boolean
}

export function ActivityChart({ data, loading }: ActivityChartProps) {
  if (loading) {
    return (
      <Card className="col-span-4 border-none shadow-none bg-muted/10">
        <CardHeader>
          <CardTitle className="text-sm font-medium">Activity Volume</CardTitle>
        </CardHeader>
        <CardContent className="h-[200px] flex items-center justify-center text-muted-foreground">
            Loading...
        </CardContent>
      </Card>
    )
  }

  // Group data by hour
  const groupedData = data.reduce((acc: any, log: any) => {
    const date = new Date(log.timestamp)
    // Round down to hour
    date.setMinutes(0, 0, 0)
    const key = date.toISOString()

    if (!acc[key]) {
      acc[key] = { time: key, display: date.getHours().toString().padStart(2, '0') + ":00", count: 0, errors: 0 }
    }
    acc[key].count++
    if (log.error) {
      acc[key].errors++
    }
    return acc
  }, {})

  const chartData = Object.values(groupedData).sort((a: any, b: any) => a.time.localeCompare(b.time))

  return (
    <Card className="col-span-4">
      <CardHeader>
        <CardTitle className="text-sm font-medium">Activity Volume (Last 24h)</CardTitle>
      </CardHeader>
      <CardContent className="pl-2">
        <div className="h-[200px] w-full">
            <ResponsiveContainer width="100%" height="100%">
            <BarChart data={chartData}>
                <XAxis
                dataKey="display"
                stroke="#888888"
                fontSize={10}
                tickLine={false}
                axisLine={false}
                minTickGap={30}
                />
                <YAxis
                stroke="#888888"
                fontSize={10}
                tickLine={false}
                axisLine={false}
                />
                <Tooltip
                    contentStyle={{
                        backgroundColor: "hsl(var(--popover))",
                        border: "1px solid hsl(var(--border))",
                        borderRadius: "6px",
                        fontSize: "12px"
                    }}
                    cursor={{fill: 'hsl(var(--muted)/0.3)'}}
                />
                <Bar
                    dataKey="count"
                    fill="hsl(var(--primary))"
                    radius={[2, 2, 0, 0]}
                    maxBarSize={40}
                />
                 <Bar
                    dataKey="errors"
                    fill="hsl(var(--destructive))"
                    radius={[2, 2, 0, 0]}
                    maxBarSize={40}
                />
            </BarChart>
            </ResponsiveContainer>
        </div>
      </CardContent>
    </Card>
  )
}
