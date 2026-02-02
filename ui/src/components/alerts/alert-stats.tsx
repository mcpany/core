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
import { isToday, parseISO } from "date-fns";

/**
 * AlertStats component.
 * @returns The rendered component.
 */
export function AlertStats() {
  const [stats, setStats] = useState({
    activeCritical: 0,
    activeWarning: 0,
    resolvedToday: 0,
    totalToday: 0
  });
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const fetchStats = async () => {
        try {
            const alerts: Alert[] = await apiClient.listAlerts();

            let critical = 0;
            let warning = 0;
            let resolved = 0;
            let total = 0;

            alerts.forEach(alert => {
                const date = parseISO(alert.timestamp);
                const today = isToday(date);

                if (today) {
                    total++;
                }

                if (alert.status === 'active') {
                    if (alert.severity === 'critical') critical++;
                    if (alert.severity === 'warning') warning++;
                }

                if (alert.status === 'resolved' && today) {
                    resolved++;
                }
            });

            setStats({
                activeCritical: critical,
                activeWarning: warning,
                resolvedToday: resolved,
                totalToday: total
            });
        } catch (error) {
            console.error("Failed to fetch alert stats", error);
        } finally {
            setLoading(false);
        }
    };

    fetchStats();
    // Poll every 30s
    const interval = setInterval(fetchStats, 30000);
    return () => clearInterval(interval);
  }, []);

  if (loading) {
      return (
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
            {[1, 2, 3, 4].map(i => (
                <Card key={i} className="animate-pulse">
                    <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                        <div className="h-4 w-24 bg-muted rounded"></div>
                        <div className="h-4 w-4 bg-muted rounded"></div>
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
             Current active warnings
          </p>
        </CardContent>
      </Card>
      <Card>
        <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
          <CardTitle className="text-sm font-medium">Resolved (Today)</CardTitle>
          <CheckCircle2 className="h-4 w-4 text-green-500" />
        </CardHeader>
        <CardContent>
          <div className="text-2xl font-bold">{stats.resolvedToday}</div>
          <p className="text-xs text-muted-foreground">
            Alerts resolved today
          </p>
        </CardContent>
      </Card>
      <Card>
        <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
          <CardTitle className="text-sm font-medium">Total Incidents (Today)</CardTitle>
          <Activity className="h-4 w-4 text-muted-foreground" />
        </CardHeader>
        <CardContent>
          <div className="text-2xl font-bold">{stats.totalToday}</div>
          <p className="text-xs text-muted-foreground">
            All alerts generated today
          </p>
        </CardContent>
      </Card>
    </div>
  );
}
