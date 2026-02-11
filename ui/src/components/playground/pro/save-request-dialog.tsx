/**
 * Copyright 2025 Author(s) of MCP Any
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
    DialogTrigger,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { usePlaygroundLibrary } from "@/hooks/use-playground-library";
import { Bookmark, Plus } from "lucide-react";
import { toast } from "sonner";

interface SaveRequestDialogProps {
    toolName: string;
    args: Record<string, unknown>;
    trigger?: React.ReactNode;
}

export function SaveRequestDialog({ toolName, args, trigger }: SaveRequestDialogProps) {
    const { collections, createCollection, saveRequestToCollection } = usePlaygroundLibrary();
    const [open, setOpen] = useState(false);
    const [requestName, setRequestName] = useState(`${toolName} Request`);
    const [selectedCollectionId, setSelectedCollectionId] = useState<string>("");
    const [newCollectionName, setNewCollectionName] = useState("");
    const [isCreatingCollection, setIsCreatingCollection] = useState(false);

    const handleSave = () => {
        let targetCollectionId = selectedCollectionId;

        if (isCreatingCollection) {
            if (!newCollectionName.trim()) return;
            const newCol = createCollection(newCollectionName);
            targetCollectionId = newCol.id;
        }

        if (!targetCollectionId) {
            toast.error("Please select or create a collection.");
            return;
        }

        saveRequestToCollection(targetCollectionId, {
            name: requestName,
            toolName,
            args,
        });

        toast.success("Request saved to library.");
        setOpen(false);
    };

    return (
        <Dialog open={open} onOpenChange={setOpen}>
            <DialogTrigger asChild>
                {trigger || (
                    <Button variant="ghost" size="icon" className="h-6 w-6 text-muted-foreground hover:text-foreground">
                        <Bookmark className="h-4 w-4" />
                    </Button>
                )}
            </DialogTrigger>
            <DialogContent>
                <DialogHeader>
                    <DialogTitle>Save Request</DialogTitle>
                    <DialogDescription>
                        Save this tool execution to your library for future use.
                    </DialogDescription>
                </DialogHeader>
                <div className="grid gap-4 py-4">
                    <div className="space-y-2">
                        <label className="text-sm font-medium">Request Name</label>
                        <Input value={requestName} onChange={(e) => setRequestName(e.target.value)} />
                    </div>
                    <div className="space-y-2">
                        <label className="text-sm font-medium">Collection</label>
                        {collections.length === 0 || isCreatingCollection ? (
                            <div className="flex gap-2">
                                <Input
                                    placeholder="New Collection Name"
                                    value={newCollectionName}
                                    onChange={(e) => setNewCollectionName(e.target.value)}
                                />
                                {collections.length > 0 && (
                                    <Button variant="ghost" onClick={() => setIsCreatingCollection(false)}>Cancel</Button>
                                )}
                            </div>
                        ) : (
                            <div className="flex gap-2">
                                <Select value={selectedCollectionId} onValueChange={setSelectedCollectionId}>
                                    <SelectTrigger className="flex-1">
                                        <SelectValue placeholder="Select a collection" />
                                    </SelectTrigger>
                                    <SelectContent>
                                        {collections.map((c) => (
                                            <SelectItem key={c.id} value={c.id}>{c.name}</SelectItem>
                                        ))}
                                    </SelectContent>
                                </Select>
                                <Button variant="outline" size="icon" onClick={() => setIsCreatingCollection(true)} title="Create New Collection">
                                    <Plus className="h-4 w-4" />
                                </Button>
                            </div>
                        )}
                    </div>
                </div>
                <DialogFooter>
                    <Button onClick={handleSave} disabled={!requestName || (!selectedCollectionId && !newCollectionName)}>
                        Save
                    </Button>
                </DialogFooter>
            </DialogContent>
        </Dialog>
    );
}
