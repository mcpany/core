/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import React, { useState, useEffect } from "react";
import { Card, CardContent, CardHeader, CardTitle, CardDescription, CardFooter } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription, DialogFooter } from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { apiClient, HITLApproval } from "@/lib/client";
import { useToast } from "@/hooks/use-toast";
import { Badge } from "@/components/ui/badge";
import { Loader2, ShieldAlert, CheckCircle2, XCircle } from "lucide-react";

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
 */

export function HitlDashboard() {
    const [approvals, setApprovals] = useState<HITLApproval[]>([]);
    const [mfaDialogOpen, setMfaDialogOpen] = useState(false);
    const [mfaCode, setMfaCode] = useState("");
    const [pendingApprovalId, setPendingApprovalId] = useState<string | null>(null);
    const [loading, setLoading] = useState(true);
    const { toast } = useToast();

    useEffect(() => {
        fetchApprovals();
        const interval = setInterval(fetchApprovals, 3000);
        return () => clearInterval(interval);
    }, []);

    const fetchApprovals = async () => {
        try {
            const data = await apiClient.getHITLApprovals();
            setApprovals(data || []);
        } catch (err) {
            console.error("Failed to fetch HITL approvals", err);
            toast({
                variant: "destructive",
                title: "Error",
                description: "Failed to fetch pending approvals.",
            });
        } finally {
            setLoading(false);
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
            await apiClient.resolveHITLApproval(id, action, code);
            toast({
                title: action === "approved" ? "Action Approved" : "Action Denied",
                description: `Successfully ${action} the requested action.`,
            });
            fetchApprovals();
        } catch (err) {
            console.error("Action failed", err);
            toast({
                variant: "destructive",
                title: "Error",
                description: `Failed to execute action: ${err instanceof Error ? err.message : "Unknown error"}`,
            });
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

    if (loading) {
        return (
            <div className="flex h-[400px] items-center justify-center">
                <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
            </div>
        );
    }

    if (approvals.length === 0) {
        return (
            <div className="flex flex-col items-center justify-center h-[400px] border border-dashed rounded-lg bg-muted/20">
                <ShieldAlert className="h-12 w-12 text-muted-foreground mb-4" />
                <h3 className="text-xl font-medium tracking-tight">No pending approvals</h3>
                <p className="text-sm text-muted-foreground mt-2">
                    All clear. There are no intercepted actions requiring human intervention.
                </p>
            </div>
        );
    }

    return (
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
            {approvals.map(a => (
                <Card key={a.id} data-testid="hitl-card" className="backdrop-blur-sm bg-background/50 flex flex-col justify-between">
                    <CardHeader className="pb-3">
                        <div className="flex justify-between items-start mb-2">
                            <Badge variant={a.status === "pending" ? "default" : "secondary"}>
                                {a.status}
                            </Badge>
                            {a.requireMfa && (
                                <Badge variant="outline" className="border-amber-500/50 text-amber-500">
                                    MFA Required
                                </Badge>
                            )}
                        </div>
                        <CardTitle className="text-lg font-semibold">{a.tool}</CardTitle>
                        <CardDescription className="line-clamp-2" title={a.intent}>
                            {a.intent}
                        </CardDescription>
                    </CardHeader>
                    <CardContent className="pb-4">
                        <div className="text-sm font-medium text-muted-foreground mb-1">Execution ID</div>
                        <code className="text-xs bg-muted px-2 py-1 rounded block truncate" title={a.id}>
                            {a.id}
                        </code>
                    </CardContent>
                    <CardFooter className="pt-0">
                        {a.status === "pending" ? (
                            <div className="flex gap-2 w-full">
                                <Button className="flex-1" onClick={() => handleAction(a.id, "approved")} variant="default">
                                    <CheckCircle2 className="w-4 h-4 mr-2" />
                                    Approve
                                </Button>
                                <Button className="flex-1" onClick={() => handleAction(a.id, "denied")} variant="destructive">
                                    <XCircle className="w-4 h-4 mr-2" />
                                    Deny
                                </Button>
                            </div>
                        ) : (
                            <div className="text-sm font-medium uppercase text-muted-foreground w-full text-center py-2 bg-muted/50 rounded">
                                Status: {a.status}
                            </div>
                        )}
                    </CardFooter>
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
