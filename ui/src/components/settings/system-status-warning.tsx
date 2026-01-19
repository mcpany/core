/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client"

import React, { useEffect, useState } from "react"
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert"
import { AlertCircle, X } from "lucide-react"
import { Button } from "@/components/ui/button"

export function SystemStatusWarning() {
  const [warnings, setWarnings] = useState<string[]>([])
  const [visible, setVisible] = useState(true)

  useEffect(() => {
    // Check local storage for dismissal
    const dismissed = localStorage.getItem("system-warning-dismissed")
    if (dismissed === "true") {
      setVisible(false)
      return
    }

    const fetchStatus = async () => {
      try {
        const response = await fetch("/api/v1/system/status")
        if (response.ok) {
          const data = await response.json()
          if (data.security_warnings && data.security_warnings.length > 0) {
            setWarnings(data.security_warnings)
          }
        }
      } catch (error) {
        console.error("Failed to fetch system status for warnings", error)
      }
    }

    fetchStatus()
  }, [])

  if (!visible || warnings.length === 0) {
    return null
  }

  const handleDismiss = () => {
    setVisible(false)
    localStorage.setItem("system-warning-dismissed", "true")
  }

  return (
    <div className="mb-4">
      <Alert variant="destructive" className="relative">
        <AlertCircle className="h-4 w-4" />
        <AlertTitle>Configuration Warnings</AlertTitle>
        <AlertDescription>
          <ul className="list-disc list-inside mt-2">
            {warnings.map((w, i) => (
              <li key={i}>{w}</li>
            ))}
          </ul>
        </AlertDescription>
        <Button
          variant="ghost"
          size="icon"
          className="absolute top-2 right-2 h-6 w-6 text-destructive hover:text-destructive/80"
          onClick={handleDismiss}
        >
          <X className="h-4 w-4" />
        </Button>
      </Alert>
    </div>
  )
}
