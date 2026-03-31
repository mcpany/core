/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import React from "react";
import { ScopesDashboard } from "@/components/scopes/scopes-dashboard";

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
