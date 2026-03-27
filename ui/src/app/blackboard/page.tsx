/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import React from "react";
import { BlackboardDashboard } from "@/components/blackboard/blackboard-dashboard";

/**
 * BlackboardPage component for displaying the Blackboard dashboard.
 *
 * Summary: Renders the root page view for the Blackboard.
 *
 * Parameters:
 *   - None.
 *
 * Returns:
 *   - JSX.Element: The rendered page layout containing the Blackboard dashboard.
 *
 * Errors/Throws:
 *   - None explicitly thrown by the component.
 *
 * Side Effects:
 *   - Renders child components which may have their own side effects.
 */
export default function BlackboardPage() {
    return (
        <div className="container py-6 space-y-6">
            <div>
                <h1 className="text-3xl font-bold tracking-tight">Blackboard Lineage Inspector</h1>
                <p className="text-muted-foreground">
                    Visualize and debug Agent-Bound Blackboard data across different Intent Scopes.
                </p>
            </div>
            <BlackboardDashboard />
        </div>
    );
}
