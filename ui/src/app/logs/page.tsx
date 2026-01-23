/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */


import { LogStream } from "@/components/logs/log-stream"
import { Suspense } from "react"
import { Loader2 } from "lucide-react"

/**
 * LogsPage component.
 * @returns The rendered component.
 */
export default function LogsPage() {
  return (
    <div className="h-full p-4 md:p-8 pt-6">
      <Suspense fallback={<div className="flex h-full items-center justify-center"><Loader2 className="h-6 w-6 animate-spin" /></div>}>
        <LogStream />
      </Suspense>
    </div>
  )
}
