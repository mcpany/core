/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */


import { Suspense } from "react"
import { Loader2 } from "lucide-react"
import { LogStream } from "@/components/logs/log-stream"

/**
 * LogsPage component.
 * @returns The rendered component.
 */
export default function LogsPage() {
  return (
    <div className="h-full p-4 md:p-8 pt-6">
      <Suspense fallback={<div className="flex items-center justify-center h-full"><Loader2 className="h-8 w-8 animate-spin" /></div>}>
        <LogStream />
      </Suspense>
    </div>
  )
}
