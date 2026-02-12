/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { WorkflowStep } from "@/types/workflow";
import { Loader2, CheckCircle2, XCircle, Play, Terminal, Trash2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";

interface WorkflowItemProps {
    step: WorkflowStep;
    onRun: (stepId: string) => void;
    onDelete: (stepId: string) => void;
    index: number;
}

export function WorkflowItem({ step, onRun, onDelete, index }: WorkflowItemProps) {
    return (
        <div className="group flex items-center gap-3 p-3 bg-muted/40 rounded-md border border-transparent hover:border-border transition-all">
            <div className="flex items-center justify-center w-6 h-6 rounded-full bg-background border text-[10px] text-muted-foreground font-mono shrink-0">
                {index + 1}
            </div>

            <div className="flex-1 min-w-0">
                <div className="flex items-center gap-2">
                    <Terminal className="h-3 w-3 text-primary opacity-70" />
                    <span className="font-medium text-sm truncate" title={step.toolName}>
                        {step.name || step.toolName}
                    </span>
                </div>
                <div className="text-[10px] text-muted-foreground font-mono truncate opacity-70 mt-0.5">
                    {JSON.stringify(step.arguments)}
                </div>

                {step.error && (
                    <div className="text-[10px] text-destructive mt-1 break-words">
                        {step.error}
                    </div>
                )}
            </div>

            <div className="flex items-center gap-1">
                {step.status === 'running' && <Loader2 className="h-4 w-4 animate-spin text-blue-500" />}
                {step.status === 'success' && <CheckCircle2 className="h-4 w-4 text-green-500" />}
                {step.status === 'error' && <XCircle className="h-4 w-4 text-destructive" />}

                {!step.status && (
                    <Button
                        variant="ghost"
                        size="icon"
                        className="h-6 w-6 opacity-0 group-hover:opacity-100 transition-opacity"
                        onClick={() => onRun(step.id)}
                        title="Run single step"
                    >
                        <Play className="h-3 w-3" />
                    </Button>
                )}

                <Button
                    variant="ghost"
                    size="icon"
                    className="h-6 w-6 opacity-0 group-hover:opacity-100 transition-opacity text-muted-foreground hover:text-destructive"
                    onClick={() => onDelete(step.id)}
                    title="Remove step"
                >
                    <Trash2 className="h-3 w-3" />
                </Button>
            </div>
        </div>
    );
}
