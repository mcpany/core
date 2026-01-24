/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useEffect, useState } from "react";
import {
  Avatar,
  AvatarFallback,
  AvatarImage,
} from "@/components/ui/avatar"
import { Badge } from "@/components/ui/badge"
import { apiClient } from "@/lib/client"
import { UpstreamServiceConfig } from "@proto/config/v1/upstream_service"

/**
 * ServiceStatusList component.
 * @returns The rendered component.
 */
export function ServiceStatusList() {
    const [services, setServices] = useState<UpstreamServiceConfig[]>([]);

    useEffect(() => {
        const fetchServices = async () => {
            try {
                const data = await apiClient.listServices();
                setServices(data.slice(0, 5));
            } catch (error) {
                console.error("Failed to fetch services", error);
            }
        };
        fetchServices();
    }, []);

  return (
    <div className="space-y-8">
      {services.map((service: UpstreamServiceConfig) => (
        <div key={service.id || service.name} className="flex items-center">
          <Avatar className="h-9 w-9">
            <AvatarFallback>{service.name.substring(0, 2).toUpperCase()}</AvatarFallback>
          </Avatar>
          <div className="ml-4 space-y-1">
            <p className="text-sm font-medium leading-none">{service.name}</p>
            <p className="text-sm text-muted-foreground">
              {service.version || "latest"}
            </p>
          </div>
          <div className="ml-auto font-medium">
             <Badge variant={service.disable ? "destructive" : "default"}>
                {service.disable ? "Disabled" : "Active"}
             </Badge>
          </div>
        </div>
      ))}
    </div>
  )
}
