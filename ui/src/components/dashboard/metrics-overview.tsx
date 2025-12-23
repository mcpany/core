
"use client";

import { useEffect, useState } from "react";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Activity, Zap, Server, Database } from "lucide-react";
import { Skeleton } from "@/components/ui/skeleton";

export function MetricsOverview() {
  const [metrics, setMetrics] = useState<any>(null);

  useEffect(() => {
    fetch("/api/dashboard/metrics")
      .then(res => res.json())
      .then(data => setMetrics(data));
  }, []);

  if (!metrics) {
      return <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
          {[1,2,3,4].map(i => <Skeleton key={i} className="h-[120px] rounded-xl" />)}
      </div>
  }

  return (
    <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
      <Card className="backdrop-blur-xl bg-background/60 border-muted/20 shadow-sm hover:shadow-md transition-all duration-300">
        <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
          <CardTitle className="text-sm font-medium">Total Requests</CardTitle>
          <Activity className="h-4 w-4 text-muted-foreground" />
        </CardHeader>
        <CardContent>
          <div className="text-2xl font-bold">{metrics.totalRequests.toLocaleString()}</div>
          <p className="text-xs text-muted-foreground">
             {metrics.requestsChange > 0 ? "+" : ""}{metrics.requestsChange}% from last month
          </p>
        </CardContent>
      </Card>
      <Card className="backdrop-blur-xl bg-background/60 border-muted/20 shadow-sm hover:shadow-md transition-all duration-300">
        <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
          <CardTitle className="text-sm font-medium">Avg Latency</CardTitle>
          <Zap className="h-4 w-4 text-muted-foreground" />
        </CardHeader>
        <CardContent>
          <div className="text-2xl font-bold">{metrics.avgLatency}</div>
          <p className="text-xs text-muted-foreground">
             {metrics.latencyChange > 0 ? "+" : ""}{metrics.latencyChange}% from last month
          </p>
        </CardContent>
      </Card>
      <Card className="backdrop-blur-xl bg-background/60 border-muted/20 shadow-sm hover:shadow-md transition-all duration-300">
        <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
          <CardTitle className="text-sm font-medium">Active Services</CardTitle>
          <Server className="h-4 w-4 text-muted-foreground" />
        </CardHeader>
        <CardContent>
          <div className="text-2xl font-bold">{metrics.activeServices}</div>
          <p className="text-xs text-muted-foreground">
             {metrics.servicesChange > 0 ? "+" : ""}{metrics.servicesChange} newly connected
          </p>
        </CardContent>
      </Card>
      <Card className="backdrop-blur-xl bg-background/60 border-muted/20 shadow-sm hover:shadow-md transition-all duration-300">
        <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
          <CardTitle className="text-sm font-medium">Resources Served</CardTitle>
          <Database className="h-4 w-4 text-muted-foreground" />
        </CardHeader>
        <CardContent>
          <div className="text-2xl font-bold">{metrics.resourcesServed}</div>
          <p className="text-xs text-muted-foreground">
             {metrics.resourcesChange > 0 ? "+" : ""}{metrics.resourcesChange} since last hour
          </p>
        </CardContent>
      </Card>
    </div>
  );
}
