/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import React from "react";
import { HitlDashboard } from "@/components/hitl/hitl-dashboard";

/**
 * HitlPage component for displaying the HITL approvals dashboard.
 *
 * Summary: Renders the root page view for the HITL dashboard.
 *
 * Parameters:
 *   - None.
 *
 * Returns:
 *   - JSX.Element: The rendered page layout containing the HITL dashboard.
 *
 * Errors/Throws:
 *   - None explicitly thrown by the component.
 *
 * Side Effects:
 *   - Renders child components which may have their own side effects.
 */
export default function HitlPage() {
    return (
        <div className="container py-6 space-y-6">
            <div>
                <h1 className="text-3xl font-bold tracking-tight">HITL Approvals</h1>
                <p className="text-muted-foreground">
                    Review and approve or deny suspended actions intercepted by the HITL middleware.
                </p>
            </div>
            <HitlDashboard />
        </div>
    );
}
