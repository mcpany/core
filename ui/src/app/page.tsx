"use client"

import * as React from "react"
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card"
import { StatusBadge } from "@/components/ui-custom/status-badge"
import { GlassCard } from "@/components/ui-custom/glass-card"
import { Activity, Server, Zap, Database, ArrowUpRight, ArrowDownRight, Clock } from "lucide-react"
import {
  Area,
  AreaChart,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
  CartesianGrid,
  BarChart,
  Bar,
} from "recharts"

const mockData = [
  { time: "10:00", req: 400, latency: 240 },
  { time: "10:05", req: 300, latency: 139 },
  { time: "10:10", req: 200, latency: 980 },
  { time: "10:15", req: 278, latency: 390 },
  { time: "10:20", req: 189, latency: 480 },
  { time: "10:25", req: 239, latency: 380 },
  { time: "10:30", req: 349, latency: 430 },
]

const servicesStatus = [
  { name: "weather-service", status: "active", uptime: "99.9%" },
  { name: "memory-store", status: "inactive", uptime: "0%" },
  { name: "local-files", status: "active", uptime: "99.9%" },
  { name: "gpt-4-proxy", status: "active", uptime: "98.5%" },
  { name: "sql-connector", status: "warning", uptime: "95.0%" },
]

export default function Dashboard() {
  const [mounted, setMounted] = React.useState(false)

  React.useEffect(() => {
    setMounted(true)
  }, [])

  if (!mounted) return null

  return (
    <div className="space-y-8">
      <div>
        <h2 className="text-3xl font-bold tracking-tight">Dashboard</h2>
        <p className="text-muted-foreground mt-2">
          Overview of your MCP Any infrastructure.
        </p>
      </div>

      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        <GlassCard className="p-6 flex flex-col justify-between space-y-2" hoverEffect>
          <div className="flex items-center justify-between text-sm font-medium text-muted-foreground">
            <span>Total Requests</span>
            <Activity className="h-4 w-4 text-primary" />
          </div>
          <div className="text-2xl font-bold">45.2k</div>
          <div className="text-xs text-muted-foreground flex items-center gap-1">
            <span className="text-emerald-500 flex items-center">
              <ArrowUpRight className="h-3 w-3 mr-0.5" /> +20.1%
            </span>
            from last month
          </div>
        </GlassCard>

        <GlassCard className="p-6 flex flex-col justify-between space-y-2" hoverEffect>
           <div className="flex items-center justify-between text-sm font-medium text-muted-foreground">
            <span>Active Services</span>
            <Server className="h-4 w-4 text-blue-500" />
          </div>
          <div className="text-2xl font-bold">12</div>
          <div className="text-xs text-muted-foreground flex items-center gap-1">
             <span className="text-emerald-500 flex items-center">
              <ArrowUpRight className="h-3 w-3 mr-0.5" /> +2
            </span>
            new since yesterday
          </div>
        </GlassCard>

         <GlassCard className="p-6 flex flex-col justify-between space-y-2" hoverEffect>
           <div className="flex items-center justify-between text-sm font-medium text-muted-foreground">
            <span>Avg Latency</span>
            <Clock className="h-4 w-4 text-amber-500" />
          </div>
          <div className="text-2xl font-bold">124ms</div>
          <div className="text-xs text-muted-foreground flex items-center gap-1">
             <span className="text-emerald-500 flex items-center">
              <ArrowDownRight className="h-3 w-3 mr-0.5" /> -10ms
            </span>
            improved
          </div>
        </GlassCard>

        <GlassCard className="p-6 flex flex-col justify-between space-y-2" hoverEffect>
           <div className="flex items-center justify-between text-sm font-medium text-muted-foreground">
            <span>Error Rate</span>
            <Zap className="h-4 w-4 text-red-500" />
          </div>
          <div className="text-2xl font-bold">0.4%</div>
           <div className="text-xs text-muted-foreground flex items-center gap-1">
             <span className="text-emerald-500 flex items-center">
              <ArrowDownRight className="h-3 w-3 mr-0.5" /> -0.1%
            </span>
            improved
          </div>
        </GlassCard>
      </div>

      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-7">
        <GlassCard className="col-span-4" hoverEffect>
          <CardHeader>
            <CardTitle>Traffic Overview</CardTitle>
            <CardDescription>
              Requests per second over time.
            </CardDescription>
          </CardHeader>
          <CardContent className="pl-2">
            <ResponsiveContainer width="100%" height={300}>
              <AreaChart data={mockData}>
                 <defs>
                  <linearGradient id="colorReq" x1="0" y1="0" x2="0" y2="1">
                    <stop offset="5%" stopColor="hsl(var(--primary))" stopOpacity={0.3}/>
                    <stop offset="95%" stopColor="hsl(var(--primary))" stopOpacity={0}/>
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
                    contentStyle={{ backgroundColor: 'hsl(var(--popover))', borderColor: 'hsl(var(--border))', borderRadius: '8px' }}
                    itemStyle={{ color: 'hsl(var(--popover-foreground))' }}
                />
                <CartesianGrid strokeDasharray="3 3" vertical={false} stroke="hsl(var(--border))" />
                <Area
                    type="monotone"
                    dataKey="req"
                    stroke="hsl(var(--primary))"
                    fillOpacity={1}
                    fill="url(#colorReq)"
                    strokeWidth={2}
                />
              </AreaChart>
            </ResponsiveContainer>
          </CardContent>
        </GlassCard>

        <GlassCard className="col-span-3" hoverEffect>
          <CardHeader>
            <CardTitle>Service Status</CardTitle>
            <CardDescription>
              Real-time health of connected MCP servers.
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              {servicesStatus.map((service) => (
                <div key={service.name} className="flex items-center justify-between border-b pb-2 last:border-0 last:pb-0">
                  <div className="flex items-center gap-3">
                    <div className="flex flex-col">
                        <span className="text-sm font-medium leading-none">{service.name}</span>
                        <span className="text-xs text-muted-foreground mt-1">Uptime: {service.uptime}</span>
                    </div>
                  </div>
                  <StatusBadge status={service.status as any} />
                </div>
              ))}
            </div>
          </CardContent>
        </GlassCard>
      </div>
    </div>
  )
}
