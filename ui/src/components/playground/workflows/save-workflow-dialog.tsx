/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState } from "react";
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Workflow } from "@/types/workflow";

interface SaveWorkflowDialogProps {
    open: boolean;
    onOpenChange: (open: boolean) => void;
    existingWorkflows: Workflow[];
    onSave: (workflowId: string | null, newName: string) => void;
}

/**
 * SaveWorkflowDialog component.
 * @param props - The component props.
 * @param props.open - The open state.
 * @param props.onOpenChange - The onOpenChange handler.
 * @param props.existingWorkflows - The existingWorkflows list.
 * @param props.onSave - The onSave handler.
 * @returns The rendered component.
 */
export function SaveWorkflowDialog({ open, onOpenChange, existingWorkflows, onSave }: SaveWorkflowDialogProps) {
    const [mode, setMode] = useState<"new" | "existing">("new");
    const [newName, setNewName] = useState("");
    const [selectedId, setSelectedId] = useState<string>("");

    const handleSave = () => {
        if (mode === "new") {
            if (!newName) return;
            onSave(null, newName);
        } else {
            if (!selectedId) return;
            onSave(selectedId, "");
        }
        onOpenChange(false);
        setNewName("");
        setSelectedId("");
    };

    return (
        <Dialog open={open} onOpenChange={onOpenChange}>
            <DialogContent className="sm:max-w-[425px]">
                <DialogHeader>
                    <DialogTitle>Save to Workflow</DialogTitle>
                    <DialogDescription>
                        Add this request as a test step to a workflow.
                    </DialogDescription>
                </DialogHeader>
                <div className="grid gap-4 py-4">
                    <div className="flex flex-col gap-2">
                        <Label>Save Destination</Label>
                        <div className="flex gap-2">
                            <Button
                                variant={mode === "new" ? "default" : "outline"}
                                onClick={() => setMode("new")}
                                className="flex-1"
                            >
                                New Workflow
                            </Button>
                            <Button
                                variant={mode === "existing" ? "default" : "outline"}
                                onClick={() => setMode("existing")}
                                className="flex-1"
                                disabled={existingWorkflows.length === 0}
                            >
                                Existing
                            </Button>
                        </div>
                    </div>

                    {mode === "new" ? (
                        <div className="grid gap-2">
                            <Label htmlFor="name">Workflow Name</Label>
                            <Input
                                id="name"
                                value={newName}
                                onChange={(e) => setNewName(e.target.value)}
                                placeholder="e.g. User Registration Flow"
                                autoFocus
                            />
                        </div>
                    ) : (
                        <div className="grid gap-2">
                            <Label>Select Workflow</Label>
                            <Select value={selectedId} onValueChange={setSelectedId}>
                                <SelectTrigger>
                                    <SelectValue placeholder="Select a workflow..." />
                                </SelectTrigger>
                                <SelectContent>
                                    {existingWorkflows.map(w => (
                                        <SelectItem key={w.id} value={w.id}>{w.name}</SelectItem>
                                    ))}
                                </SelectContent>
                            </Select>
                        </div>
                    )}
                </div>
                <DialogFooter>
                    <Button variant="outline" onClick={() => onOpenChange(false)}>Cancel</Button>
                    <Button onClick={handleSave} disabled={mode === "new" ? !newName : !selectedId}>Save</Button>
                </DialogFooter>
            </DialogContent>
        </Dialog>
    );
}
