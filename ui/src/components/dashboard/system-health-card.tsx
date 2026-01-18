/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useEffect, useState } from "react";
import {
  Activity,
  ShieldAlert,
  ShieldCheck,
  Server,
  Network,
  Clock,
  GitCommit
} from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { SystemStatus } from "@/types/system-status";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";

function formatUptime(seconds: number): string {
  const days = Math.floor(seconds / (3600 * 24));
  const hours = Math.floor((seconds % (3600 * 24)) / 3600);
  const minutes = Math.floor((seconds % 3600) / 60);

  const parts = [];
  if (days > 0) parts.push(`${days}d`);
  if (hours > 0) parts.push(`${hours}h`);
  if (minutes > 0) parts.push(`${minutes}m`);
  if (parts.length === 0) return `${Math.floor(seconds)}s`;

  return parts.join(" ");
}

export function SystemHealthCard() {
  const [status, setStatus] = useState<SystemStatus | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    async function fetchStatus() {
      try {
        const res = await fetch("/api/v1/system/status");
        if (res.ok) {
          const data = await res.json();
          setStatus(data);
        }
      } catch (error) {
        console.error("Failed to fetch system status", error);
      } finally {
        setLoading(false);
      }
    }
    fetchStatus();
    const interval = setInterval(fetchStatus, 5000);
    return () => clearInterval(interval);
  }, []);

  if (loading) {
    return (
        <Card className="col-span-full animate-pulse">
            <CardHeader><CardTitle>System Health</CardTitle></CardHeader>
            <CardContent>Loading system status...</CardContent>
        </Card>
    );
  }

  if (!status) {
      return (
          <Card className="col-span-full border-red-500">
              <CardHeader><CardTitle className="text-red-500">System Offline</CardTitle></CardHeader>
              <CardContent>Failed to retrieve system status.</CardContent>
          </Card>
      );
  }

  const hasWarnings = status.security_warnings.length > 0;

  return (
    <Card className="col-span-full lg:col-span-3">
      <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
        <CardTitle className="text-base font-semibold">System Health</CardTitle>
        <div className="flex items-center space-x-2">
            {hasWarnings ? (
                <ShieldAlert className="h-5 w-5 text-yellow-500" />
            ) : (
                <ShieldCheck className="h-5 w-5 text-green-500" />
            )}
        </div>
      </CardHeader>
      <CardContent className="space-y-4">
        {hasWarnings && (
            <Alert variant="destructive" className="bg-yellow-500/10 text-yellow-600 border-yellow-500/50">
                <ShieldAlert className="h-4 w-4" />
                <AlertTitle>Security Warning</AlertTitle>
                <AlertDescription>
                    {status.security_warnings.map((w, i) => (
                        <div key={i}>{w}</div>
                    ))}
                </AlertDescription>
            </Alert>
        )}

        <div className="grid grid-cols-2 gap-4">
            <div className="flex flex-col space-y-1">
                <span className="text-xs text-muted-foreground flex items-center">
                    <Clock className="h-3 w-3 mr-1" /> Uptime
                </span>
                <span className="font-mono text-sm">{formatUptime(status.uptime_seconds)}</span>
            </div>
            <div className="flex flex-col space-y-1">
                <span className="text-xs text-muted-foreground flex items-center">
                    <Network className="h-3 w-3 mr-1" /> Active Connections
                </span>
                <span className="font-mono text-sm">{status.active_connections}</span>
            </div>
            <div className="flex flex-col space-y-1">
                <span className="text-xs text-muted-foreground flex items-center">
                    <Server className="h-3 w-3 mr-1" /> Ports (HTTP/gRPC)
                </span>
                <span className="font-mono text-sm">{status.bound_http_port} / {status.bound_grpc_port > 0 ? status.bound_grpc_port : "N/A"}</span>
            </div>
            <div className="flex flex-col space-y-1">
                <span className="text-xs text-muted-foreground flex items-center">
                    <GitCommit className="h-3 w-3 mr-1" /> Version
                </span>
                <span className="font-mono text-sm">{status.version}</span>
            </div>
        </div>
      </CardContent>
    </Card>
  );
}
