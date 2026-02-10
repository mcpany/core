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
import { apiClient } from "@/lib/client";
import {
    Dialog,
    DialogContent,
    DialogHeader,
    DialogTitle,
    DialogDescription,
    DialogFooter,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { useToast } from "@/hooks/use-toast";
import { useRouter } from "next/navigation";

/**
 * StacksPage component.
 * Displays a list of service stacks and allows creating new ones.
 * @returns The rendered component.
 */
export default function StacksPage() {
    const [collections, setCollections] = useState<any[]>([]);
    const [loading, setLoading] = useState(true);
    const [isCreateOpen, setIsCreateOpen] = useState(false);
    const [newStackName, setNewStackName] = useState("");
    const [creating, setCreating] = useState(false);
    const { toast } = useToast();
    const router = useRouter();

    useEffect(() => {
        loadCollections();
    }, []);

    const loadCollections = async () => {
        try {
            const data = await apiClient.listCollections();
            setCollections(data || []);
        } catch (e) {
            console.error(e);
            toast({
                variant: "destructive",
                title: "Error",
                description: "Failed to load stacks."
            });
        } finally {
            setLoading(false);
        }
    };

    const handleCreate = async () => {
        if (!newStackName.trim()) return;
        setCreating(true);
        try {
            await apiClient.saveCollection({
                name: newStackName,
                services: []
            });
            toast({
                title: "Stack Created",
                description: `Stack ${newStackName} has been created.`
            });
            setIsCreateOpen(false);
            setNewStackName("");
            loadCollections();
            router.push(`/stacks/${newStackName}`);
        } catch (e) {
            console.error(e);
            toast({
                variant: "destructive",
                title: "Error",
                description: "Failed to create stack."
            });
        } finally {
            setCreating(false);
        }
    };

    return (
        <div className="space-y-6 flex-1 p-8 pt-6">
            <div className="flex items-center justify-between">
                <div className="flex flex-col gap-2">
                    <h1 className="text-2xl font-bold tracking-tight">Stacks</h1>
                    <p className="text-muted-foreground">Manage your MCP Any configuration stacks.</p>
                </div>
                <Button onClick={() => setIsCreateOpen(true)}>
                    <Plus className="mr-2 h-4 w-4" /> Create Stack
                </Button>
            </div>

            {loading ? (
                <div className="flex justify-center py-10">
                    <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
                </div>
            ) : (
                <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
                    {collections.length === 0 && (
                        <div className="col-span-full text-center text-muted-foreground py-10">
                            No stacks found. Create one to get started.
                        </div>
                    )}
                    {collections.map((stack) => (
                        <Link key={stack.name} href={`/stacks/${stack.name}`}>
                            <Card className="hover:shadow-md transition-all cursor-pointer group border-transparent shadow-sm bg-card hover:bg-muted/50">
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
                                            <div className="text-xs text-muted-foreground font-mono truncate max-w-[200px]">
                                                {stack.id || stack.name}
                                            </div>
                                        </div>
                                    </div>

                                    <div className="flex items-center justify-between text-xs text-muted-foreground mt-4 pt-4 border-t">
                                        <div className="flex items-center gap-1.5">
                                            <span className="relative flex h-2 w-2">
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

            <Dialog open={isCreateOpen} onOpenChange={setIsCreateOpen}>
                <DialogContent>
                    <DialogHeader>
                        <DialogTitle>Create New Stack</DialogTitle>
                        <DialogDescription>
                            Enter a name for your new configuration stack.
                        </DialogDescription>
                    </DialogHeader>
                    <div className="grid gap-4 py-4">
                        <div className="grid gap-2">
                            <Label htmlFor="name">Stack Name</Label>
                            <Input
                                id="name"
                                value={newStackName}
                                onChange={(e) => setNewStackName(e.target.value)}
                                placeholder="my-stack"
                                onKeyDown={(e) => e.key === "Enter" && handleCreate()}
                            />
                        </div>
                    </div>
                    <DialogFooter>
                        <Button variant="outline" onClick={() => setIsCreateOpen(false)}>Cancel</Button>
                        <Button onClick={handleCreate} disabled={creating || !newStackName.trim()}>
                            {creating && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
                            Create
                        </Button>
                    </DialogFooter>
                </DialogContent>
            </Dialog>
        </div>
    );
}
