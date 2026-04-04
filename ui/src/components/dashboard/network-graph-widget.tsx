/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */



import React from "react";
import { ReactFlowProvider } from "@xyflow/react";
import { NetworkGraphFlow } from "@/components/network/network-graph-client";

/**
 * NetworkGraphWidget executes the NetworkGraphWidget logic.
 *
 * Summary: Executes the NetworkGraphWidget logic.
 *
 * @param function NetworkGraphWidget( - The function NetworkGraphWidget( parameter.
 * @returns The result of the operation.
 * @throws An error if the operation fails.
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
