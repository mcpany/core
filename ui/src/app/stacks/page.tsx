/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useEffect, useCallback } from "react";
import { apiClient } from "@/lib/client";
import { StackList, Collection } from "@/components/stacks/stack-list";
import { CreateStackDialog } from "@/components/stacks/create-stack-dialog";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { toast } from "@/hooks/use-toast";
import {
    AlertDialog,
    AlertDialogAction,
    AlertDialogCancel,
    AlertDialogContent,
    AlertDialogDescription,
    AlertDialogFooter,
    AlertDialogHeader,
    AlertDialogTitle,
} from "@/components/ui/alert-dialog";

/**
 * StacksPage component.
 * @returns The rendered component.
 */
export default function StacksPage() {
    const [stacks, setStacks] = useState<Collection[]>([]);
    const [loading, setLoading] = useState(true);
    const [stackToDelete, setStackToDelete] = useState<string | null>(null);

    const fetchStacks = useCallback(async () => {
        setLoading(true);
        try {
            const res = await apiClient.listCollections();
            if (Array.isArray(res)) {
                setStacks(res);
            } else {
                setStacks([]);
            }
        } catch (e) {
            console.error("Failed to fetch stacks", e);
            toast({
                variant: "destructive",
                title: "Error",
                description: "Failed to load stacks."
            });
        } finally {
            setLoading(false);
        }
    }, []);

    useEffect(() => {
        fetchStacks();
    }, [fetchStacks]);

    const confirmDelete = async () => {
        if (!stackToDelete) return;

        try {
            await apiClient.deleteCollection(stackToDelete);
            toast({
                title: "Stack Deleted",
                description: `Stack "${stackToDelete}" has been removed.`
            });
            fetchStacks();
        } catch (e) {
            console.error("Failed to delete stack", e);
            toast({
                variant: "destructive",
                title: "Error",
                description: "Failed to delete stack."
            });
        } finally {
            setStackToDelete(null);
        }
    };

    return (
        <div className="flex-1 space-y-4 p-8 pt-6">
            <div className="flex items-center justify-between">
                <div>
                    <h2 className="text-3xl font-bold tracking-tight">Stacks</h2>
                    <p className="text-muted-foreground">Manage your MCP Any configuration stacks.</p>
                </div>
                <CreateStackDialog onStackCreated={fetchStacks} />
            </div>

            <Card className="backdrop-blur-sm bg-background/50">
                <CardHeader>
                    <CardTitle>Stacks</CardTitle>
                    <CardDescription>
                        Configuration collections grouping multiple services.
                    </CardDescription>
                </CardHeader>
                <CardContent>
                    <StackList
                        stacks={stacks}
                        isLoading={loading}
                        onDelete={setStackToDelete}
                    />
                </CardContent>
            </Card>

            <AlertDialog open={!!stackToDelete} onOpenChange={(open) => !open && setStackToDelete(null)}>
                <AlertDialogContent>
                    <AlertDialogHeader>
                        <AlertDialogTitle>Are you absolutely sure?</AlertDialogTitle>
                        <AlertDialogDescription>
                            This action cannot be undone. This will permanently delete the stack
                            <span className="font-medium text-foreground"> {stackToDelete} </span>
                            and all its associated service configurations.
                        </AlertDialogDescription>
                    </AlertDialogHeader>
                    <AlertDialogFooter>
                        <AlertDialogCancel>Cancel</AlertDialogCancel>
                        <AlertDialogAction onClick={confirmDelete} className="bg-destructive text-destructive-foreground hover:bg-destructive/90">
                            Delete
                        </AlertDialogAction>
                    </AlertDialogFooter>
                </AlertDialogContent>
            </AlertDialog>
        </div>
    );
}
