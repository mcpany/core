/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useEffect, useState } from "react";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { AlertTriangle, WifiOff } from "lucide-react";

interface CheckResult {
  status: string;
  message?: string;
  latency?: string;
}

interface DoctorReport {
  status: string;
  timestamp: string;
  checks: Record<string, CheckResult>;
}

export function SystemStatusBanner() {
  const [report, setReport] = useState<DoctorReport | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchHealth = async () => {
      try {
        const res = await fetch("/doctor");
        if (!res.ok) {
          throw new Error(`Failed to fetch health status: ${res.statusText}`);
        }
        const data = await res.json();
        setReport(data);
        setError(null);
      } catch (err) {
        setError(err instanceof Error ? err.message : "Unknown error");
        setReport(null);
      }
    };

    fetchHealth();
    const interval = setInterval(fetchHealth, 30000); // Poll every 30s
    return () => clearInterval(interval);
  }, []);

  if (error) {
    return (
      <div className="p-4 pb-0">
        <Alert variant="destructive">
            <WifiOff className="h-4 w-4" />
            <AlertTitle>Connection Error</AlertTitle>
            <AlertDescription>
            Could not connect to the server health check. Is the backend running? ({error})
            </AlertDescription>
        </Alert>
      </div>
    );
  }

  if (!report || report.status === "healthy" || report.status === "ok") {
    return null;
  }

  // Degraded state
  const issues: string[] = [];
  if (report.checks) {
      Object.entries(report.checks).forEach(([name, check]) => {
          if (check.status !== "ok") {
            // Capitalize first letter of name
            const niceName = name.charAt(0).toUpperCase() + name.slice(1);
            issues.push(`${niceName}: ${check.message || "Unknown issue"}`);
          }
      });
  }

  return (
    <div className="p-4 pb-0">
        <Alert className="border-yellow-500/50 text-yellow-600 dark:border-yellow-500 dark:text-yellow-400 [&>svg]:text-yellow-600 dark:[&>svg]:text-yellow-400">
        <AlertTriangle className="h-4 w-4" />
        <AlertTitle>System Status: Degraded</AlertTitle>
        <AlertDescription>
            The server is running but encountered the following issues:
            <ul className="list-disc pl-4 mt-1 space-y-1">
            {issues.map((issue, i) => (
                <li key={i}>{issue}</li>
            ))}
            </ul>
        </AlertDescription>
        </Alert>
    </div>
  );
}
