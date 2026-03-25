/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import React, { useState } from "react";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Button } from "@/components/ui/button";

export function HitlDashboard() {
    const [approvals, setApprovals] = useState([
        { id: "1", tool: "database.drop_table", intent: "Drop users table", status: "pending" }
    ]);

    const handleAction = (id: string, action: "approved" | "denied") => {
        setApprovals(prev => prev.map(a => a.id === id ? { ...a, status: action } : a));
    };

    return (
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
            {approvals.map(a => (
                <Card key={a.id}>
                    <CardHeader>
                        <CardTitle>{a.tool}</CardTitle>
                        <CardDescription>Intent: {a.intent}</CardDescription>
                    </CardHeader>
                    <CardContent>
                        {a.status === "pending" ? (
                            <div className="flex gap-2">
                                <Button onClick={() => handleAction(a.id, "approved")} variant="default">Approve</Button>
                                <Button onClick={() => handleAction(a.id, "denied")} variant="destructive">Deny</Button>
                            </div>
                        ) : (
                            <div className="text-sm font-medium uppercase text-muted-foreground">
                                Status: {a.status}
                            </div>
                        )}
                    </CardContent>
                </Card>
            ))}
        </div>
    );
}
