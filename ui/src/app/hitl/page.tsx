/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import React from "react";
import { HitlDashboard } from "@/components/hitl/hitl-dashboard";

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
