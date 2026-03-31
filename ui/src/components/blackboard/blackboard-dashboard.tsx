/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import React, { useEffect, useState } from "react";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";

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
export function BlackboardDashboard() {
    interface BlackboardKey {
        id: string;
        agentId: string;
        key: string;
        value: string;
        intent: string;
    }
    const [keys, setKeys] = useState<BlackboardKey[]>([]);

    useEffect(() => {
        // Engineer solution: fetch keys dynamically from the new API
        fetch('/api/v1/blackboard/keys')
            .then(res => res.json())
            .then(data => setKeys(data || []))
            .catch(console.error);
    }, []);

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
