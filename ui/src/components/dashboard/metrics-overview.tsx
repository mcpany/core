/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useEffect, useState } from "react";
import {
  Users,
  Activity,
  Server,
  Zap,
  ArrowUpRight,
  ArrowDownRight
} from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";

interface Metric {
  label: string;
  value: string;
  change?: string;
  trend?: "up" | "down" | "neutral";
  icon: string;
}

const iconMap: Record<string, any> = {
  Users,
  Activity,
  Server,
  Zap,
};

export function MetricsOverview() {
  const [metrics, setMetrics] = useState<Metric[]>([]);

  useEffect(() => {
    async function fetchMetrics() {
      try {
        const res = await fetch("/api/dashboard/metrics");
        if (res.ok) {
          const data = await res.json();
          setMetrics(data);
        }
      } catch (error) {
        console.error("Failed to fetch metrics", error);
      }
    }
    fetchMetrics();
    // Poll every 5 seconds for real-time updates
    const interval = setInterval(fetchMetrics, 5000);
    return () => clearInterval(interval);
  }, []);

  if (metrics.length === 0) {
    return <div className="text-muted-foreground">Loading metrics...</div>;
  }

  return (
    <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
      {metrics.map((metric) => {
        const Icon = iconMap[metric.icon] || Activity;
        return (
          <Card key={metric.label} className="backdrop-blur-sm bg-background/50 border shadow-sm hover:shadow-md transition-all">
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">
                {metric.label}
              </CardTitle>
              <Icon className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{metric.value}</div>
              {metric.change && (
                <p className="text-xs text-muted-foreground flex items-center mt-1">
                  {metric.trend === "up" ? (
                    <ArrowUpRight className="h-3 w-3 text-green-500 mr-1" />
                  ) : (
                    <ArrowDownRight className="h-3 w-3 text-red-500 mr-1" />
                  )}
                  <span className={metric.trend === "up" ? "text-green-500" : "text-red-500"}>
                    {metric.change}
                  </span>
                  <span className="ml-1">from last month</span>
                </p>
              )}
            </CardContent>
          </Card>
        );
      })}
    </div>
  );
}
