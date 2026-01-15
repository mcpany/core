/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { AlertCircle } from "lucide-react"
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert"
import { useDoctor } from "@/hooks/useDoctor"

export function ConfigStatusBanner() {
  const { report } = useDoctor(5000); // Poll every 5 seconds

  if (!report) return null;

  const configCheck = report.checks['configuration'];

  // Only show if the configuration check is present and degraded
  if (!configCheck || configCheck.status !== 'degraded') {
    return null;
  }

  return (
    <div className="px-4 pt-4">
      <Alert variant="destructive">
        <AlertCircle className="h-4 w-4" />
        <AlertTitle>Configuration Error</AlertTitle>
        <AlertDescription>
          The server is running with a stale configuration because the last reload failed.
          <div className="mt-2 font-mono text-xs bg-destructive/10 p-2 rounded whitespace-pre-wrap">
            {configCheck.message || "Unknown error"}
          </div>
        </AlertDescription>
      </Alert>
    </div>
  )
}
