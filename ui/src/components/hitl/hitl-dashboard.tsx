/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import React, { useState, useEffect } from "react";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription, DialogFooter } from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { apiClient, HitlApproval } from "@/lib/client";
import { ShieldAlert, CheckCircle, XCircle, ShieldCheck, Clock } from "lucide-react";
import { useToast } from "@/hooks/use-toast";

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
    const [approvals, setApprovals] = useState<HitlApproval[]>([]);
    const [mfaDialogOpen, setMfaDialogOpen] = useState(false);
    const [mfaCode, setMfaCode] = useState("");
    const [pendingApprovalId, setPendingApprovalId] = useState<string | null>(null);
    const [isProcessing, setIsProcessing] = useState<Record<string, boolean>>({});
    const { toast } = useToast();

    useEffect(() => {
        fetchApprovals();
        const interval = setInterval(fetchApprovals, 3000);
        return () => clearInterval(interval);
    }, []);

    const fetchApprovals = async () => {
        try {
            const data = await apiClient.getHitlApprovals();
            setApprovals(data || []);
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
        setIsProcessing(prev => ({ ...prev, [id]: true }));
        try {
            await apiClient.actionHitlApproval(id, action, code);
            toast({
                title: action === "approved" ? "Action Approved" : "Action Denied",
                description: `Successfully ${action} the suspended action.`,
                variant: action === "denied" ? "destructive" : "default"
            });
            await fetchApprovals();
        } catch (err) {
            console.error("Action failed", err);
            toast({
                title: "Action Failed",
                description: "Failed to process the HITL approval.",
                variant: "destructive"
            });
        } finally {
            setIsProcessing(prev => ({ ...prev, [id]: false }));
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

    if (approvals.length === 0) {
        return (
            <div className="flex flex-col items-center justify-center p-12 mt-8 border-2 border-dashed rounded-xl bg-muted/10">
                <ShieldCheck className="w-16 h-16 mb-4 text-muted-foreground opacity-20" />
                <h3 className="text-xl font-medium tracking-tight">No pending approvals</h3>
                <p className="text-muted-foreground mt-2 text-center max-w-sm">
                    There are currently no intercepted actions awaiting your review.
                </p>
            </div>
        );
    }

    return (
        <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
            {approvals.map(a => (
                <Card key={a.id} className="backdrop-blur-md bg-background/60 border-white/10 shadow-lg animate-in fade-in slide-in-from-bottom-4 transition-all duration-300">
                    <CardHeader className="pb-3 border-b border-border/50 bg-muted/20">
                        <div className="flex justify-between items-start">
                            <div className="flex gap-2 items-center">
                                <ShieldAlert className={`w-5 h-5 ${a.requireMfa ? 'text-amber-500' : 'text-blue-500'}`} />
                                <CardTitle className="text-lg">{a.tool}</CardTitle>
                            </div>
                            <span className="flex items-center gap-1 text-xs font-medium px-2 py-1 bg-background/50 rounded-full border border-border/50">
                                <Clock className="w-3 h-3" /> Pending
                            </span>
                        </div>
                    </CardHeader>
                    <CardContent className="pt-4 space-y-4">
                        <div className="space-y-1">
                            <p className="text-sm font-medium text-muted-foreground">Intent</p>
                            <p className="text-sm leading-relaxed">{a.intent}</p>
                        </div>

                        {a.requireMfa && (
                            <div className="text-xs font-medium text-amber-500 bg-amber-500/10 px-3 py-2 rounded-md flex items-center gap-2">
                                <ShieldAlert className="w-4 h-4" />
                                MFA validation required for approval
                            </div>
                        )}

                        {a.status === "pending" ? (
                            <div className="flex gap-3 pt-2">
                                <Button
                                    onClick={() => handleAction(a.id, "approved")}
                                    variant="default"
                                    className="flex-1 gap-2 shadow-sm"
                                    disabled={isProcessing[a.id]}
                                >
                                    <CheckCircle className="w-4 h-4" /> Approve
                                </Button>
                                <Button
                                    onClick={() => handleAction(a.id, "denied")}
                                    variant="destructive"
                                    className="flex-1 gap-2 shadow-sm"
                                    disabled={isProcessing[a.id]}
                                >
                                    <XCircle className="w-4 h-4" /> Deny
                                </Button>
                            </div>
                        ) : (
                            <div className="text-sm font-medium uppercase text-muted-foreground pt-2">
                                Status: {a.status}
                            </div>
                        )}
                    </CardContent>
                </Card>
            ))}

            <Dialog open={mfaDialogOpen} onOpenChange={setMfaDialogOpen}>
                <DialogContent className="sm:max-w-md backdrop-blur-xl bg-background/95">
                    <DialogHeader>
                        <DialogTitle className="flex items-center gap-2">
                            <ShieldAlert className="w-5 h-5 text-amber-500" />
                            Multi-Factor Authentication
                        </DialogTitle>
                        <DialogDescription>
                            This is a sensitive operation. Please enter your MFA code to authorize.
                        </DialogDescription>
                    </DialogHeader>
                    <div className="py-4">
                        <Input
                            type="text"
                            placeholder="Enter 6-digit code"
                            value={mfaCode}
                            onChange={(e) => setMfaCode(e.target.value)}
                            className="text-center tracking-widest text-lg font-mono"
                            onKeyDown={(e) => {
                                if (e.key === "Enter") {
                                    handleMfaSubmit();
                                }
                            }}
                            autoFocus
                        />
                    </div>
                    <DialogFooter className="sm:justify-between">
                        <Button variant="ghost" onClick={() => setMfaDialogOpen(false)}>Cancel</Button>
                        <Button
                            onClick={handleMfaSubmit}
                            disabled={mfaCode.length === 0}
                            className="gap-2"
                        >
                            <ShieldCheck className="w-4 h-4" />
                            Verify & Approve
                        </Button>
                    </DialogFooter>
                </DialogContent>
            </Dialog>
        </div>
    );
}
