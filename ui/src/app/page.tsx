/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useEffect, useState } from "react";
import { DashboardGrid } from "@/components/dashboard/dashboard-grid";
import { Button } from "@/components/ui/button";
import { DashboardProvider } from "@/components/dashboard/dashboard-context";
import { ServiceFilter } from "@/components/dashboard/service-filter";
import { TimeRangeFilter } from "@/components/dashboard/time-range-filter";
import { apiClient } from "@/lib/client";
import { OnboardingView } from "@/components/onboarding/onboarding-view";
import { Loader2 } from "lucide-react";

/**
 * The main dashboard page component.
 * Displays an overview of metrics, service health, and request volume.
 * @returns The dashboard page.
 */
export default function DashboardPage() {
  const [loading, setLoading] = useState(true);
  const [serviceCount, setServiceCount] = useState<number | null>(null);

  useEffect(() => {
    apiClient.listServices()
      .then((services) => {
        setServiceCount(services.length);
      })
      .catch((err) => {
        console.error("Failed to list services", err);
        // On error, default to 0 to show onboarding which might be helpful to debug or retry
        setServiceCount(0);
      })
      .finally(() => setLoading(false));
  }, []);

  if (loading) {
      return (
          <div className="flex items-center justify-center h-[calc(100vh-4rem)]">
              <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
          </div>
      );
  }

  if (serviceCount === 0) {
      return <OnboardingView />;
  }

  return (
    <DashboardProvider>
      <div className="flex-1 space-y-4 p-8 pt-6">
        <div className="flex items-center justify-between space-y-2">
          <h1 className="text-3xl font-bold tracking-tight">Dashboard</h1>
          <div className="flex items-center space-x-2">
            <ServiceFilter />
            <TimeRangeFilter />
            <Button>Download Report</Button>
          </div>
        </div>
        <div className="space-y-4">
          <DashboardGrid />
        </div>
      </div>
    </DashboardProvider>
  );
}
