/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { AlertCircle, CheckCircle2, AlertTriangle, Activity } from "lucide-react";
import { Alert } from "./types";
import { useMemo } from "react";
import { isToday } from "date-fns";

interface AlertStatsProps {
    alerts: Alert[];
}

/**
 * AlertStats component.
 * @param props - The component props.
 * @param props.alerts - The list of alerts.
 * @returns The rendered component.
 */
export function AlertStats({ alerts }: AlertStatsProps) {

  const stats = useMemo(() => {
    const activeCritical = alerts.filter(a => a.status === 'active' && a.severity === 'critical').length;
    const activeWarning = alerts.filter(a => a.status === 'active' && a.severity === 'warning').length;

    // Total incidents created today
    const totalToday = alerts.filter(a => isToday(new Date(a.timestamp))).length;

    // MTTR calculation (for alerts resolved today)
    const resolvedToday = alerts.filter(a =>
        a.status === 'resolved' &&
        a.resolvedAt &&
        isToday(new Date(a.resolvedAt))
    );

    let mttr = "N/A";
    if (resolvedToday.length > 0) {
        const totalDurationMs = resolvedToday.reduce((sum, a) => {
            const start = new Date(a.timestamp).getTime();
            const end = new Date(a.resolvedAt!).getTime();
            return sum + (end - start);
        }, 0);
        const avgMs = totalDurationMs / resolvedToday.length;
        const avgMinutes = Math.round(avgMs / 60000);
        mttr = `${avgMinutes}m`;
    }

    return {
        activeCritical,
        activeWarning,
        mttr,
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
             Alerts requiring immediate attention
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
             Potential issues to investigate
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
            Mean Time To Resolution
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
            Reported today
          </p>
        </CardContent>
      </Card>
    </div>
  );
}
