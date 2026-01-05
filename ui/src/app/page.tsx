/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { MetricsOverview } from "@/components/dashboard/metrics-overview";
import { ServiceHealthWidget } from "@/components/dashboard/service-health-widget";
import { RequestVolumeChart } from "@/components/dashboard/request-volume-chart";
import { Button } from "@/components/ui/button";

export default function DashboardPage() {
  return (
    <div className="flex-1 space-y-4 p-4 md:p-8 pt-6">
      <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4 sm:gap-0">
        <h1 className="text-2xl sm:text-3xl font-bold tracking-tight">Dashboard</h1>
        <div className="flex items-center space-x-2 w-full sm:w-auto">
          <Button className="w-full sm:w-auto">Download Report</Button>
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
