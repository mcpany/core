/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useEffect, useState } from "react";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { AlertTriangle, WifiOff, AlertCircle } from "lucide-react";
import { apiClient, DoctorReport } from "@/lib/client";

/**
 * SystemStatusBanner component.
 * @returns The rendered component.
 */
export function SystemStatusBanner() {
  const [report, setReport] = useState<DoctorReport | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchHealth = async () => {
      try {
        const data = await apiClient.getDoctorStatus();
        setReport(data);
        setError(null);
      } catch (err) {
        // Fail silently for network errors to avoid spamming the user if the server is just restarting
        // But if we want to show connection error like before:
        setError(err instanceof Error ? err.message : "Unknown error");
        setReport(null);
      }
    };

    fetchHealth();
    const interval = setInterval(fetchHealth, 5000); // 30s might be too slow for config updates, using 5s like ConfigBanner
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
    // Double check specific configuration status even if overall says ok (unlikely but safe)
    // Actually if overall is ok, config must be ok.
    return null;
  }

  // Check for specific configuration error to give it special treatment
  const configCheck = report.checks?.configuration;
  if (configCheck && configCheck.status !== "ok") {
      return (
        <div className="p-4 pb-0">
            <Alert variant="destructive">
            <AlertCircle className="h-4 w-4" />
            <AlertTitle>Configuration Error</AlertTitle>
            <AlertDescription className="flex flex-col gap-2">
                <p>
                The server configuration failed to reload. The server is running with a stale configuration.
                </p>
                {configCheck.message && (
                    <p className="font-mono text-xs bg-black/10 p-2 rounded whitespace-pre-wrap">
                    {configCheck.message}
                    </p>
                )}
                {configCheck.diff && (
                    <div className="mt-2">
                        <p className="text-xs font-semibold mb-1">Configuration Diff:</p>
                        <pre className="font-mono text-[10px] leading-tight bg-black/10 p-2 rounded overflow-x-auto whitespace-pre">
                            {configCheck.diff}
                        </pre>
                    </div>
                )}
            </AlertDescription>
            </Alert>
        </div>
      );
  }

  // Fallback for other degraded states
  const issues: string[] = [];
  if (report.checks) {
      Object.entries(report.checks).forEach(([name, check]) => {
          // Explicit cast or ensure type is correct
          // report.checks is Record<string, DoctorCheckResult> in client.ts
          // But I imported DoctorReport from client.ts which uses DoctorCheckResult
          const c = check as any; // Temporary fix or better type assertion
          if (c.status !== "ok") {
            // Capitalize first letter of name
            const niceName = name.charAt(0).toUpperCase() + name.slice(1);
            issues.push(`${niceName}: ${c.message || "Unknown issue"}`);
          }
      });
  }

  if (issues.length === 0) return null;

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
