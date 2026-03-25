/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { AgentFlow } from "@/components/visualizer/agent-flow";

/**
 * Summary: VisualizerPage component.
 *
 * Parameters:
 *   - None.
 *
 * Returns:
 *   - React.ReactNode: The rendered component.
 *
 * Throws/Errors:
 *   - None.
 */
export default function VisualizerPage() {
  return (
    <div className="flex flex-col h-full p-6 space-y-6">
      <div className="flex flex-col gap-2">
        <h1 className="text-3xl font-bold tracking-tight">Agent Flow Visualizer</h1>
        <p className="text-muted-foreground">
          Live visualization of concurrent agent flows, tool executions, and resource dependencies.
        </p>
      </div>
      <div className="flex-1 min-h-[500px]">
        <AgentFlow />
      </div>
    </div>
  );
}
