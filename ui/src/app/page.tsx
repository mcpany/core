/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useEffect, useState } from "react";
import { apiClient } from "@/lib/client";
import { DashboardGrid } from "@/components/dashboard/dashboard-grid";
import { Button } from "@/components/ui/button";
import { DashboardProvider } from "@/components/dashboard/dashboard-context";
import { ServiceFilter } from "@/components/dashboard/service-filter";
import { TimeRangeFilter } from "@/components/dashboard/time-range-filter";
import { Sparkles, ArrowRight, Loader2 } from "lucide-react";
import Link from "next/link";

/**
 * The main dashboard page component.
 * Displays an overview of metrics, service health, and request volume.
 * @returns The dashboard page.
 */
export default function DashboardPage() {
  const [hasServices, setHasServices] = useState<boolean | null>(null);

  useEffect(() => {
    apiClient.listServices().then(services => {
        setHasServices(services && services.length > 0);
    }).catch((e) => {
        console.error("Failed to check services", e);
        // Fallback to true (show dashboard) on error so user isn't stuck
        setHasServices(true);
    });
  }, []);

  if (hasServices === null) {
      return (
          <div className="flex items-center justify-center h-[calc(100vh-100px)]">
              <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
          </div>
      );
  }

  if (hasServices === false) {
      return (
        <div className="flex-1 space-y-4 p-8 pt-6 flex flex-col items-center justify-center min-h-[60vh]">
             <div className="w-full max-w-lg text-center space-y-6">
                <div className="mx-auto bg-primary/10 p-6 rounded-full w-fit mb-4 animate-pulse">
                    <Sparkles className="w-12 h-12 text-primary" />
                </div>
                <h1 className="text-4xl font-bold tracking-tight">Welcome to MCP Any</h1>
                <p className="text-xl text-muted-foreground">
                    Your universal gateway is ready. Connect your first service to get started.
                </p>
                <div className="pt-4">
                    <Link href="/setup">
                        <Button size="lg" className="w-full max-w-sm text-lg h-12 gap-2 shadow-lg hover:shadow-xl transition-all">
                        Run Setup Wizard <ArrowRight className="w-5 h-5" />
                        </Button>
                    </Link>
                </div>
            </div>
        </div>
      );
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
