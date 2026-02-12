/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState } from "react";
import { Workflow, WorkflowStep } from "@/types/workflow";
import { WorkflowItem } from "./workflow-item";
import { runWorkflow } from "./workflow-runner";
import { Plus, Play, Trash2, ChevronRight, ChevronDown, ListChecks } from "lucide-react";
import { Button } from "@/components/ui/button";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Collapsible, CollapsibleContent, CollapsibleTrigger } from "@/components/ui/collapsible";
import { toast } from "@/hooks/use-toast";

interface WorkflowListProps {
    workflows: Workflow[];
    setWorkflows: (workflows: Workflow[]) => void;
}

/**
 * WorkflowList component.
 * @param props - The component props.
 * @param props.workflows - The list of workflows.
 * @param props.setWorkflows - Function to update the workflows.
 * @returns The rendered component.
 */
export function WorkflowList({ workflows, setWorkflows }: WorkflowListProps) {
    const [runningId, setRunningId] = useState<string | null>(null);
    const [openIds, setOpenIds] = useState<Set<string>>(new Set());

    const toggleOpen = (id: string) => {
        const newSet = new Set(openIds);
        if (newSet.has(id)) newSet.delete(id);
        else newSet.add(id);
        setOpenIds(newSet);
    };

    const handleDeleteWorkflow = (e: React.MouseEvent, id: string) => {
        e.stopPropagation();
        if (confirm("Are you sure you want to delete this workflow?")) {
            setWorkflows(workflows.filter(w => w.id !== id));
        }
    };

    const handleDeleteStep = (workflowId: string, stepId: string) => {
        setWorkflows(workflows.map(w => {
            if (w.id !== workflowId) return w;
            return { ...w, steps: w.steps.filter(s => s.id !== stepId) };
        }));
    };

    const handleRunWorkflow = async (e: React.MouseEvent, workflow: Workflow) => {
        e.stopPropagation();
        if (runningId) return;
        setRunningId(workflow.id);

        // Reset status
        const resetSteps = workflow.steps.map(s => ({ ...s, status: undefined, result: undefined, error: undefined }));
        const updatedWorkflow = { ...workflow, steps: resetSteps as WorkflowStep[] }; // Cast to fix type mismatch if any

        // Optimistic update to clear status
        setWorkflows(workflows.map(w => w.id === workflow.id ? updatedWorkflow : w));

        await runWorkflow(updatedWorkflow, (stepId, updates) => {
            setWorkflows(currentWorkflows => currentWorkflows.map(w => {
                if (w.id !== workflow.id) return w;
                return {
                    ...w,
                    steps: w.steps.map(s => s.id === stepId ? { ...s, ...updates } : s)
                };
            }));
        });

        setRunningId(null);
        toast({ title: "Workflow Completed", description: `Finished running ${workflow.name}` });
    };

    const handleRunStep = async (workflowId: string, stepId: string) => {
        // Find step
        const workflow = workflows.find(w => w.id === workflowId);
        if (!workflow) return;
        const step = workflow.steps.find(s => s.id === stepId);
        if (!step) return;

        // Create a temporary mini-workflow to run just this step using the runner
        const tempWorkflow = { ...workflow, steps: [step] };

        await runWorkflow(tempWorkflow, (sid, updates) => {
             setWorkflows(currentWorkflows => currentWorkflows.map(w => {
                if (w.id !== workflowId) return w;
                return {
                    ...w,
                    steps: w.steps.map(s => s.id === stepId ? { ...s, ...updates } : s)
                };
            }));
        });
    };

    return (
        <div className="flex flex-col h-full">
            <div className="p-4 border-b">
                <div className="flex items-center justify-between mb-2">
                    <h3 className="font-semibold text-sm flex items-center gap-2">
                        <ListChecks className="h-4 w-4" /> Workflows
                    </h3>
                    <Button variant="ghost" size="icon" className="h-6 w-6" onClick={() => {
                        const newWorkflow: Workflow = {
                            id: crypto.randomUUID(),
                            name: "New Workflow",
                            steps: [],
                            createdAt: new Date().toISOString(),
                            updatedAt: new Date().toISOString()
                        };
                        setWorkflows([...workflows, newWorkflow]);
                        toggleOpen(newWorkflow.id);
                    }}>
                        <Plus className="h-4 w-4" />
                    </Button>
                </div>
                <p className="text-xs text-muted-foreground">
                    Save tool executions as test cases.
                </p>
            </div>

            <ScrollArea className="flex-1">
                <div className="p-2 flex flex-col gap-2">
                    {workflows.length === 0 && (
                        <div className="text-center p-8 text-xs text-muted-foreground">
                            No workflows created yet.
                            <br/>
                            Execute a tool and click "Save" to start.
                        </div>
                    )}
                    {workflows.map(workflow => (
                        <Collapsible
                            key={workflow.id}
                            open={openIds.has(workflow.id)}
                            onOpenChange={() => toggleOpen(workflow.id)}
                            className="border rounded-md bg-card"
                        >
                            <div className="flex items-center p-2 hover:bg-accent/50 rounded-t-md transition-colors group">
                                <CollapsibleTrigger asChild>
                                    <Button variant="ghost" size="sm" className="p-0 h-6 w-6 mr-2">
                                        {openIds.has(workflow.id) ? <ChevronDown className="h-4 w-4" /> : <ChevronRight className="h-4 w-4" />}
                                    </Button>
                                </CollapsibleTrigger>
                                <span className="text-sm font-medium flex-1 truncate">{workflow.name}</span>

                                <div className="flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
                                    <Button
                                        variant="ghost"
                                        size="icon"
                                        className="h-6 w-6"
                                        onClick={(e) => handleRunWorkflow(e, workflow)}
                                        disabled={!!runningId}
                                        title="Run Workflow"
                                    >
                                        <Play className="h-3 w-3 text-green-500" />
                                    </Button>
                                    <Button
                                        variant="ghost"
                                        size="icon"
                                        className="h-6 w-6 text-destructive"
                                        onClick={(e) => handleDeleteWorkflow(e, workflow.id)}
                                        title="Delete Workflow"
                                    >
                                        <Trash2 className="h-3 w-3" />
                                    </Button>
                                </div>
                            </div>

                            <CollapsibleContent className="border-t bg-muted/10 p-2 space-y-2">
                                {workflow.steps.length === 0 ? (
                                    <div className="text-xs text-muted-foreground italic text-center py-2">
                                        No steps. Save a tool call to add one.
                                    </div>
                                ) : (
                                    workflow.steps.map((step, idx) => (
                                        <WorkflowItem
                                            key={step.id}
                                            step={step}
                                            index={idx}
                                            onRun={() => handleRunStep(workflow.id, step.id)}
                                            onDelete={() => handleDeleteStep(workflow.id, step.id)}
                                        />
                                    ))
                                )}
                            </CollapsibleContent>
                        </Collapsible>
                    ))}
                </div>
            </ScrollArea>
        </div>
    );
}
