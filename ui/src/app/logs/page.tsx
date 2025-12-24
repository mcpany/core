/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */


import { LogStream } from "@/components/logs/log-stream"

export default function LogsPage() {
  return (
    <div className="h-full p-4 md:p-8 pt-6">
      <LogStream />
    </div>
  )
}
