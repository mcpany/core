/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import React, { useState } from "react";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";

/**
 * Summary: Visualizes granular capability tokens grouped by agent roles within a grid layout.
 *
 * Returns:
 *   - React.JSX.Element: The rendered ScopesDashboard interface.
 *
 * Throws/Errors:
 *   - None.
 */
export function ScopesDashboard() {
    const [roles, setRoles] = useState([
        { id: "1", role: "default", scopes: ["read"] },
        { id: "2", role: "admin", scopes: ["fs:read:/tmp", "db:write:users", "system:manage"] },
        { id: "3", role: "agent-a", scopes: ["fs:read:/app", "api:request"] }
    ]);

    return (
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
            {roles.map(r => (
                <Card key={r.id}>
                    <CardHeader>
                        <CardTitle>Role: {r.role}</CardTitle>
                        <CardDescription>Configured Capability Tokens</CardDescription>
                    </CardHeader>
                    <CardContent>
                        <div className="flex flex-wrap gap-2">
                            {r.scopes.map((scope, index) => (
                                <Badge key={index} variant="secondary">{scope}</Badge>
                            ))}
                        </div>
                    </CardContent>
                </Card>
            ))}
        </div>
    );
}
