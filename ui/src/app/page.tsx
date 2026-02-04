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
import { OnboardingHero } from "@/components/dashboard/onboarding-hero";
import { apiClient } from "@/lib/client";
import { Loader2 } from "lucide-react";

/**
 * The main dashboard page component.
 * Displays an overview of metrics, service health, and request volume.
 * @returns The dashboard page.
 */
export default function DashboardPage() {
  const [hasServices, setHasServices] = useState<boolean | null>(null);

  useEffect(() => {
    async function checkServices() {
        try {
            const services = await apiClient.listServices();
            // Handle both array and object response formats for robustness
            const list = Array.isArray(services) ? services : (services?.services || []);
            setHasServices(list.length > 0);
        } catch (e) {
            console.error("Failed to list services", e);
            // On error, fallback to dashboard view to avoid blocking the user
            setHasServices(true);
        }
    }
    checkServices();
  }, []);

  if (hasServices === null) {
      return (
        <div className="flex h-[calc(100vh-4rem)] items-center justify-center">
            <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
        </div>
      );
  }

  if (!hasServices) {
      return (
          <div className="flex-1 p-8 pt-6 h-[calc(100vh-4rem)] overflow-y-auto">
              <OnboardingHero />
          </div>
      )
  }

  return (
    <DashboardProvider>
      <div className="flex-1 space-y-4 p-8 pt-6">
        <div className="flex items-center justify-between space-y-2">
          <h1 className="text-3xl font-bold tracking-tight">Dashboard</h1>
          <div className="flex items-center space-x-2">
            <ServiceFilter />
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
