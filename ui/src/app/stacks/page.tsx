/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useEffect } from "react";
import { apiClient } from "@/lib/client";
import { ServiceCollection } from "@/lib/marketplace-service";
import { StackList } from "@/components/stacks/stack-list";
import { StackEditor } from "@/components/stacks/stack-editor";
import { Button } from "@/components/ui/button";
import { Plus, Download } from "lucide-react";
import { useToast } from "@/hooks/use-toast";
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription } from "@/components/ui/dialog";

export default function StacksPage() {
    const [stacks, setStacks] = useState<ServiceCollection[]>([]);
    const [loading, setLoading] = useState(true);
    const [selectedStackId, setSelectedStackId] = useState<string | null>(null);
    const [isEditorOpen, setIsEditorOpen] = useState(false);
    const { toast } = useToast();

    const fetchStacks = async () => {
        setLoading(true);
        try {
            const res = await apiClient.listCollections();
            setStacks(res || []);
        } catch (e) {
            console.error("Failed to fetch stacks", e);
            toast({
                title: "Error",
                description: "Failed to load stacks.",
                variant: "destructive"
            });
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        fetchStacks();
    }, []);

    const handleCreate = () => {
        setSelectedStackId("new");
        setIsEditorOpen(true);
    };

    const handleEdit = (stack: ServiceCollection) => {
        setSelectedStackId(stack.name);
        setIsEditorOpen(true);
    };

    const handleDelete = async (stack: ServiceCollection) => {
        if (!confirm(`Are you sure you want to delete stack "${stack.name}"?`)) return;
        try {
            await apiClient.deleteCollection(stack.name);
            fetchStacks();
            toast({ title: "Stack Deleted", description: `Stack ${stack.name} has been removed.` });
        } catch (e: any) {
            toast({ title: "Delete Failed", description: e.message, variant: "destructive" });
        }
    };

    const handleApply = async (stack: ServiceCollection) => {
        try {
            await apiClient.applyStack(stack.name);
            toast({ title: "Stack Applied", description: `Services in ${stack.name} have been updated.` });
        } catch (e: any) {
            toast({ title: "Apply Failed", description: e.message, variant: "destructive" });
        }
    };

    return (
        <div className="flex-1 space-y-4 p-8 pt-6 h-[calc(100vh-4rem)] flex flex-col">
            <div className="flex items-center justify-between">
                <div>
                    <h2 className="text-3xl font-bold tracking-tight">Stacks</h2>
                    <p className="text-muted-foreground">
                        Manage infrastructure as code using YAML stacks.
                    </p>
                </div>
                <Button onClick={handleCreate}>
                    <Plus className="mr-2 h-4 w-4" /> Add Stack
                </Button>
            </div>

            <StackList
                stacks={stacks}
                isLoading={loading}
                onEdit={handleEdit}
                onDelete={handleDelete}
                onApply={handleApply}
            />

            <Dialog open={isEditorOpen} onOpenChange={setIsEditorOpen}>
                <DialogContent className="max-w-4xl h-[80vh] flex flex-col">
                    <DialogHeader>
                        <DialogTitle>{selectedStackId === "new" ? "New Stack" : `Edit ${selectedStackId}`}</DialogTitle>
                        <DialogDescription>
                            Define your services in YAML format.
                        </DialogDescription>
                    </DialogHeader>
                    {selectedStackId && (
                        <StackEditor
                            stackId={selectedStackId}
                            onClose={() => setIsEditorOpen(false)}
                            onSaved={() => {
                                fetchStacks();
                                setIsEditorOpen(false);
                            }}
                        />
                    )}
                </DialogContent>
            </Dialog>
        </div>
    );
}
