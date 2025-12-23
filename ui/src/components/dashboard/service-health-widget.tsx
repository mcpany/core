/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useEffect, useState } from "react";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { CheckCircle2, XCircle, AlertCircle } from "lucide-react";

interface ServiceStatus {
  id?: string;
  name: string;
  status: "healthy" | "unhealthy" | "degraded";
  uptime: string;
  version: string;
}

export function ServiceHealthWidget() {
  const [services, setServices] = useState<ServiceStatus[]>([]);

  useEffect(() => {
    async function fetchServices() {
      try {
        const res = await fetch("/api/dashboard/health");
        if (res.ok) {
          const data = await res.json();
          setServices(data);
        }
      } catch (error) {
        console.error("Failed to fetch service health", error);
      }
    }
    fetchServices();
    const interval = setInterval(fetchServices, 5000);
    return () => clearInterval(interval);
  }, []);

  return (
    <Card className="col-span-3 backdrop-blur-sm bg-background/50">
      <CardHeader>
        <CardTitle>Service Health</CardTitle>
        <CardDescription>Real-time status of connected upstream services.</CardDescription>
      </CardHeader>
      <CardContent>
        <div className="space-y-4">
          {services.map((service, index) => (
            <div key={service.id || index} className="flex items-center justify-between p-2 rounded-lg hover:bg-muted/50 transition-colors">
              <div className="flex items-center space-x-4">
                {service.status === "healthy" && <CheckCircle2 className="text-green-500 h-5 w-5" />}
                {service.status === "degraded" && <AlertCircle className="text-yellow-500 h-5 w-5" />}
                {service.status === "unhealthy" && <XCircle className="text-red-500 h-5 w-5" />}
                <div>
                  <p className="text-sm font-medium leading-none">{service.name}</p>
                  <p className="text-xs text-muted-foreground">{service.version}</p>
                </div>
              </div>
              <div className="flex items-center space-x-4">
                <div className="text-right">
                    <p className="text-sm font-medium">{service.uptime}</p>
                    <p className="text-xs text-muted-foreground">Uptime</p>
                </div>
                <Badge variant={service.status === "healthy" ? "default" : service.status === "degraded" ? "secondary" : "destructive"}>
                    {service.status}
                </Badge>
              </div>
            </div>
          ))}
        </div>
      </CardContent>
    </Card>
  );
}
