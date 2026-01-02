/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { MetricsOverview } from "@/components/dashboard/metrics-overview";
import { ServiceHealthWidget } from "@/components/dashboard/service-health-widget";
import { RequestVolumeChart } from "@/components/dashboard/request-volume-chart";
import { Button } from "@/components/ui/button";

/**
 * The main dashboard page component.
 *
 * Displays an overview of metrics, service health, and request volume.
 */
export default function DashboardPage() {
  return (
    <div className="flex-1 space-y-4 p-8 pt-6">
      <div className="flex items-center justify-between space-y-2">
        <h2 className="text-3xl font-bold tracking-tight">Dashboard</h2>
        <div className="flex items-center space-x-2">
          <Button>Download Report</Button>
        </div>
      </div>
      <div className="space-y-4">
        <MetricsOverview />
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-7">
          <ServiceHealthWidget />
          <RequestVolumeChart />
        </div>
      </div>
    </div>
  );
}
