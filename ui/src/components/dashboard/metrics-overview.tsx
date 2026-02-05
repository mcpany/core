/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useEffect, useState, memo } from "react";
import { useDashboard } from "@/components/dashboard/dashboard-context";
import {
  Users,
  Activity,
  Server,
  Zap,
  ArrowUpRight,
  ArrowDownRight,
  Database,
  MessageSquare,
  Clock,
  AlertCircle
} from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { SystemHealthCard } from "./system-health-card";

// Metric interface now imported from @/lib/client

const iconMap: Record<string, any> = {
  Users,
  Activity,
  Server,
  Zap,
  Database,
  MessageSquare,
  Clock,
  AlertCircle
};

// ⚡ Bolt Optimization: Extracted and memoized MetricItem to prevent unnecessary re-renders
// when only some metrics change during polling.
/**
 * MetricItem component.
 * @param props - The component props.
 * @param props.metric - The metric property.
 * @returns The rendered component.
 */
const MetricItem = memo(function MetricItem({ metric }: { metric: Metric }) {
  const Icon = iconMap[metric.icon] || Activity;
  const isPositiveTrend = metric.trend === "up";
  // For latency and errors, down is usually good (green), up is bad (red)
  const isReverseTrend = metric.label.includes("Latency") || metric.label.includes("Error");

  let trendColor = isPositiveTrend ? "text-green-500" : "text-red-500";
  if (isReverseTrend) {
      trendColor = isPositiveTrend ? "text-red-500" : "text-green-500";
  }

  return (
    <Card className="backdrop-blur-xl bg-background/60 border border-white/20 shadow-sm hover:shadow-lg transition-all duration-300">
      <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
        <CardTitle className="text-sm font-medium text-muted-foreground">
          {metric.label}
        </CardTitle>
        <Icon className="h-4 w-4 text-muted-foreground opacity-70" />
      </CardHeader>
      <CardContent>
        <div className="text-2xl font-bold tracking-tight">{metric.value}</div>
        <div className="flex items-center justify-between mt-1">
            {metric.change && (
          <p className={`text-xs flex items-center ${trendColor}`}>
            {metric.trend === "up" ? (
              <ArrowUpRight className="h-3 w-3 mr-1" />
            ) : (
              <ArrowDownRight className="h-3 w-3 mr-1" />
            )}
            <span>
              {metric.change}
            </span>
          </p>
        )}
          {metric.subLabel && (
            <span className="text-xs text-muted-foreground opacity-80">{metric.subLabel}</span>
          )}
        </div>
      </CardContent>
    </Card>
  );
});

// Memoized to prevent unnecessary re-renders when parent components update.
// This component manages its own state and data fetching, so it only needs to re-render
// when its own state changes, not when the parent re-renders.
import { apiClient, Metric } from "@/lib/client";

// ... (Icon map and MetricItem remain same)

/**
 * MetricsOverview displays a grid of key system metrics (e.g., QPS, Latency, Users)
 * and the system health status. It fetches data periodically from the API.
 * @returns The rendered MetricsOverview component.
 */
export const MetricsOverview = memo(function MetricsOverview() {
  const [metrics, setMetrics] = useState<Metric[]>([]);
  const { serviceId, timeRange } = useDashboard();

  useEffect(() => {
    async function fetchMetrics() {
      try {
        const data = await apiClient.getDashboardMetrics(serviceId, timeRange);
        setMetrics(data);
      } catch (error) {
        console.error("Failed to fetch metrics", error);
      }
    }
    fetchMetrics();
    // Poll every 5 seconds for real-time updates
    const interval = setInterval(() => {
      // ⚡ Bolt Optimization: Pause polling when tab is not visible to save bandwidth
      if (!document.hidden) {
        fetchMetrics();
      }
    }, 5000);

    // ⚡ Bolt Optimization: Refresh immediately when tab becomes visible
    const onVisibilityChange = () => {
      if (!document.hidden) {
        fetchMetrics();
      }
    };
    document.addEventListener("visibilitychange", onVisibilityChange);

    return () => {
      clearInterval(interval);
      document.removeEventListener("visibilitychange", onVisibilityChange);
    };
  }, [serviceId, timeRange]);

  if (metrics.length === 0) {
    return <div className="text-muted-foreground animate-pulse">Loading dashboard metrics...</div>;
  }

  return (
    <div className="space-y-4">
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        {metrics.map((metric) => (
          <MetricItem key={metric.label} metric={metric} />
        ))}
      </div>
      <SystemHealthCard />
    </div>
  );
});
