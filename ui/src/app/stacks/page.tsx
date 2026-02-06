/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useEffect } from "react";
import Link from "next/link";
import { Layers, Cuboid, Plus, Loader2 } from "lucide-react";
import { Card, CardHeader, CardTitle, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
    DialogTrigger,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { apiClient } from "@/lib/client";
import { useToast } from "@/hooks/use-toast";

interface Collection {
    name: string;
    description?: string;
    services?: unknown[];
}

/**
 * StacksPage component.
 * @returns The rendered component.
 */
export default function StacksPage() {
    const [stacks, setStacks] = useState<Collection[]>([]);
    const [isLoading, setIsLoading] = useState(true);
    const [isCreateOpen, setIsCreateOpen] = useState(false);
    const [newStackName, setNewStackName] = useState("");
    const [isCreating, setIsCreating] = useState(false);
    const { toast } = useToast();

    const fetchStacks = async () => {
        setIsLoading(true);
        try {
            const res = await apiClient.listCollections();
            setStacks(res || []);
        } catch (e) {
            console.error("Failed to fetch stacks", e);
            toast({
                variant: "destructive",
                title: "Error",
                description: "Failed to load stacks."
            });
        } finally {
            setIsLoading(false);
        }
    };

    useEffect(() => {
        fetchStacks();
    }, []);

    const handleCreate = async () => {
        if (!newStackName.trim()) return;
        setIsCreating(true);
        try {
            // Create empty collection
            await apiClient.saveCollection({
                name: newStackName,
                description: "New stack",
                services: []
            });
            toast({
                title: "Stack Created",
                description: `Stack "${newStackName}" created successfully.`
            });
            setIsCreateOpen(false);
            setNewStackName("");
            // Refresh list
            fetchStacks();
        } catch (e) {
            console.error("Failed to create stack", e);
            toast({
                variant: "destructive",
                title: "Error",
                description: "Failed to create stack."
            });
        } finally {
            setIsCreating(false);
        }
    };

    return (
        <div className="space-y-6 flex-1 p-8 pt-6">
            <div className="flex items-center justify-between">
                <div className="flex flex-col gap-2">
                    <h1 className="text-3xl font-bold tracking-tight">Stacks</h1>
                    <p className="text-muted-foreground">Manage your MCP Any configuration stacks.</p>
                </div>
                <Dialog open={isCreateOpen} onOpenChange={setIsCreateOpen}>
                    <DialogTrigger asChild>
                        <Button>
                            <Plus className="mr-2 h-4 w-4" /> Create Stack
                        </Button>
                    </DialogTrigger>
                    <DialogContent>
                        <DialogHeader>
                            <DialogTitle>Create New Stack</DialogTitle>
                            <DialogDescription>
                                Enter a name for your new service stack.
                            </DialogDescription>
                        </DialogHeader>
                        <div className="grid gap-4 py-4">
                            <div className="grid grid-cols-4 items-center gap-4">
                                <Label htmlFor="name" className="text-right">
                                    Name
                                </Label>
                                <Input
                                    id="name"
                                    value={newStackName}
                                    onChange={(e) => setNewStackName(e.target.value)}
                                    className="col-span-3"
                                    placeholder="my-stack"
                                />
                            </div>
                        </div>
                        <DialogFooter>
                            <Button variant="outline" onClick={() => setIsCreateOpen(false)}>Cancel</Button>
                            <Button onClick={handleCreate} disabled={!newStackName.trim() || isCreating}>
                                {isCreating && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
                                Create
                            </Button>
                        </DialogFooter>
                    </DialogContent>
                </Dialog>
            </div>

            {isLoading ? (
                <div className="flex items-center justify-center h-64">
                    <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
                </div>
            ) : stacks.length === 0 ? (
                <div className="text-center py-12 border-2 border-dashed rounded-lg">
                    <Layers className="h-10 w-10 text-muted-foreground mx-auto mb-4" />
                    <h3 className="text-lg font-medium">No stacks found</h3>
                    <p className="text-muted-foreground mb-4">Create a new stack to get started.</p>
                    <Button onClick={() => setIsCreateOpen(true)}>Create Stack</Button>
                </div>
            ) : (
                <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
                    {stacks.map((stack) => (
                        <Link key={stack.name} href={`/stacks/${stack.name}`}>
                            <Card className="hover:shadow-md transition-all cursor-pointer group border-transparent shadow-sm bg-card hover:bg-muted/50 h-full">
                                <CardHeader className="flex flex-row items-center justify-between pb-2 space-y-0">
                                    <CardTitle className="text-sm font-medium text-muted-foreground">
                                        Stack
                                    </CardTitle>
                                    <Cuboid className="h-4 w-4 text-muted-foreground group-hover:text-primary transition-colors" />
                                </CardHeader>
                                <CardContent>
                                    <div className="flex items-center gap-3 mb-4">
                                        <div className="p-2.5 bg-primary/10 rounded-lg group-hover:scale-105 transition-transform">
                                            <Layers className="h-6 w-6 text-primary" />
                                        </div>
                                        <div className="overflow-hidden">
                                            <div className="text-2xl font-bold tracking-tight truncate" title={stack.name}>{stack.name}</div>
                                            <div className="text-xs text-muted-foreground truncate">{stack.description || "No description"}</div>
                                        </div>
                                    </div>

                                    <div className="flex items-center justify-between text-xs text-muted-foreground mt-4 pt-4 border-t">
                                        <div className="flex items-center gap-1.5">
                                            <span className="relative flex h-2 w-2">
                                                <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-green-400 opacity-75"></span>
                                                <span className="relative inline-flex rounded-full h-2 w-2 bg-green-500"></span>
                                            </span>
                                            Active
                                        </div>
                                        <div>
                                            {stack.services ? stack.services.length : 0} Services
                                        </div>
                                    </div>
                                </CardContent>
                            </Card>
                        </Link>
                    ))}
                </div>
            )}
        </div>
    );
}
