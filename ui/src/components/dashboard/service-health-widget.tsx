/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useEffect, useState } from "react";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { CheckCircle2, AlertTriangle, XCircle } from "lucide-react";

interface ServiceHealth {
  id: string;
  name: string;
  status: "healthy" | "degraded" | "unhealthy";
  latency: string;
  uptime: string;
}

export function ServiceHealthWidget() {
  const [services, setServices] = useState<ServiceHealth[]>([]);

  useEffect(() => {
    async function fetchHealth() {
      try {
        const res = await fetch("/api/dashboard/health");
        if (res.ok) {
          const data = await res.json();
          setServices(data);
        }
      } catch (error) {
        console.error("Failed to fetch health data", error);
      }
    }
    fetchHealth();
    const interval = setInterval(fetchHealth, 10000); // Poll every 10s
    return () => clearInterval(interval);
  }, []);

  const getStatusIcon = (status: string) => {
    switch (status) {
      case "healthy":
        return <CheckCircle2 className="h-4 w-4 text-green-500" />;
      case "degraded":
        return <AlertTriangle className="h-4 w-4 text-yellow-500" />;
      case "unhealthy":
        return <XCircle className="h-4 w-4 text-red-500" />;
      default:
        return <AlertTriangle className="h-4 w-4 text-gray-500" />;
    }
  };

  const getStatusColor = (status: string) => {
      switch (status) {
        case "healthy": return "bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-300";
        case "degraded": return "bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-300";
        case "unhealthy": return "bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-300";
        default: return "bg-gray-100 text-gray-800";
      }
  };

  return (
    <Card className="col-span-4 backdrop-blur-sm bg-background/50">
      <CardHeader>
        <CardTitle>System Health</CardTitle>
        <CardDescription>
          Real-time status of critical services.
        </CardDescription>
      </CardHeader>
      <CardContent>
        <div className="space-y-4">
          {services.map((service) => (
            <div
              key={service.id}
              className="flex items-center justify-between space-x-4 border-b pb-4 last:border-0 last:pb-0"
            >
              <div className="flex items-center space-x-4">
                {getStatusIcon(service.status)}
                <div>
                  <p className="text-sm font-medium leading-none">{service.name}</p>
                  <p className="text-sm text-muted-foreground">Latency: {service.latency}</p>
                </div>
              </div>
              <div className="flex items-center space-x-2">
                 <Badge variant="outline" className={getStatusColor(service.status)}>
                    {service.status}
                 </Badge>
                 <span className="text-xs text-muted-foreground w-12 text-right">{service.uptime}</span>
              </div>
            </div>
          ))}
        </div>
      </CardContent>
    </Card>
  );
}
