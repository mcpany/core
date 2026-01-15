/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client"

import * as React from "react"
import { AlertCircle } from "lucide-react"

import {
  Alert,
  AlertDescription,
  AlertTitle,
} from "@/components/ui/alert"
import { apiClient } from "@/lib/client"

export function ConfigHealthBanner() {
  const [error, setError] = React.useState<string | null>(null)

  const checkHealth = React.useCallback(async () => {
    try {
      const report = await apiClient.getDoctorStatus()
      const configCheck = report.checks?.configuration

      if (configCheck && configCheck.status !== "ok") {
        setError(configCheck.message || "Configuration is in a degraded state.")
      } else {
        setError(null)
      }
    } catch (e) {
      // Fail silently for network errors to avoid spamming the user if the server is just restarting
      console.error("Failed to check system health", e)
    }
  }, [])

  React.useEffect(() => {
    checkHealth()
    const interval = setInterval(checkHealth, 5000)
    return () => clearInterval(interval)
  }, [checkHealth])

  if (!error) {
    return null
  }

  return (
    <div className="p-4 pb-0">
        <Alert variant="destructive">
        <AlertCircle className="h-4 w-4" />
        <AlertTitle>Configuration Error</AlertTitle>
        <AlertDescription className="flex flex-col gap-2">
            <p>
            The server configuration failed to reload. The server is running with a stale configuration.
            </p>
            <p className="font-mono text-xs bg-black/10 p-2 rounded whitespace-pre-wrap">
            {error}
            </p>
        </AlertDescription>
        </Alert>
    </div>
  )
}
