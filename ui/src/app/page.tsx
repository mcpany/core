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
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { ArrowRight, Sparkles } from "lucide-react";
import Link from "next/link";

/**
 * The main dashboard page component.
 * Displays an overview of metrics, service health, and request volume.
 * @returns The dashboard page.
 */
export default function DashboardPage() {
  const [hasServices, setHasServices] = useState<boolean | null>(null);

  useEffect(() => {
    const checkServices = async () => {
      try {
        const services = await apiClient.listServices();
        setHasServices(services && services.length > 0);
      } catch (e) {
        console.error("Failed to check services", e);
        // Default to showing dashboard if error, to avoid blocking
        setHasServices(true);
      }
    };
    checkServices();
  }, []);

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

        {hasServices === false && (
            <div className="mb-8 animate-in fade-in slide-in-from-top-4 duration-500">
                <Card className="bg-gradient-to-r from-primary/10 via-primary/5 to-background border-primary/20">
                    <CardHeader>
                        <div className="flex items-center gap-2 mb-2">
                            <div className="bg-primary/20 p-2 rounded-full">
                                <Sparkles className="w-5 h-5 text-primary" />
                            </div>
                            <CardTitle>Get Started with MCP Any</CardTitle>
                        </div>
                        <CardDescription className="text-base">
                            It looks like you haven't connected any services yet.
                            Use our setup wizard to quickly configure your first integration.
                        </CardDescription>
                    </CardHeader>
                    <CardContent>
                        <Link href="/setup">
                            <Button size="lg" className="gap-2 shadow-md">
                                Launch Setup Wizard <ArrowRight className="w-4 h-4" />
                            </Button>
                        </Link>
                    </CardContent>
                </Card>
            </div>
        )}

        <div className="space-y-4">
          <DashboardGrid />
        </div>
      </div>
    </DashboardProvider>
  );
}
