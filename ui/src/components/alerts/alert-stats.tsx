/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useEffect, useState } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { AlertCircle, CheckCircle2, AlertTriangle, Activity } from "lucide-react";
import { apiClient } from "@/lib/client";
import { Alert } from "./types";
import { differenceInMinutes, isToday } from "date-fns";

/**
 * AlertStats component.
 * @returns The rendered component.
 */
export function AlertStats() {
  const [stats, setStats] = useState({
    activeCritical: 0,
    activeWarning: 0,
    mttr: "-",
    totalToday: 0
  });

  useEffect(() => {
    const fetchStats = async () => {
      try {
        const alerts = await apiClient.listAlerts();

        const activeCritical = alerts.filter((a: Alert) => a.severity === 'critical' && a.status === 'active').length;
        const activeWarning = alerts.filter((a: Alert) => a.severity === 'warning' && a.status === 'active').length;
        const totalToday = alerts.filter((a: Alert) => isToday(new Date(a.timestamp))).length;

        // Calculate MTTR for resolved alerts today
        const resolvedToday = alerts.filter((a: Alert) =>
            a.status === 'resolved' &&
            a.resolvedAt &&
            isToday(new Date(a.resolvedAt))
        );

        let mttr = "-";
        if (resolvedToday.length > 0) {
            const totalDuration = resolvedToday.reduce((acc: number, a: Alert) => {
                if (!a.resolvedAt) return acc;
                const duration = differenceInMinutes(new Date(a.resolvedAt), new Date(a.timestamp));
                return acc + duration;
            }, 0);
            const avg = Math.round(totalDuration / resolvedToday.length);
            mttr = `${avg}m`;
        }

        setStats({
            activeCritical,
            activeWarning,
            mttr,
            totalToday
        });

      } catch (error) {
        console.error("Failed to fetch alert stats", error);
      }
    };

    fetchStats();
    // Poll every 30s
    const interval = setInterval(fetchStats, 30000);
    return () => clearInterval(interval);
  }, []);

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
             Current active warning alerts
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
             Alerts triggered today
          </p>
        </CardContent>
      </Card>
    </div>
  );
}
