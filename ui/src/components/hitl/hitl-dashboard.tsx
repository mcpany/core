/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import React, { useState } from "react";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription, DialogFooter } from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";

/**
 * Intent: Document HitlDashboard
 *
 * Params:
 *   - None
 *
 * Errors:
 *   - None
 *
 * HitlDashboard component for managing Human-in-the-Loop approvals.
 *
 * Summary: Renders a dashboard for reviewing and managing pending HITL approvals.
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
 *   - Uses local React state to manage approval statuses.
 */
interface HITLApproval {
    id: string;
    tool: string;
    intent: string;
    status: string;
    requireMfa: boolean;
}

/**
 * Renders a dashboard for reviewing and managing pending HITL approvals.
 *
 * Summary: Displays a dashboard component for HITL (Human-in-the-Loop) approvals.
 *
 * Parameters:
 *   - None.
 *
 * Returns:
 *   - JSX.Element: The rendered dashboard component.
 *
 * Throws/Errors:
 *   - None explicitly thrown by the component itself.
 */
export function HitlDashboard() {
    const [approvals, setApprovals] = React.useState<HITLApproval[]>([]);
    const [mfaDialogOpen, setMfaDialogOpen] = useState(false);
    const [mfaCode, setMfaCode] = useState("");
    const [pendingApprovalId, setPendingApprovalId] = useState<string | null>(null);

    React.useEffect(() => {
        fetchApprovals();
        const interval = setInterval(fetchApprovals, 3000);
        return () => clearInterval(interval);
    }, []);

    const fetchApprovals = async () => {
        try {
            const res = await fetch("/api/v1/hitl/approvals");
            if (res.ok) {
                const data = await res.json();
                setApprovals(data || []);
            }
        } catch (err) {
            console.error("Failed to fetch HITL approvals", err);
        }
    };

    const handleAction = (id: string, action: "approved" | "denied") => {
        const approval = approvals.find(a => a.id === id);
        if (action === "approved" && approval?.requireMfa) {
            setPendingApprovalId(id);
            setMfaDialogOpen(true);
            return;
        }
        executeAction(id, action);
    };

    const executeAction = async (id: string, action: "approved" | "denied", code?: string) => {
        try {
            await fetch(`/api/v1/hitl/approvals/${id}`, {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({ action, mfaCode: code || "" })
            });
            fetchApprovals();
        } catch (err) {
            console.error("Action failed", err);
        }
    };

    const handleMfaSubmit = () => {
        if (pendingApprovalId && mfaCode.length > 0) {
            executeAction(pendingApprovalId, "approved", mfaCode);
            setMfaDialogOpen(false);
            setMfaCode("");
            setPendingApprovalId(null);
        }
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

            <Dialog open={mfaDialogOpen} onOpenChange={setMfaDialogOpen}>
                <DialogContent>
                    <DialogHeader>
                        <DialogTitle>Multi-Factor Authentication Required</DialogTitle>
                        <DialogDescription>
                            Please enter your MFA code to approve this sensitive action.
                        </DialogDescription>
                    </DialogHeader>
                    <Input
                        type="text"
                        placeholder="MFA Code"
                        value={mfaCode}
                        onChange={(e) => setMfaCode(e.target.value)}
                        onKeyDown={(e) => {
                            if (e.key === "Enter") {
                                handleMfaSubmit();
                            }
                        }}
                        autoFocus
                    />
                    <DialogFooter>
                        <Button variant="outline" onClick={() => setMfaDialogOpen(false)}>Cancel</Button>
                        <Button onClick={handleMfaSubmit} disabled={mfaCode.length === 0}>Verify & Approve</Button>
                    </DialogFooter>
                </DialogContent>
            </Dialog>
        </div>
    );
}
