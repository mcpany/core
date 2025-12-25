/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */


import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Activity, Server, AlertCircle, Clock } from "lucide-react"

interface MetricCardProps {
    title: string
    value: string | number
    icon: React.ElementType
    description: string
    trend?: string
}

export function MetricCard({ title, value, icon: Icon, description, trend }: MetricCardProps) {
  return (
    <Card>
      <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
        <CardTitle className="text-sm font-medium">
          {title}
        </CardTitle>
        <Icon className="h-4 w-4 text-muted-foreground" />
      </CardHeader>
      <CardContent>
        <div className="text-2xl font-bold">{value}</div>
        <p className="text-xs text-muted-foreground">
          {description}
        </p>
      </CardContent>
    </Card>
  )
}
