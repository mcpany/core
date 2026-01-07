/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useEffect, useState } from "react";
import { MetricsOverview, MetricsData } from "@/components/dashboard/metrics-overview";
import { ServiceHealthWidget, ServiceHealthData } from "@/components/dashboard/service-health-widget";
import { RequestVolumeChart, RequestVolumeData } from "@/components/dashboard/request-volume-chart";
import { Button } from "@/components/ui/button";
import { apiClient } from "@/lib/client";

export default function DashboardPage() {
  const [metrics, setMetrics] = useState<MetricsData | null>(null);
  const [healthData, setHealthData] = useState<ServiceHealthData[]>([]);
  const [volumeData, setVolumeData] = useState<RequestVolumeData[]>([]);

  useEffect(() => {
    async function loadData() {
        try {
            // In a real scenario, these would be real API calls
            // For now, we mock them via the client or here if the client doesn't support dashboard specific endpoints yet.
            // But consistent with the plan, we should use the client.
            // Let's assume we add dashboard methods to apiClient or use existing ones to aggregate.

            // Simulating API response delay
            await new Promise(r => setTimeout(r, 500));

            setMetrics({
                activeServices: 12,
                requestsPerSec: 2354,
                avgLatency: 46,
                activeResources: 1204
            });

            setHealthData([
                { name: "Payments Service", status: "active", uptime: "99.9%", latency: "12ms" },
                { name: "Auth Service", status: "active", uptime: "99.99%", latency: "8ms" },
                { name: "User Profile", status: "warning", uptime: "98.5%", latency: "145ms" },
                { name: "Email Gateway", status: "active", uptime: "100%", latency: "45ms" },
                { name: "Legacy API", status: "inactive", uptime: "0%", latency: "-" },
            ]);

            setVolumeData([
                { name: "00:00", total: 4000 },
                { name: "04:00", total: 3000 },
                { name: "08:00", total: 2000 },
                { name: "12:00", total: 2780 },
                { name: "16:00", total: 1890 },
                { name: "20:00", total: 2390 },
                { name: "24:00", total: 3490 },
            ]);

        } catch (e) {
            console.error("Failed to load dashboard data", e);
        }
    }
    loadData();
  }, []);

  return (
    <div className="flex-1 space-y-4 p-8 pt-6">
      <div className="flex items-center justify-between space-y-2">
        <h1 className="text-3xl font-bold tracking-tight">Dashboard</h1>
        <div className="flex items-center space-x-2">
          <Button>Download Report</Button>
        </div>
      </div>
      <div className="space-y-4">
        {metrics && <MetricsOverview data={metrics} />}
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-7">
          <ServiceHealthWidget services={healthData} />
          <RequestVolumeChart data={volumeData} />
        </div>
      </div>
    </div>
  );
}
