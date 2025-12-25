/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { MetricsOverview } from "@/components/dashboard/metrics-overview";
import { ServiceHealthWidget } from "@/components/dashboard/service-health-widget";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Activity, BarChart, Zap } from "lucide-react";

export default function DashboardPage() {
  return (
    <div className="flex-1 space-y-4 p-8 pt-6">
      <div className="flex items-center justify-between space-y-2">
        <h2 className="text-3xl font-bold tracking-tight">Dashboard</h2>
        <div className="flex items-center space-x-2">
           {/* Date Range Picker could go here */}
        </div>
      </div>
      <div className="space-y-4">
        <MetricsOverview />
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-7">
          <ServiceHealthWidget />
          <Card className="col-span-3 backdrop-blur-sm bg-background/50">
            <CardHeader>
              <CardTitle>Recent Activity</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="space-y-8">
                {[1, 2, 3].map((_, i) => (
                    <div key={i} className="flex items-center">
                        <div className="h-9 w-9 rounded-full bg-blue-500/10 flex items-center justify-center border border-blue-500/20">
                            <Activity className="h-4 w-4 text-blue-500" />
                        </div>
                        <div className="ml-4 space-y-1">
                            <p className="text-sm font-medium leading-none">Service Updated</p>
                            <p className="text-sm text-muted-foreground">
                                Payment Gateway configuration changed.
                            </p>
                        </div>
                        <div className="ml-auto font-medium text-xs text-muted-foreground">
                            {i * 15 + 2}m ago
                        </div>
                    </div>
                ))}
              </div>
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  );
}
