/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import React, { useState } from "react";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Button } from "@/components/ui/button";

/**
 * Intent: Document BlackboardDashboard
 *
 * Params:
 *   - None
 *
 * Errors:
 *   - None
 *
 * BlackboardDashboard component for managing shared blackboard keys.
 *
 * Summary: Displays a list of shared agent blackboard variables.
 *
 * Parameters:
 *   - None.
 *
 * Returns:
 *   - JSX.Element: The rendered dashboard component.
 *
 * Errors/Throws:
 *   - None explicitly thrown by the component itself.
 *
 * Side Effects:
 *   - Uses local React state to manage keys.
 */
interface BlackboardKey {
    agentId: string;
    key: string;
    value: string;
    intent?: string; // Kept for UI backwards compatibility if provided
}

export function BlackboardDashboard() {
    const [keys, setKeys] = useState<BlackboardKey[]>([]);

    React.useEffect(() => {
        fetchKeys();
        const interval = setInterval(fetchKeys, 3000);
        return () => clearInterval(interval);
    }, []);

    const fetchKeys = async () => {
        try {
            const res = await fetch("/api/v1/blackboard/keys");
            if (res.ok) {
                const data = await res.json();
                setKeys(data || []);
            }
        } catch (err) {
            console.error("Failed to fetch blackboard keys", err);
        }
    };

    return (
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
            {keys.map(k => (
                <Card key={`${k.agentId}-${k.key}`}>
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
