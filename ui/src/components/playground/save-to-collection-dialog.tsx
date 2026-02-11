/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useEffect } from "react";
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Collection, CollectionService, TestCase } from "@/lib/collection-service";
import { toast } from "@/hooks/use-toast";
import { Label } from "@/components/ui/label";

interface SaveToCollectionDialogProps {
    open: boolean;
    onOpenChange: (open: boolean) => void;
    toolName: string;
    toolArgs: Record<string, unknown>;
}

/**
 * Dialog component for saving a tool execution as a test case in a collection.
 *
 * @param props - The component props
 * @param props.open - Whether the dialog is open
 * @param props.onOpenChange - Callback when open state changes
 * @param props.toolName - The name of the tool to save
 * @param props.toolArgs - The arguments used for the tool
 */
export function SaveToCollectionDialog({ open, onOpenChange, toolName, toolArgs }: SaveToCollectionDialogProps) {
    const [collections, setCollections] = useState<Collection[]>([]);
    const [selectedCollectionId, setSelectedCollectionId] = useState<string>("");
    const [testCaseName, setTestCaseName] = useState("");

    useEffect(() => {
        if (open) {
            const list = CollectionService.list();
            setCollections(list);
            if (list.length > 0) {
                setSelectedCollectionId(list[0].id);
            }
            setTestCaseName(`Test ${toolName}`);
        }
    }, [open, toolName]);

    const handleSave = () => {
        if (!selectedCollectionId || !testCaseName.trim()) return;

        const newTestCase: TestCase = {
            id: crypto.randomUUID(),
            name: testCaseName.trim(),
            toolName: toolName,
            args: toolArgs,
            createdAt: Date.now()
        };

        CollectionService.addTestCase(selectedCollectionId, newTestCase);

        toast({
            title: "Saved to Collection",
            description: `Added "${testCaseName}" to collection.`
        });

        onOpenChange(false);
        // Force refresh of CollectionsPanel?
        // Since both use CollectionService reading from localStorage, we might need a custom event or context.
        // For now, assume CollectionsPanel polls or we rely on page refresh/re-mount or parent state update.
        // Actually, I added a `refreshTrigger` in CollectionsPanel but that's local state.
        // To update CollectionsPanel immediately, we should use an event listener or context.
        // I'll stick to simple implementation: The user might need to click "Refresh" or reload if they have the panel open.
        // Or better, trigger a window storage event?
        // window.dispatchEvent(new Event("storage"));
        // CollectionsPanel doesn't listen to storage event.
        // I will dispatch a custom event.
        window.dispatchEvent(new CustomEvent("mcpany-collection-updated"));
    };

    return (
        <Dialog open={open} onOpenChange={onOpenChange}>
            <DialogContent>
                <DialogHeader>
                    <DialogTitle>Save to Collection</DialogTitle>
                    <DialogDescription>
                        Save this tool execution as a reusable test case.
                    </DialogDescription>
                </DialogHeader>
                <div className="grid gap-4 py-4">
                    <div className="grid gap-2">
                        <Label>Test Case Name</Label>
                        <Input
                            value={testCaseName}
                            onChange={(e) => setTestCaseName(e.target.value)}
                            placeholder="e.g. Valid Input Test"
                        />
                    </div>
                    <div className="grid gap-2">
                        <Label>Collection</Label>
                        <Select value={selectedCollectionId} onValueChange={setSelectedCollectionId}>
                            <SelectTrigger>
                                <SelectValue placeholder="Select a collection" />
                            </SelectTrigger>
                            <SelectContent>
                                {collections.map((c) => (
                                    <SelectItem key={c.id} value={c.id}>{c.name}</SelectItem>
                                ))}
                            </SelectContent>
                        </Select>
                        {collections.length === 0 && (
                            <p className="text-xs text-destructive">No collections found. Create one in the sidebar first.</p>
                        )}
                    </div>
                </div>
                <DialogFooter>
                    <Button variant="outline" onClick={() => onOpenChange(false)}>Cancel</Button>
                    <Button onClick={handleSave} disabled={collections.length === 0}>Save</Button>
                </DialogFooter>
            </DialogContent>
        </Dialog>
    );
}
