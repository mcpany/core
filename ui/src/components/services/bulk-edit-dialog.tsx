/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState } from "react";
import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Tag, Clock } from "lucide-react";

interface BulkEditDialogProps {
    open: boolean;
    onOpenChange: (open: boolean) => void;
    selectedCount: number;
    onApply: (updates: { tags?: string[], resilience?: { timeout?: string } }) => void;
    onCancel: () => void;
}

/**
 * BulkEditDialog component.
 * Allows bulk editing of tags and configuration for multiple services.
 *
 * @param props - The component props.
 * @param props.open - Whether the dialog is open.
 * @param props.onOpenChange - Callback when the dialog open state changes.
 * @param props.selectedCount - The number of selected services.
 * @param props.onApply - Callback when changes are applied.
 * @param props.onCancel - Callback when the dialog is cancelled.
 * @returns The rendered component.
 */
export function BulkEditDialog({ open, onOpenChange, selectedCount, onApply, onCancel }: BulkEditDialogProps) {
    const [tags, setTags] = useState("");
    const [timeout, setTimeout] = useState("");
    const [activeTab, setActiveTab] = useState("tags");

    const handleApply = () => {
        const updates: { tags?: string[], resilience?: { timeout?: string } } = {};

        // Only include fields that have values (basic logic for now)
        // Ideally we might want checkboxes to indicate intent to update specific fields
        if (tags.trim()) {
            updates.tags = tags.split(",").map(t => t.trim()).filter(Boolean);
        }

        if (timeout.trim()) {
            updates.resilience = { timeout: timeout.trim() };
        }

        onApply(updates);
        // Reset
        setTags("");
        setTimeout("");
        onOpenChange(false);
    };

    return (
        <Dialog open={open} onOpenChange={onOpenChange}>
            <DialogContent className="sm:max-w-md">
                <DialogHeader>
                    <DialogTitle>Bulk Edit Services</DialogTitle>
                    <DialogDescription>
                        Update {selectedCount} selected services. Changes will be applied to all selected services.
                    </DialogDescription>
                </DialogHeader>

                <Tabs value={activeTab} onValueChange={setActiveTab} className="w-full">
                    <TabsList className="grid w-full grid-cols-2">
                        <TabsTrigger value="tags" className="flex items-center gap-2">
                            <Tag className="h-4 w-4" /> Tags
                        </TabsTrigger>
                        <TabsTrigger value="config" className="flex items-center gap-2">
                            <Clock className="h-4 w-4" /> Configuration
                        </TabsTrigger>
                    </TabsList>

                    <div className="py-4">
                        <TabsContent value="tags" className="mt-0 space-y-4">
                            <div className="space-y-2">
                                <Label htmlFor="bulk-tags">Add Tags (comma separated)</Label>
                                <Input
                                    id="bulk-tags"
                                    placeholder="production, web, internal"
                                    value={tags}
                                    onChange={(e) => setTags(e.target.value)}
                                />
                                <p className="text-xs text-muted-foreground">
                                    These tags will be added to the existing tags of selected services.
                                </p>
                            </div>
                        </TabsContent>

                        <TabsContent value="config" className="mt-0 space-y-4">
                            <div className="space-y-2">
                                <Label htmlFor="bulk-timeout">Timeout</Label>
                                <Input
                                    id="bulk-timeout"
                                    placeholder="30s"
                                    value={timeout}
                                    onChange={(e) => setTimeout(e.target.value)}
                                />
                                <p className="text-xs text-muted-foreground">
                                    Set the request timeout (e.g. 10s, 1m). This overwrites existing timeouts.
                                </p>
                            </div>
                        </TabsContent>
                    </div>
                </Tabs>

                <DialogFooter>
                    <Button variant="outline" onClick={onCancel}>Cancel</Button>
                    <Button onClick={handleApply}>Apply Changes</Button>
                </DialogFooter>
            </DialogContent>
        </Dialog>
    );
}
