/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import React from "react";
import { ScopesDashboard } from "@/components/scopes/scopes-dashboard";

/**
 * Summary: Renders the ScopesPage component, serving as the dashboard interface for configuring capability-based tokens.
 *
 * Params:
 *   - None.
 *
 * Returns:
 *   - React.JSX.Element: The rendered ScopesPage component.
 *
 * Throws/Errors:
 *   - None.
 */
export default function ScopesPage() {
    return (
        <div className="container py-6 space-y-6">
            <div>
                <h1 className="text-3xl font-bold tracking-tight">Granular Scopes</h1>
                <p className="text-muted-foreground">
                    Capability-based token system enabling Least Privilege security for agents.
                </p>
            </div>
            <ScopesDashboard />
        </div>
    );
}
