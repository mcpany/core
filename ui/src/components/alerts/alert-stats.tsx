/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { AlertCircle, CheckCircle2, AlertTriangle, Activity } from "lucide-react";
import { Alert } from "./types";
import { useMemo } from "react";

interface AlertStatsProps {
  alerts: Alert[];
}

/**
 * AlertStats component.
 * @param props - The component props.
 * @returns The rendered component.
 */
export function AlertStats({ alerts }: AlertStatsProps) {
  const stats = useMemo(() => {
    const activeCritical = alerts.filter(a => a.severity === "critical" && a.status === "active").length;
    const activeWarning = alerts.filter(a => a.severity === "warning" && a.status === "active").length;

    // Total today
    const now = new Date();
    const today = new Date(now.getFullYear(), now.getMonth(), now.getDate());
    const totalToday = alerts.filter(a => new Date(a.timestamp) >= today).length;

    return {
        activeCritical,
        activeWarning,
        mttr: "N/A",
        totalToday
    };
  }, [alerts]);

  return (
    <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
      <Card>
        <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
          <CardTitle className="text-sm font-medium">Active Critical</CardTitle>
          <AlertCircle className="h-4 w-4 text-red-500" />
        </CardHeader>
        <CardContent>
          <div className="text-2xl font-bold text-red-500">{stats.activeCritical}</div>
          <p className="text-xs text-muted-foreground">
            Current active critical alerts
          </p>
        </CardContent>
      </Card>
      <Card>
        <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
          <CardTitle className="text-sm font-medium">Active Warnings</CardTitle>
          <AlertTriangle className="h-4 w-4 text-yellow-500" />
        </CardHeader>
        <CardContent>
          <div className="text-2xl font-bold text-yellow-500">{stats.activeWarning}</div>
          <p className="text-xs text-muted-foreground">
            Current active warnings
          </p>
        </CardContent>
      </Card>
      <Card>
        <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
          <CardTitle className="text-sm font-medium">MTTR (Today)</CardTitle>
          <CheckCircle2 className="h-4 w-4 text-green-500" />
        </CardHeader>
        <CardContent>
          <div className="text-2xl font-bold">{stats.mttr}</div>
          <p className="text-xs text-muted-foreground">
            Avg resolution time
          </p>
        </CardContent>
      </Card>
      <Card>
        <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
          <CardTitle className="text-sm font-medium">Total Incidents</CardTitle>
          <Activity className="h-4 w-4 text-muted-foreground" />
        </CardHeader>
        <CardContent>
          <div className="text-2xl font-bold">{stats.totalToday}</div>
          <p className="text-xs text-muted-foreground">
            Recorded today
          </p>
        </CardContent>
      </Card>
    </div>
  );
}
