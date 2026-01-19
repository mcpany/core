/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { SystemDiagnostic } from "@/components/diagnostics/system-diagnostic";

/**
 * Diagnostics page component.
 * Displays system health and diagnostics.
 * @returns The rendered component.
 */
export default function DiagnosticsPage() {
  return (
    <div className="flex-1 space-y-4 p-8 pt-6">
      <div className="flex items-center justify-between space-y-2">
        <h1 className="text-3xl font-bold tracking-tight">System Diagnostics</h1>
      </div>
      <div className="hidden md:block">
        <p className="text-muted-foreground mb-4">
          Run comprehensive health checks on the MCP Any server, upstream services, and configuration.
        </p>
      </div>

      <SystemDiagnostic />
    </div>
  );
}
