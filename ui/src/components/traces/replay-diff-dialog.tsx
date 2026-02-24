/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import React, { useState, useEffect } from "react";
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle, DialogFooter } from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { DiffViewer } from "@/components/services/editor/diff-viewer";
import { Trace } from "@/types/trace";
import { apiClient } from "@/lib/client";
import { Loader2, RefreshCcw, AlertTriangle } from "lucide-react";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";

interface ReplayDiffDialogProps {
    open: boolean;
    onOpenChange: (open: boolean) => void;
    trace: Trace | null;
}

/**
 * ReplayDiffDialog component.
 * Allows replaying a tool call from a trace and viewing the diff between original and new output.
 */
export function ReplayDiffDialog({ open, onOpenChange, trace }: ReplayDiffDialogProps) {
    const [loading, setLoading] = useState(false);
    const [replayResult, setReplayResult] = useState<any>(null);
    const [error, setError] = useState<string | null>(null);

    useEffect(() => {
        if (open && trace) {
            handleReplay();
        } else {
            // Reset state on close
            setReplayResult(null);
            setError(null);
        }
    }, [open, trace]);

    const handleReplay = async () => {
        if (!trace || trace.rootSpan.type !== 'tool') return;

        setLoading(true);
        setError(null);
        setReplayResult(null);

        try {
            const toolName = trace.rootSpan.name;
            const args = trace.rootSpan.input || {};

            const result = await apiClient.executeTool({
                name: toolName,
                arguments: args
            });

            setReplayResult(result);
        } catch (e: any) {
            setError(e.message || "Failed to replay tool execution.");
        } finally {
            setLoading(false);
        }
    };

    if (!trace) return null;

    const originalOutput = JSON.stringify(trace.rootSpan.output || {}, null, 2);
    const newOutput = replayResult
        ? JSON.stringify(replayResult, null, 2)
        : error
            ? `// Replay Failed\n${error}`
            : "// Waiting for execution...";

    return (
        <Dialog open={open} onOpenChange={onOpenChange}>
            <DialogContent className="max-w-4xl h-[80vh] flex flex-col">
                <DialogHeader>
                    <DialogTitle className="flex items-center gap-2">
                        <RefreshCcw className="h-5 w-5 text-primary" />
                        Replay & Diff Analysis
                    </DialogTitle>
                    <DialogDescription>
                        Re-executing <strong>{trace.rootSpan.name}</strong> with original inputs.
                    </DialogDescription>
                </DialogHeader>

                <div className="flex-1 min-h-0 py-4 flex flex-col gap-4">
                    {loading && (
                        <div className="flex items-center gap-2 text-sm text-muted-foreground animate-pulse">
                            <Loader2 className="h-4 w-4 animate-spin" />
                            Running tool...
                        </div>
                    )}

                    {error && (
                        <Alert variant="destructive">
                            <AlertTriangle className="h-4 w-4" />
                            <AlertTitle>Execution Failed</AlertTitle>
                            <AlertDescription>
                                The replay execution encountered an error. The diff below compares the original output with the error message.
                            </AlertDescription>
                        </Alert>
                    )}

                    <div className="flex-1 border rounded-md overflow-hidden bg-muted/10">
                        <div className="flex items-center justify-between px-4 py-2 border-b bg-muted/20 text-xs font-medium text-muted-foreground">
                            <span>Original Output ({new Date(trace.timestamp).toLocaleTimeString()})</span>
                            <span>Replay Output (Now)</span>
                        </div>
                        <DiffViewer
                            original={originalOutput}
                            modified={newOutput}
                            language="json"
                        />
                    </div>
                </div>

                <DialogFooter>
                    <Button variant="outline" onClick={() => onOpenChange(false)}>Close</Button>
                    <Button onClick={handleReplay} disabled={loading}>
                        <RefreshCcw className="mr-2 h-4 w-4" />
                        Rerun
                    </Button>
                </DialogFooter>
            </DialogContent>
        </Dialog>
    );
}
