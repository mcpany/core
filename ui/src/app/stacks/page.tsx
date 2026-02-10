/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useEffect } from "react";
import Link from "next/link";
import { Layers, Cuboid, Loader2, Plus, RefreshCw } from "lucide-react";
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { apiClient } from "@/lib/client";
import { useToast } from "@/hooks/use-toast";
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
import { useRouter } from "next/navigation";

interface Collection {
    name: string;
    services: any[];
}

/**
 * StacksPage component.
 * @returns The rendered component.
 */
export default function StacksPage() {
    const [collections, setCollections] = useState<Collection[]>([]);
    const [loading, setLoading] = useState(true);
    const { toast } = useToast();
    const router = useRouter();

    // Create Stack State
    const [isCreateOpen, setIsCreateOpen] = useState(false);
    const [newStackName, setNewStackName] = useState("");
    const [isCreating, setIsCreating] = useState(false);

    const fetchCollections = async () => {
        setLoading(true);
        try {
            const data = await apiClient.listCollections();
            // Ensure data is array
            setCollections(Array.isArray(data) ? data : []);
        } catch (error) {
            console.error("Failed to load stacks", error);
            toast({
                variant: "destructive",
                title: "Error",
                description: "Failed to load stacks."
            });
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        fetchCollections();
    }, []);

    const handleCreateStack = async () => {
        if (!newStackName.trim()) {
            toast({
                variant: "destructive",
                title: "Invalid Name",
                description: "Stack name cannot be empty."
            });
            return;
        }

        setIsCreating(true);
        try {
            // Create empty stack
            await apiClient.saveCollection({
                name: newStackName,
                services: []
            });
            toast({
                title: "Stack Created",
                description: `Stack "${newStackName}" has been created.`
            });
            setIsCreateOpen(false);
            setNewStackName("");
            // Redirect to editor
            router.push(`/stacks/${newStackName}`);
        } catch (error) {
            console.error("Failed to create stack", error);
            toast({
                variant: "destructive",
                title: "Creation Failed",
                description: "Failed to create new stack."
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
                <div className="flex items-center gap-2">
                    <Button variant="outline" size="sm" onClick={fetchCollections} disabled={loading}>
                        <RefreshCw className={`mr-2 h-4 w-4 ${loading ? 'animate-spin' : ''}`} />
                        Refresh
                    </Button>
                    <Dialog open={isCreateOpen} onOpenChange={setIsCreateOpen}>
                        <DialogTrigger asChild>
                            <Button>
                                <Plus className="mr-2 h-4 w-4" /> Create Stack
                            </Button>
                        </DialogTrigger>
                        <DialogContent className="sm:max-w-[425px]">
                            <DialogHeader>
                                <DialogTitle>Create New Stack</DialogTitle>
                                <DialogDescription>
                                    Enter a unique name for your new configuration stack.
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
                                        placeholder="my-new-stack"
                                        className="col-span-3"
                                        autoFocus
                                        onKeyDown={(e) => {
                                            if (e.key === "Enter") handleCreateStack();
                                        }}
                                    />
                                </div>
                            </div>
                            <DialogFooter>
                                <Button onClick={handleCreateStack} disabled={isCreating || !newStackName.trim()}>
                                    {isCreating && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
                                    Create & Edit
                                </Button>
                            </DialogFooter>
                        </DialogContent>
                    </Dialog>
                </div>
            </div>

            {loading && collections.length === 0 ? (
                <div className="flex items-center justify-center h-64">
                    <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
                </div>
            ) : collections.length === 0 ? (
                <div className="flex flex-col items-center justify-center h-64 border rounded-lg bg-muted/10 border-dashed">
                    <Layers className="h-10 w-10 text-muted-foreground mb-4" />
                    <h3 className="text-lg font-medium">No Stacks Found</h3>
                    <p className="text-sm text-muted-foreground mb-4">Create your first stack to get started.</p>
                    <Button onClick={() => setIsCreateOpen(true)}>Create Stack</Button>
                </div>
            ) : (
                <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
                    {collections.map((stack) => (
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
                                            <div className="text-xs text-muted-foreground font-mono truncate">ID: {stack.name}</div>
                                        </div>
                                    </div>

                                    <div className="flex items-center justify-between text-xs text-muted-foreground mt-4 pt-4 border-t">
                                        <div className="flex items-center gap-1.5">
                                            <span className="relative flex h-2 w-2">
                                                <span className="relative inline-flex rounded-full h-2 w-2 bg-green-500"></span>
                                            </span>
                                            Configured
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
