/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import React, { useState } from 'react';
import { useSystemHealth } from '@/hooks/use-system-health';
import { CheckCircle, AlertTriangle, XCircle, RefreshCw } from 'lucide-react';
import { cn } from '@/lib/utils';
import { Button } from '@/components/ui/button';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
} from "@/components/ui/dialog";

export function SystemHealthIndicator() {
  const { report, loading, error, refresh } = useSystemHealth();
  const [isOpen, setIsOpen] = useState(false);

  // If initial load, maybe show nothing or a skeleton?
  // But let's just show a discreet dot.

  let statusColor = 'bg-gray-400';
  let Icon = RefreshCw;
  let label = 'Connecting...';

  if (error) {
    statusColor = 'bg-red-500';
    Icon = XCircle;
    label = 'Connection Error';
  } else if (report) {
    if (report.status === 'healthy' || report.status === 'ok') {
      statusColor = 'bg-green-500';
      Icon = CheckCircle;
      label = 'System Healthy';
    } else {
      statusColor = 'bg-yellow-500';
      Icon = AlertTriangle;
      label = 'System Degraded';
    }
  }

  return (
    <>
      <button
        onClick={() => setIsOpen(true)}
        className="flex items-center gap-2 px-3 py-1.5 rounded-full hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors text-xs font-medium border border-transparent hover:border-gray-200 dark:hover:border-gray-700"
      >
        <span className={cn("relative flex h-2.5 w-2.5")}>
          <span className={cn("animate-ping absolute inline-flex h-full w-full rounded-full opacity-75", statusColor)}></span>
          <span className={cn("relative inline-flex rounded-full h-2.5 w-2.5", statusColor)}></span>
        </span>
        <span className="text-gray-600 dark:text-gray-300 hidden md:inline">{label}</span>
      </button>

      <Dialog open={isOpen} onOpenChange={setIsOpen}>
        <DialogContent className="sm:max-w-md">
            <DialogHeader>
                <DialogTitle className="flex items-center gap-2">
                    <Icon className={cn("h-5 w-5", error ? "text-red-500" : report?.status === 'healthy' ? "text-green-500" : "text-yellow-500")} />
                    System Status
                </DialogTitle>
                <DialogDescription>
                    Real-time connectivity status of the MCP Any server.
                </DialogDescription>
            </DialogHeader>
          <div className="space-y-4">
            <div className="flex items-center justify-between">
                <span className="text-sm text-gray-500">Last updated: {report ? new Date(report.timestamp).toLocaleTimeString() : '-'}</span>
                <Button variant="ghost" size="sm" onClick={refresh} disabled={loading}>
                    <RefreshCw className={cn("h-4 w-4 mr-1", loading && "animate-spin")} />
                    Refresh
                </Button>
            </div>

            {error && (
              <div className="p-3 bg-red-50 dark:bg-red-900/20 text-red-600 dark:text-red-400 rounded-md text-sm">
                Failed to connect to server: {error.message}
              </div>
            )}

            {report && (
              <div className="space-y-2">
                {Object.entries(report.checks).map(([name, result]) => (
                  <div key={name} className="flex items-center justify-between p-2 rounded-md bg-gray-50 dark:bg-gray-800/50 border border-gray-100 dark:border-gray-700">
                    <div className="flex flex-col">
                        <span className="font-medium text-sm capitalize">{name.replace(/_/g, ' ')}</span>
                        {result.message && <span className="text-xs text-red-500">{result.message}</span>}
                    </div>
                    <div className="flex items-center gap-2">
                        {result.latency && <span className="text-xs text-gray-400 font-mono">{result.latency}</span>}
                        <StatusBadge status={result.status} />
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>
        </DialogContent>
      </Dialog>
    </>
  );
}

function StatusBadge({ status }: { status: string }) {
  if (status === 'ok' || status === 'healthy') {
    return <span className="px-2 py-0.5 rounded-full bg-green-100 dark:bg-green-900/30 text-green-700 dark:text-green-400 text-[10px] font-bold uppercase tracking-wider">OK</span>;
  }
  return <span className="px-2 py-0.5 rounded-full bg-red-100 dark:bg-red-900/30 text-red-700 dark:text-red-400 text-[10px] font-bold uppercase tracking-wider">{status}</span>;
}
