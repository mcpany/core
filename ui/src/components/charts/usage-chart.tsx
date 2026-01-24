"use client";

import { useState, useEffect, useMemo } from "react";
import {
  Bar,
  CartesianGrid,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
  Line,
  ComposedChart,
  Legend
} from "recharts";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { apiClient } from "@/lib/client";
import { Skeleton } from "@/components/ui/skeleton";
import { format, subHours, subDays, parseISO, startOfHour, startOfMinute, startOfDay, addMinutes } from "date-fns";
import { Activity } from "lucide-react";

export interface AuditLogEntry {
  timestamp: string;
  tool_name: string;
  duration_ms: number;
  error?: string;
}

interface UsageChartProps {
  toolName: string;
}

type TimeRange = "1h" | "24h" | "7d";

export function UsageChart({ toolName }: UsageChartProps) {
  const [data, setData] = useState<AuditLogEntry[]>([]);
  const [loading, setLoading] = useState(true);
  const [range, setRange] = useState<TimeRange>("24h");

  useEffect(() => {
    const fetchData = async () => {
      setLoading(true);
      try {
        const now = new Date();
        let startTime = subDays(now, 1);

        if (range === "1h") {
            startTime = subHours(now, 1);
        } else if (range === "7d") {
            startTime = subDays(now, 7);
        }

        // listAuditLogs takes ISO strings
        const logs = await apiClient.listAuditLogs({
          tool_name: toolName,
          start_time: startTime.toISOString(),
          end_time: now.toISOString(),
          limit: 1000 // reasonable limit
        });

        const list = Array.isArray(logs) ? logs : (logs.entries || []);
        setData(list);
      } catch (error) {
        console.error("Failed to fetch audit logs", error);
      } finally {
        setLoading(false);
      }
    };

    fetchData();
  }, [toolName, range]);

  const chartData = useMemo(() => {
    if (!data.length) return [];

    // Group by time bucket
    const buckets: Record<string, { time: string; count: number; errorCount: number; totalLatency: number, maxLatency: number }> = {};

    data.forEach(entry => {
        const date = parseISO(entry.timestamp);
        let bucketKey = ""; // Unique key for aggregation

        if (range === "1h") {
            bucketKey = startOfMinute(date).toISOString();
        } else if (range === "24h") {
            bucketKey = startOfHour(date).toISOString();
        } else {
             bucketKey = startOfDay(date).toISOString();
        }

        if (!buckets[bucketKey]) {
            buckets[bucketKey] = { time: bucketKey, count: 0, errorCount: 0, totalLatency: 0, maxLatency: 0 };
        }

        buckets[bucketKey].count++;
        if (entry.error) buckets[bucketKey].errorCount++;
        buckets[bucketKey].totalLatency += entry.duration_ms;
        if (entry.duration_ms > buckets[bucketKey].maxLatency) buckets[bucketKey].maxLatency = entry.duration_ms;
    });

    return Object.values(buckets).map(b => ({
        ...b,
        avgLatency: Math.round(b.totalLatency / b.count)
    }));
  }, [data, range]);

  // Sort chartData correctly
  const sortedChartData = useMemo(() => {
      const now = new Date();
      const result = [];

      let current: Date;
      let end: Date;
      let intervalMinutes: number;
      let formatStr: string;
      let uniqueKeyFn: (d: Date) => Date;

      if (range === "1h") {
          current = subHours(now, 1);
          end = now;
          intervalMinutes = 1;
          formatStr = "HH:mm";
          uniqueKeyFn = startOfMinute;
      } else if (range === "24h") {
          current = startOfHour(subDays(now, 1));
          end = startOfHour(now);
          intervalMinutes = 60;
          formatStr = "HH:mm";
          uniqueKeyFn = startOfHour;
      } else {
           // 7d
           current = subDays(now, 7);
           end = now;
           intervalMinutes = 60 * 24;
           formatStr = "MM/dd";
           uniqueKeyFn = startOfDay;
      }

      // Map data to result buckets by unique key
      const dataMap = new Map(chartData.map(d => [d.time, d]));

      while (current <= end) {
          const uniqueKey = uniqueKeyFn(current).toISOString();
          const label = format(current, formatStr);

          const d = dataMap.get(uniqueKey);

          result.push({
              time: label,
              fullDate: current.getTime(), // Optional, for reference
              count: d ? d.count : 0,
              errorCount: d ? d.errorCount : 0,
              avgLatency: d ? d.avgLatency : 0,
              maxLatency: d ? d.maxLatency : 0
          });

          current = addMinutes(current, intervalMinutes);
      }

      return result;
  }, [chartData, range]);


  if (loading) {
      return (
          <Card>
              <CardHeader>
                  <Skeleton className="h-6 w-1/3" />
              </CardHeader>
              <CardContent>
                  <Skeleton className="h-[300px] w-full" />
              </CardContent>
          </Card>
      )
  }

  if (data.length === 0) {
      // Show empty state but with the chart frame? or just "No data"?
      // Let's show "No usage data" inside the card.
  }

  return (
    <Card className="col-span-1">
      <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
        <CardTitle className="text-xl flex items-center gap-2">
          <Activity className="h-5 w-5" /> Usage Over Time
        </CardTitle>
        <Select value={range} onValueChange={(v) => setRange(v as TimeRange)}>
          <SelectTrigger className="w-[120px]">
            <SelectValue placeholder="Range" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="1h">Last Hour</SelectItem>
            <SelectItem value="24h">Last 24 Hours</SelectItem>
            <SelectItem value="7d">Last 7 Days</SelectItem>
          </SelectContent>
        </Select>
      </CardHeader>
      <CardContent className="pt-4">
        <div className="h-[300px] w-full">
            {sortedChartData.length > 0 && data.length > 0 ? (
                 <ResponsiveContainer width="100%" height="100%">
                 <ComposedChart data={sortedChartData}>
                   <CartesianGrid strokeDasharray="3 3" vertical={false} />
                   <XAxis
                     dataKey="time"
                     stroke="#888888"
                     fontSize={12}
                     tickLine={false}
                     axisLine={false}
                   />
                   <YAxis
                     yAxisId="left"
                     stroke="#888888"
                     fontSize={12}
                     tickLine={false}
                     axisLine={false}
                     tickFormatter={(value) => `${value}`}
                     label={{ value: 'Calls', angle: -90, position: 'insideLeft', style: { fill: '#888' } }}
                   />
                   <YAxis
                     yAxisId="right"
                     orientation="right"
                     stroke="#888888"
                     fontSize={12}
                     tickLine={false}
                     axisLine={false}
                     unit="ms"
                     label={{ value: 'Latency (ms)', angle: 90, position: 'insideRight', style: { fill: '#888' } }}
                   />
                   <Tooltip
                     contentStyle={{ backgroundColor: 'var(--card)', borderColor: 'var(--border)', borderRadius: '6px' }}
                     itemStyle={{ color: 'var(--foreground)' }}
                     labelStyle={{ color: 'var(--muted-foreground)' }}
                   />
                   <Legend />
                   <Bar yAxisId="left" dataKey="count" name="Executions" fill="var(--primary)" radius={[4, 4, 0, 0]} maxBarSize={40} />
                   <Bar yAxisId="left" dataKey="errorCount" name="Errors" fill="var(--destructive)" radius={[4, 4, 0, 0]} maxBarSize={40} />
                   <Line yAxisId="right" type="monotone" dataKey="avgLatency" name="Avg Latency" stroke="var(--chart-2)" strokeWidth={2} dot={false} />
                 </ComposedChart>
               </ResponsiveContainer>
            ) : (
                <div className="flex h-full items-center justify-center text-muted-foreground">
                    No usage data available for this period.
                </div>
            )}

        </div>
      </CardContent>
    </Card>
  );
}
