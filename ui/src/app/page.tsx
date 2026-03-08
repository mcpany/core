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
import { OnboardingHero } from "@/components/dashboard/onboarding-hero";
import { apiClient } from "@/lib/client";
import { Loader2 } from "lucide-react";

/**
 * The main dashboard page component.
 * Displays an overview of metrics, service health, and request volume.
 * @returns The dashboard page.
 */
export default function DashboardPage() {
  const [loading, setLoading] = useState(true);
  const [hasServices, setHasServices] = useState(false);

  useEffect(() => {
    async function checkServices() {
        try {
            const services = await apiClient.listServices();
            setHasServices(services && services.length > 0);
        } catch (e) {
            console.error("Failed to check services", e);
            // Default to showing dashboard if error, to avoid getting stuck on hero
            setHasServices(true);
        } finally {
            setLoading(false);
        }
    }
    checkServices();
  }, []);

  if (loading) {
      return (
          <div className="flex flex-1 items-center justify-center h-[calc(100vh-4rem)]">
              <div className="flex flex-col items-center gap-4 animate-pulse">
                  <div className="h-12 w-12 rounded-xl bg-primary/20 flex items-center justify-center">
                      <Loader2 className="h-6 w-6 animate-spin text-primary" />
                  </div>
                  <div className="h-4 w-24 bg-muted rounded"></div>
              </div>
          </div>
      );
  }

  if (!hasServices) {
      return <OnboardingHero />;
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
