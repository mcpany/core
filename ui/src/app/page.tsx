/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */



import { useState, useEffect } from "react";
import { DashboardGrid } from "@/components/dashboard/dashboard-grid";
import { DashboardProvider } from "@/components/dashboard/dashboard-context";
import { ServiceFilter } from "@/components/dashboard/service-filter";
import { TimeRangeFilter } from "@/components/dashboard/time-range-filter";
import { OnboardingHero } from "@/components/dashboard/onboarding-hero";
import { apiClient } from "@/lib/client";
import { Loader2 } from "lucide-react";
import { DownloadReportButton } from "@/components/dashboard/download-report-button";

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
          <div className="flex flex-1 items-center justify-center h-screen w-full">
            <div className="flex flex-col items-center gap-4 text-muted-foreground">
                <Loader2 className="h-8 w-8 animate-spin text-primary" />
                <p className="text-sm font-medium tracking-wide">Loading dashboard...</p>
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
            <DownloadReportButton />
          </div>
        </div>
        <div className="space-y-4">
          <DashboardGrid />
        </div>
      </div>
    </DashboardProvider>
  );
}
