/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { AlertCircle, CheckCircle2, AlertTriangle, Activity } from "lucide-react";
import { Alert } from "./types";

interface AlertStatsProps {
    alerts: Alert[];
}

/**
 * AlertStats component.
 * @param props.alerts The list of alerts.
 * @returns The rendered component.
 */
export function AlertStats({ alerts }: AlertStatsProps) {

  const activeCritical = alerts.filter(a => a.severity === "critical" && a.status === "active").length;
  const activeWarning = alerts.filter(a => a.severity === "warning" && a.status === "active").length;

  const now = new Date();
  const today = new Date(now.getFullYear(), now.getMonth(), now.getDate());
  const totalToday = alerts.filter(a => new Date(a.timestamp) >= today).length;

  // Mocking MTTR as "N/A" for now unless we track resolution time
  const mttr = "N/A";

  return (
    <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
      <Card>
        <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
          <CardTitle className="text-sm font-medium">Active Critical</CardTitle>
          <AlertCircle className="h-4 w-4 text-red-500" />
        </CardHeader>
        <CardContent>
          <div className="text-2xl font-bold text-red-500">{activeCritical}</div>
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
          <div className="text-2xl font-bold text-yellow-500">{activeWarning}</div>
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
          <div className="text-2xl font-bold">{mttr}</div>
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
          <div className="text-2xl font-bold">{totalToday}</div>
          <p className="text-xs text-muted-foreground">
            Recorded today
          </p>
        </CardContent>
      </Card>
    </div>
  );
}
