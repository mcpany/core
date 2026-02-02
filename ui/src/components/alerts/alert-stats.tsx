/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useEffect, useState } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { AlertCircle, CheckCircle2, AlertTriangle, Activity, Loader2 } from "lucide-react";
import { apiClient } from "@/lib/client";
import { Alert } from "./types";
import { isToday, differenceInMinutes } from "date-fns";

/**
 * AlertStats component.
 * @returns The rendered component.
 */
export function AlertStats() {
  const [stats, setStats] = useState({
    activeCritical: 0,
    activeWarning: 0,
    mttr: "N/A", // Mean Time To Resolution
    totalToday: 0
  });
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const fetchStats = async () => {
      try {
        const alerts: Alert[] = await apiClient.listAlerts();

        const activeCritical = alerts.filter(a => a.status === 'active' && a.severity === 'critical').length;
        const activeWarning = alerts.filter(a => a.status === 'active' && a.severity === 'warning').length;
        const todayAlerts = alerts.filter(a => isToday(new Date(a.timestamp)));
        const totalToday = todayAlerts.length;

        // Calculate MTTR for resolved alerts today (or all time? usually recent window)
        // Let's do all resolved alerts for simplicity or just today's resolved?
        // Feature description says "MTTR (Today)".
        const resolvedAlerts = alerts.filter(a => a.status === 'resolved' && a.resolved_at && isToday(new Date(a.resolved_at)));
        let mttr = "N/A";
        if (resolvedAlerts.length > 0) {
            const totalDuration = resolvedAlerts.reduce((acc, a) => {
                const start = new Date(a.timestamp);
                const end = new Date(a.resolved_at!);
                return acc + differenceInMinutes(end, start);
            }, 0);
            const avg = Math.round(totalDuration / resolvedAlerts.length);
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
      } finally {
        setLoading(false);
      }
    };

    fetchStats();
    // Poll every 30s?
    const interval = setInterval(fetchStats, 30000);
    return () => clearInterval(interval);
  }, []);

  if (loading) {
      return (
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
             {[1,2,3,4].map(i => (
                 <Card key={i} className="animate-pulse">
                     <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                        <div className="h-4 w-24 bg-muted rounded"></div>
                     </CardHeader>
                     <CardContent>
                        <div className="h-8 w-12 bg-muted rounded mb-2"></div>
                        <div className="h-3 w-32 bg-muted rounded"></div>
                     </CardContent>
                 </Card>
             ))}
        </div>
      );
  }

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
            Average resolution time
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
