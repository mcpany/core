/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import React from "react";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle
} from "@/components/ui/card";
import {
  GitMerge,
  Search,
  Activity
} from "lucide-react";
import { AgentChainTracer } from "@/components/dashboard/agent-chain-tracer";
import { LazyMcpDashboard } from "@/components/dashboard/lazy-mcp-dashboard";
import { MultiAgentSessionTimeline } from "@/components/dashboard/multi-agent-session-timeline";
import { UnifiedDiscoveryManager } from "@/components/dashboard/unified-discovery-manager";

/**
 * Intent: Document UniversalAgentBusPage
 *
 * Params:
 *   - None
 *
 * Errors:
 *   - None
 *
 * UniversalAgentBusPage component for managing the Universal Agent Bus.
 *
 * Summary: Renders a dashboard with metrics for the Universal Agent Bus.
 *
 * Parameters:
 *   - None.
 *
 * Returns:
 *   - JSX.Element: The rendered layout for the agent bus dashboard.
 *
 * Errors/Throws:
 *   - None explicitly thrown by the component.
 *
 * Side Effects:
 *   - Renders multiple dashboard cards.
 */
export default function UniversalAgentBusPage() {
  return (
    <div className="p-6 space-y-6">
      <div>
        <h1 className="text-3xl font-bold tracking-tight">Universal Agent Bus</h1>
        <p className="text-muted-foreground mt-2">
          Manage and map subagents dynamically. Visualize, configure, and orchestrate interactions between complex AI swarms.
        </p>
      </div>

      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
        {/* Recursive Context Dashboard */}
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Recursive Context Dashboard</CardTitle>
            <GitMerge className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">Inactive</div>
            <p className="text-xs text-muted-foreground">
              Visualize state inheritance and session tokens across agent swarms.
            </p>
          </CardContent>
        </Card>

        <MultiAgentSessionTimeline />
        <UnifiedDiscoveryManager />

      </div>

      <div className="mt-8">
        <LazyMcpDashboard />
      </div>

      <div className="mt-8">
        <AgentChainTracer />
      </div>
    </div>
  );
}
