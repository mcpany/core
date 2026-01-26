/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useEffect, useState } from "react";
import { UpstreamServiceConfig, apiClient, ValidateServiceResponse } from "@/lib/client";
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle, DialogTrigger } from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Loader2, CheckCircle2, XCircle, Clock, AlertTriangle, Play, HelpCircle } from "lucide-react";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Badge } from "@/components/ui/badge";
import { cn } from "@/lib/utils";

interface DiagnosticModalProps {
    service: UpstreamServiceConfig;
    open: boolean;
    onOpenChange: (open: boolean) => void;
}

export function DiagnosticModal({ service, open, onOpenChange }: DiagnosticModalProps) {
    const [running, setRunning] = useState(false);
    const [result, setResult] = useState<ValidateServiceResponse | null>(null);

    const runDiagnostics = async () => {
        setRunning(true);
        setResult(null);
        try {
            const res = await apiClient.validateService(service);
            setResult(res);
        } catch (e) {
            setResult({
                valid: false,
                error: String(e),
                details: "An unexpected error occurred while communicating with the server.",
                steps: [
                    { name: "API Reachability", status: "error", message: String(e) }
                ]
            });
        } finally {
            setRunning(false);
        }
    };

    useEffect(() => {
        if (open) {
            runDiagnostics();
        }
    }, [open]);

    return (
        <Dialog open={open} onOpenChange={onOpenChange}>
            <DialogContent className="max-w-2xl">
                <DialogHeader>
                    <DialogTitle className="flex items-center gap-2">
                        Connection Diagnostics
                        <Badge variant="outline">{service.name}</Badge>
                    </DialogTitle>
                    <DialogDescription>
                        Testing connectivity and configuration validity for this upstream service.
                    </DialogDescription>
                </DialogHeader>

                <div className="py-4 space-y-6">
                    {/* Status Overview */}
                    <div className="flex items-center justify-between p-4 rounded-lg border bg-muted/30">
                        <div className="flex items-center gap-3">
                            {running ? (
                                <Loader2 className="h-8 w-8 animate-spin text-primary" />
                            ) : result?.valid ? (
                                <CheckCircle2 className="h-8 w-8 text-green-500" />
                            ) : (
                                <XCircle className="h-8 w-8 text-red-500" />
                            )}
                            <div>
                                <h3 className="font-semibold text-lg">
                                    {running ? "Running tests..." : result?.valid ? "Connection Successful" : "Connection Failed"}
                                </h3>
                                <p className="text-sm text-muted-foreground">
                                    {result?.latency_ms ? `Completed in ${result.latency_ms}ms` : running ? "Please wait..." : "Ready"}
                                </p>
                            </div>
                        </div>
                        {!running && (
                            <Button size="sm" variant="outline" onClick={runDiagnostics}>
                                <Play className="h-4 w-4 mr-2" />
                                Rerun
                            </Button>
                        )}
                    </div>

                    {/* Detailed Steps */}
                    {result?.steps && (
                        <div className="space-y-2">
                            <h4 className="text-sm font-medium text-muted-foreground uppercase tracking-wider">Validation Steps</h4>
                            <div className="border rounded-md divide-y">
                                {result.steps.map((step, i) => (
                                    <div key={i} className="flex items-start gap-3 p-3">
                                        <div className="mt-0.5">
                                            {step.status === "success" && <CheckCircle2 className="h-5 w-5 text-green-500" />}
                                            {step.status === "error" && <XCircle className="h-5 w-5 text-red-500" />}
                                            {step.status === "pending" && <Loader2 className="h-5 w-5 animate-spin text-muted-foreground" />}
                                            {step.status === "skipped" && <HelpCircle className="h-5 w-5 text-muted-foreground" />}
                                        </div>
                                        <div className="flex-1">
                                            <p className="font-medium text-sm">{step.name}</p>
                                            {step.message && (
                                                <p className={cn("text-sm mt-1", step.status === "error" ? "text-red-600 dark:text-red-400" : "text-muted-foreground")}>
                                                    {step.message}
                                                </p>
                                            )}
                                        </div>
                                    </div>
                                ))}
                            </div>
                        </div>
                    )}

                    {/* Error Details */}
                    {!running && !result?.valid && result?.details && (
                        <div className="p-4 rounded-lg bg-red-50 dark:bg-red-900/10 border border-red-200 dark:border-red-900/20 text-sm">
                            <div className="flex items-center gap-2 font-semibold text-red-800 dark:text-red-300 mb-2">
                                <AlertTriangle className="h-4 w-4" />
                                Error Details
                            </div>
                            <p className="text-red-700 dark:text-red-400 whitespace-pre-wrap font-mono text-xs">
                                {result.details}
                            </p>
                        </div>
                    )}
                </div>

                <DialogFooter>
                    <Button variant="outline" onClick={() => onOpenChange(false)}>Close</Button>
                </DialogFooter>
            </DialogContent>
        </Dialog>
    );
}
