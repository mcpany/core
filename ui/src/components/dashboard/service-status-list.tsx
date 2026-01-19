/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */


import {
  Avatar,
  AvatarFallback,
  AvatarImage,
} from "@/components/ui/avatar"
import { Badge } from "@/components/ui/badge"
import { mockServices } from "@/lib/mock-data"
import { UpstreamServiceConfig } from "@proto/config/v1/upstream_service"

export function ServiceStatusList() {
    const services = mockServices.slice(0, 5); // Show top 5

  return (
    <div className="space-y-8">
      {services.map((service: UpstreamServiceConfig) => (
        <div key={service.id} className="flex items-center">
          <Avatar className="h-9 w-9">
            <AvatarFallback>{service.name.substring(0, 2).toUpperCase()}</AvatarFallback>
          </Avatar>
          <div className="ml-4 space-y-1">
            <p className="text-sm font-medium leading-none">{service.name}</p>
            <p className="text-sm text-muted-foreground">
              {service.version}
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
