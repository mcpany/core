/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client"

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { ScrollArea } from "@/components/ui/scroll-area"
import { Badge } from "@/components/ui/badge"

const services = [
  { name: "github-service", status: "online", uptime: "99.9%" },
  { name: "slack-integration", status: "online", uptime: "99.8%" },
  { name: "linear-connector", status: "degraded", uptime: "95.5%" },
  { name: "postgres-db", status: "online", uptime: "100%" },
  { name: "openai-proxy", status: "online", uptime: "99.9%" },
]

export function ServiceHealthWidget() {
  return (
    <Card className="col-span-3">
      <CardHeader>
        <CardTitle>Service Health</CardTitle>
        <CardDescription>Live status of connected MCP servers.</CardDescription>
      </CardHeader>
      <CardContent>
        <ScrollArea className="h-[300px]">
          <div className="space-y-4">
            {services.map((service) => (
              <div key={service.name} className="flex items-center justify-between border-b pb-2 last:border-0">
                <div className="flex flex-col">
                  <span className="font-medium">{service.name}</span>
                  <span className="text-xs text-muted-foreground">Uptime: {service.uptime}</span>
                </div>
                <Badge variant={service.status === "online" ? "default" : "destructive"}>
                  {service.status}
                </Badge>
              </div>
            ))}
          </div>
        </ScrollArea>
      </CardContent>
    </Card>
  )
}
