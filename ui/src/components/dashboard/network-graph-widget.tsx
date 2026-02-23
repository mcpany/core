/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import React from "react";
import { ReactFlowProvider } from "@xyflow/react";
import { NetworkGraphFlow } from "@/components/network/network-graph-client";

/**
 * NetworkGraphWidget component for the dashboard.
 * Renders a simplified network topology graph.
 * @returns The rendered component.
 */
export const NetworkGraphWidget = React.memo(function NetworkGraphWidget() {
    return (
        <div className="h-[350px] w-full border rounded-md overflow-hidden bg-muted/5 relative">
            <ReactFlowProvider>
                <NetworkGraphFlow widgetMode={true} />
            </ReactFlowProvider>
        </div>
    );
});
// ⚡ BOLT: Memoize widget to prevent re-renders during dashboard drag/resize operations.
// Randomized Selection from Top 5 High-Impact Targets
