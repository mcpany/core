/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { DashboardGrid } from "@/components/dashboard/dashboard-grid";
import { Button } from "@/components/ui/button";
import { DashboardProvider } from "@/components/dashboard/dashboard-context";
import { ServiceFilter } from "@/components/dashboard/service-filter";
import { TimeRangeFilter } from "@/components/dashboard/time-range-filter";
import { apiClient } from "@/lib/client";
import { WelcomeScreen } from "@/components/onboarding/welcome-screen";
import { RegisterServiceDialog } from "@/components/register-service-dialog";
import { Loader2 } from "lucide-react";

/**
 * The main dashboard page component.
 * Displays an overview of metrics, service health, and request volume.
 * Includes a Welcome Screen for zero-state onboarding.
 * @returns The dashboard page.
 */
export default function DashboardPage() {
  const [loading, setLoading] = useState(true);
  const [hasServices, setHasServices] = useState(false);
  const [registerOpen, setRegisterOpen] = useState(false);
  const router = useRouter();

  const checkServices = async () => {
    try {
      const services = await apiClient.listServices();
      if (services && services.length > 0) {
        setHasServices(true);
      } else {
        setHasServices(false);
      }
    } catch (e) {
      console.error("Failed to check services", e);
      // Fallback to dashboard if error, so user isn't stuck on loading or empty welcome if broken
      setHasServices(true);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    checkServices();
  }, []);

  const handleServiceAdded = () => {
      setRegisterOpen(false);
      setLoading(true);
      checkServices();
  };

  if (loading) {
      return (
          <div className="flex h-[calc(100vh-4rem)] items-center justify-center">
              <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
          </div>
      );
  }

  if (!hasServices) {
      return (
          <>
            <WelcomeScreen
                onTemplate={() => router.push('/marketplace')}
                onRegister={() => setRegisterOpen(true)}
            />
            <RegisterServiceDialog
                open={registerOpen}
                onOpenChange={setRegisterOpen}
                onSuccess={handleServiceAdded}
            />
          </>
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
