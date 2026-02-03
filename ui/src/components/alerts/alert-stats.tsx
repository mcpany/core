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

/**
 * AlertStats component.
 * @returns The rendered component.
 */
export function AlertStats() {
  const [stats, setStats] = useState({
    activeCritical: 0,
    activeWarning: 0,
    mttr: "N/A",
    totalToday: 0
  });

  useEffect(() => {
    const fetchStats = async () => {
        try {
            const alerts: Alert[] = await apiClient.listAlerts();

            let critical = 0;
            let warning = 0;
            let todayCount = 0;

            const today = new Date();
            today.setHours(0, 0, 0, 0);

            alerts.forEach((alert: Alert) => {
                if (alert.status === 'active') {
                    if (alert.severity === 'critical') critical++;
                    if (alert.severity === 'warning') warning++;
                }

                const alertDate = new Date(alert.timestamp);
                if (alertDate >= today) {
                    todayCount++;
                }
            });

            setStats({
                activeCritical: critical,
                activeWarning: warning,
                mttr: "N/A",
                totalToday: todayCount
            });
        } catch (error) {
            console.error("Failed to fetch alert stats:", error);
        }
    };

    fetchStats();
    // Refresh every 30 seconds
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
            Mean time to resolution
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
