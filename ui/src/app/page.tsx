/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */


import { MetricsOverview } from "@/components/dashboard/metrics-overview";
import { ServiceHealthWidget } from "@/components/dashboard/service-health-widget";
import { RequestVolumeChart } from "@/components/dashboard/request-volume-chart";

export default function DashboardPage() {
  return (
    <div className="flex-1 space-y-4 p-8 pt-6">
      <div className="flex items-center justify-between space-y-2">
        <h2 className="text-3xl font-bold tracking-tight">Dashboard</h2>
      </div>
      <div className="space-y-4">
        <MetricsOverview />
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-7">
          <RequestVolumeChart />
          <ServiceHealthWidget />
        </div>
      </div>
    </div>
  );
}
