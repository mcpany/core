/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useEffect } from "react";
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
  const [hasServices, setHasServices] = useState<boolean | null>(null);
  const [refreshKey, setRefreshKey] = useState(0);

  useEffect(() => {
    async function checkServices() {
        try {
            const data = await apiClient.listServices();
            // data is Array or { services: Array }
            const list = Array.isArray(data) ? data : (data.services || []);
            setHasServices(list.length > 0);
        } catch (e) {
            console.error("Failed to check services", e);
            // Assume 0 on error so user can try to register
            setHasServices(false);
        }
    }
    checkServices();
  }, [refreshKey]);

  if (hasServices === null) {
      return (
          <div className="flex items-center justify-center h-full">
              <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
          </div>
      );
  }

  if (hasServices === false) {
      return <OnboardingView onServiceRegistered={() => setRefreshKey(k => k + 1)} />;
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
