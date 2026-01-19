/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client"

import React, { useEffect, useState } from "react"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Badge } from "@/components/ui/badge"
import { AlertCircle, Activity, Globe, ShieldAlert, Clock, Terminal } from "lucide-react"

interface SystemStatus {
  uptime_seconds: number
  active_connections: number
  bound_http_port: number
  bound_grpc_port: number
  version: string
  security_warnings: string[]
}

export function SystemHealthCard() {
  const [status, setStatus] = useState<SystemStatus | null>(null)
  const [loading, setLoading] = useState(true)

  const fetchStatus = async () => {
    try {
      const response = await fetch("/api/v1/system/status")
      if (response.ok) {
        const data = await response.json()
        setStatus(data)
      }
    } catch (error) {
      console.error("Failed to fetch system status", error)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    fetchStatus()
    const interval = setInterval(fetchStatus, 5000)
    return () => clearInterval(interval)
  }, [])

  const formatUptime = (seconds: number) => {
    const hrs = Math.floor(seconds / 3600)
    const mins = Math.floor((seconds % 3600) / 60)
    const secs = seconds % 60
    return `${hrs}h ${mins}m ${secs}s`
  }

  if (loading && !status) {
    return (
      <Card className="col-span-1 shadow-sm">
        <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
          <CardTitle className="text-sm font-medium">System Health</CardTitle>
          <Activity className="h-4 w-4 text-muted-foreground animate-pulse" />
        </CardHeader>
        <CardContent>
          <div className="text-xs text-muted-foreground">Loading system status...</div>
        </CardContent>
      </Card>
    )
  }

  return (
    <Card className="col-span-1 shadow-sm border-t-4 border-t-emerald-500">
      <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
        <CardTitle className="text-sm font-medium">System Health</CardTitle>
        <Activity className="h-4 w-4 text-emerald-500" />
      </CardHeader>
      <CardContent>
        <div className="grid gap-2">
          <div className="flex items-center justify-between">
            <div className="flex items-center text-xs text-muted-foreground">
              <Clock className="mr-1 h-3 w-3" /> Uptime
            </div>
            <div className="text-xs font-mono">{status ? formatUptime(status.uptime_seconds) : "---"}</div>
          </div>
          <div className="flex items-center justify-between">
            <div className="flex items-center text-xs text-muted-foreground">
              <Globe className="mr-1 h-3 w-3" /> Connections
            </div>
            <div className="text-xs font-bold">{status?.active_connections ?? 0} active</div>
          </div>
          <div className="flex items-center justify-between border-t pt-2 mt-1">
            <div className="flex items-center text-xs text-muted-foreground">
              <Terminal className="mr-1 h-3 w-3" /> HTTP Port
            </div>
            <Badge variant="secondary" className="text-[10px] h-4">:{status?.bound_http_port}</Badge>
          </div>
          <div className="flex items-center justify-between">
            <div className="flex items-center text-xs text-muted-foreground">
              <ShieldAlert className="mr-1 h-3 w-3" /> gRPC Port
            </div>
            <Badge variant="secondary" className="text-[10px] h-4">:{status?.bound_grpc_port}</Badge>
          </div>

          {status?.security_warnings && status.security_warnings.length > 0 && (
            <div className="mt-2 p-2 bg-amber-50 rounded-md border border-amber-200">
              <div className="flex items-center text-[10px] font-bold text-amber-700 uppercase tracking-wider mb-1">
                <ShieldAlert className="mr-1 h-3 w-3" /> Security Warnings
              </div>
              <ul className="list-disc list-inside">
                {status.security_warnings.map((warning, idx) => (
                  <li key={idx} className="text-[10px] text-amber-800">{warning}</li>
                ))}
              </ul>
            </div>
          )}

          <div className="mt-2 text-[10px] text-right text-muted-foreground italic">
            v{status?.version}
          </div>
        </div>
      </CardContent>
    </Card>
  )
}
