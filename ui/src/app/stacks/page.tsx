/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useEffect } from "react";
import Link from "next/link";
import { Layers, Cuboid, Plus, Trash2, Loader2, FileText, MoreHorizontal } from "lucide-react";
import { Card, CardHeader, CardTitle, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle, DialogTrigger } from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { apiClient } from "@/lib/client";
import { useToast } from "@/hooks/use-toast";
import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
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

interface Stack {
    name: string;
    services: any[]; // Or more specific type
}

/**
 * StacksPage component.
 * @returns The rendered component.
 */
export default function StacksPage() {
    const [stacks, setStacks] = useState<Stack[]>([]);
    const [isLoading, setIsLoading] = useState(true);
    const [isCreateOpen, setIsCreateOpen] = useState(false);
    const [newStackName, setNewStackName] = useState("");
    const [isCreating, setIsCreating] = useState(false);
    const [stackToDelete, setStackToDelete] = useState<string | null>(null);
    const { toast } = useToast();

    const fetchStacks = async () => {
        setIsLoading(true);
        try {
            const res = await apiClient.listCollections();
            setStacks(Array.isArray(res) ? res : []);
        } catch (error) {
            console.error("Failed to list stacks", error);
            toast({
                title: "Error",
                description: "Failed to load stacks.",
                variant: "destructive",
            });
        } finally {
            setIsLoading(false);
        }
    };

    useEffect(() => {
        fetchStacks();
    }, []);

    const handleCreateStack = async () => {
        const name = newStackName.trim();
        if (!name) return;

        setIsCreating(true);
        try {
            await apiClient.createCollection({ name: name, services: [] });
            toast({
                title: "Stack Created",
                description: `Stack ${name} created successfully.`,
            });
            setIsCreateOpen(false);
            setNewStackName("");
            fetchStacks();
        } catch (error) {
            console.error(error);
            toast({
                title: "Error",
                description: "Failed to create stack.",
                variant: "destructive",
            });
        } finally {
            setIsCreating(false);
        }
    };

    const confirmDelete = async () => {
        if (!stackToDelete) return;
        try {
            await apiClient.deleteCollection(stackToDelete);
            toast({
                title: "Stack Deleted",
                description: `Stack ${stackToDelete} deleted successfully.`,
            });
            fetchStacks();
        } catch (error) {
            console.error(error);
            toast({
                title: "Error",
                description: "Failed to delete stack.",
                variant: "destructive",
            });
        } finally {
            setStackToDelete(null);
        }
    };

    return (
        <div className="space-y-6 p-8 pt-6">
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
                                Enter a unique name for your new configuration stack.
                            </DialogDescription>
                        </DialogHeader>
                        <div className="grid gap-4 py-4">
                            <div className="grid gap-2">
                                <Label htmlFor="name">Stack Name</Label>
                                <Input
                                    id="name"
                                    placeholder="my-stack"
                                    value={newStackName}
                                    onChange={(e) => setNewStackName(e.target.value)}
                                />
                            </div>
                        </div>
                        <DialogFooter>
                            <Button variant="outline" onClick={() => setIsCreateOpen(false)}>Cancel</Button>
                            <Button onClick={handleCreateStack} disabled={isCreating || !newStackName.trim()}>
                                {isCreating ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : null}
                                Create
                            </Button>
                        </DialogFooter>
                    </DialogContent>
                </Dialog>
            </div>

            {isLoading ? (
                 <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
                     {[...Array(3)].map((_, i) => (
                         <div key={i} className="h-40 rounded-xl bg-muted animate-pulse" />
                     ))}
                 </div>
            ) : (
                <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
                    {stacks.length === 0 && (
                        <div className="col-span-full flex flex-col items-center justify-center p-12 border-2 border-dashed rounded-lg bg-muted/10 text-muted-foreground">
                            <Layers className="h-12 w-12 mb-4 opacity-20" />
                            <h3 className="text-lg font-medium">No stacks found</h3>
                            <p className="mb-4">Create a new stack to get started.</p>
                            <Button variant="outline" onClick={() => setIsCreateOpen(true)}>Create Stack</Button>
                        </div>
                    )}
                    {stacks.map((stack) => (
                        <Card key={stack.name} className="hover:shadow-md transition-all group border-transparent shadow-sm bg-card hover:bg-muted/50 flex flex-col relative">
                            <Link href={`/stacks/${stack.name}`} className="flex-1">
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
                                        <div>
                                            <div className="text-2xl font-bold tracking-tight">{stack.name}</div>
                                            <div className="text-xs text-muted-foreground font-mono">
                                                {Array.isArray(stack.services) ? stack.services.length : 0} Services
                                            </div>
                                        </div>
                                    </div>
                                </CardContent>
                            </Link>
                             <div className="absolute top-4 right-4">
                                <DropdownMenu>
                                    <DropdownMenuTrigger asChild>
                                        <Button variant="ghost" size="icon" className="h-8 w-8 opacity-0 group-hover:opacity-100 transition-opacity">
                                            <span className="sr-only">Open menu</span>
                                            <MoreHorizontal className="h-4 w-4" />
                                        </Button>
                                    </DropdownMenuTrigger>
                                    <DropdownMenuContent align="end">
                                        <DropdownMenuItem asChild>
                                            <Link href={`/stacks/${stack.name}`}>
                                                <FileText className="mr-2 h-4 w-4" /> Edit Configuration
                                            </Link>
                                        </DropdownMenuItem>
                                        <DropdownMenuItem className="text-destructive focus:text-destructive" onClick={() => setStackToDelete(stack.name)}>
                                            <Trash2 className="mr-2 h-4 w-4" /> Delete Stack
                                        </DropdownMenuItem>
                                    </DropdownMenuContent>
                                </DropdownMenu>
                            </div>
                        </Card>
                    ))}
                </div>
            )}

            <AlertDialog open={!!stackToDelete} onOpenChange={(open) => !open && setStackToDelete(null)}>
                <AlertDialogContent>
                    <AlertDialogHeader>
                        <AlertDialogTitle>Are you sure?</AlertDialogTitle>
                        <AlertDialogDescription>
                            This action cannot be undone. This will permanently delete the stack <span className="font-medium text-foreground">"{stackToDelete}"</span>.
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
