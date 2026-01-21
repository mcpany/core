/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */


import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Activity, Server, AlertCircle, Clock } from "lucide-react"

interface MetricCardProps {
    /** The title of the metric card. */
    title: string
    /** The value to display (e.g., number of requests, latency). */
    value: string | number
    /** The icon component to display in the header. */
    icon: React.ElementType
    /** A description or subtitle for the metric. */
    description: string
    /** Optional trend indicator (e.g., "+5%"). currently unused in UI but reserved for future. */
    trend?: string
}

/**
 * MetricCard displays a single metric in a card format.
 * It shows a title, a value, an icon, and a description.
 *
 * @param props - The properties for the metric card.
 */
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
