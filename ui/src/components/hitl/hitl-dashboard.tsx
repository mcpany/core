/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import React, { useState } from "react";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";

export function HitlDashboard() {
    const [approvals, setApprovals] = useState([
        { id: "1", tool: "database.drop_table", intent: "Drop users table", status: "pending", requireMFA: true },
        { id: "2", tool: "system.restart", intent: "Restart backend service", status: "pending", requireMFA: false }
    ]);

    const [mfaDialogOpen, setMfaDialogOpen] = useState(false);
    const [mfaToken, setMfaToken] = useState("");
    const [activeApprovalId, setActiveApprovalId] = useState<string | null>(null);

    const handleAction = (id: string, action: "approved" | "denied") => {
        const approval = approvals.find(a => a.id === id);
        if (approval && approval.requireMFA && action === "approved") {
            setActiveApprovalId(id);
            setMfaDialogOpen(true);
            return;
        }
        updateStatus(id, action);
    };

    const updateStatus = (id: string, action: "approved" | "denied") => {
        setApprovals(prev => prev.map(a => a.id === id ? { ...a, status: action } : a));
    };

    const handleMfaSubmit = () => {
        if (mfaToken.length >= 6 && activeApprovalId) {
            updateStatus(activeApprovalId, "approved");
            setMfaDialogOpen(false);
            setMfaToken("");
            setActiveApprovalId(null);
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
                        <DialogTitle>MFA Required</DialogTitle>
                        <DialogDescription>
                            This is a high-risk action. Please enter your Multi-Factor Authentication token to approve.
                        </DialogDescription>
                    </DialogHeader>
                    <div className="grid gap-4 py-4">
                        <div className="grid gap-2">
                            <Label htmlFor="mfa-token">MFA Token</Label>
                            <Input
                                id="mfa-token"
                                value={mfaToken}
                                onChange={(e) => setMfaToken(e.target.value)}
                                placeholder="123456"
                                maxLength={6}
                            />
                        </div>
                    </div>
                    <DialogFooter>
                        <Button variant="outline" onClick={() => setMfaDialogOpen(false)}>Cancel</Button>
                        <Button onClick={handleMfaSubmit} disabled={mfaToken.length < 6}>Verify & Approve</Button>
                    </DialogFooter>
                </DialogContent>
            </Dialog>
        </div>
    );
}
