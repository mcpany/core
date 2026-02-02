/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { AlertCircle, CheckCircle2, AlertTriangle, Activity, Loader2 } from "lucide-react";
import { useEffect, useState } from "react";
import { apiClient } from "@/lib/client";

interface AlertStats {
  activeCritical: number;
  activeWarning: number;
  mttr: string;
  totalToday: number;
}

/**
 * AlertStats component.
 * @returns The rendered component.
 */
export function AlertStats() {
  const [stats, setStats] = useState<AlertStats | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const fetchStats = async () => {
      try {
        const data = await apiClient.getAlertStats();
        setStats(data);
      } catch (error) {
        console.error("Failed to load alert stats:", error);
      } finally {
        setLoading(false);
      }
    };
    fetchStats();
  }, []);

  if (loading) {
     return <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4 h-[120px] items-center justify-center">
         <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
     </div>;
  }

  const displayStats = stats || {
    activeCritical: 0,
    activeWarning: 0,
    mttr: "0m",
    totalToday: 0
  };

  return (
    <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
      <Card>
        <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
          <CardTitle className="text-sm font-medium">Active Critical</CardTitle>
          <AlertCircle className="h-4 w-4 text-red-500" />
        </CardHeader>
        <CardContent>
          <div className="text-2xl font-bold text-red-500">{displayStats.activeCritical}</div>
          {/* Trends disabled for now */}
          {/* <p className="text-xs text-muted-foreground">
            +1 since last hour
          </p> */}
        </CardContent>
      </Card>
      <Card>
        <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
          <CardTitle className="text-sm font-medium">Active Warnings</CardTitle>
          <AlertTriangle className="h-4 w-4 text-yellow-500" />
        </CardHeader>
        <CardContent>
          <div className="text-2xl font-bold text-yellow-500">{displayStats.activeWarning}</div>
          {/* <p className="text-xs text-muted-foreground">
            -2 since last hour
          </p> */}
        </CardContent>
      </Card>
      <Card>
        <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
          <CardTitle className="text-sm font-medium">MTTR (Today)</CardTitle>
          <CheckCircle2 className="h-4 w-4 text-green-500" />
        </CardHeader>
        <CardContent>
          <div className="text-2xl font-bold">{displayStats.mttr}</div>
          {/* <p className="text-xs text-muted-foreground">
            -2m from yesterday
          </p> */}
        </CardContent>
      </Card>
      <Card>
        <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
          <CardTitle className="text-sm font-medium">Total Incidents</CardTitle>
          <Activity className="h-4 w-4 text-muted-foreground" />
        </CardHeader>
        <CardContent>
          <div className="text-2xl font-bold">{displayStats.totalToday}</div>
          {/* <p className="text-xs text-muted-foreground">
            +12% from average
          </p> */}
        </CardContent>
      </Card>
    </div>
  );
}
