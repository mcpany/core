/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import React, { useState } from "react";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Button } from "@/components/ui/button";

export function BlackboardDashboard() {
    const [keys, setKeys] = useState([
        { id: "1", agentId: "agent-a", key: "session_token", value: "abc-123", intent: "auth" },
        { id: "2", agentId: "agent-b", key: "last_query", value: "select * from users", intent: "database_read" }
    ]);

    return (
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
            {keys.map(k => (
                <Card key={k.id}>
                    <CardHeader>
                        <CardTitle>{k.key}</CardTitle>
                        <CardDescription>Agent ID: {k.agentId}</CardDescription>
                    </CardHeader>
                    <CardContent>
                        <div className="space-y-2">
                            <div className="text-sm font-medium">Value: <span className="font-normal text-muted-foreground">{k.value}</span></div>
                            <div className="text-sm font-medium">Intent: <span className="font-normal text-muted-foreground">{k.intent}</span></div>
                        </div>
                    </CardContent>
                </Card>
            ))}
        </div>
    );
}
