/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useEffect, useCallback } from "react";
import Link from "next/link";
import { Layers, Cuboid, Plus, Loader2 } from "lucide-react";
import { Card, CardHeader, CardTitle, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { apiClient } from "@/lib/client";
import { CreateStackDialog } from "./create-stack-dialog";
import { useToast } from "@/hooks/use-toast";

// Define a local interface since we can't easily import from proto in client-side code sometimes
interface Collection {
    name: string;
    description: string;
    version: string;
    priority: number;
    services: any[];
}

export function StackList() {
    const [stacks, setStacks] = useState<Collection[]>([]);
    const [loading, setLoading] = useState(true);
    const [isCreateOpen, setIsCreateOpen] = useState(false);
    const { toast } = useToast();

    const fetchStacks = useCallback(async () => {
        setLoading(true);
        try {
            const data = await apiClient.listCollections();
            // Backend might return array directly or object with list
            const list = Array.isArray(data) ? data : (data.collections || []);
            setStacks(list);
        } catch (error) {
            console.error("Failed to load stacks", error);
            toast({
                variant: "destructive",
                title: "Error",
                description: "Failed to load stacks.",
            });
        } finally {
            setLoading(false);
        }
    }, [toast]);

    useEffect(() => {
        fetchStacks();
    }, [fetchStacks]);

    if (loading && stacks.length === 0) {
        return (
            <div className="flex items-center justify-center h-64">
                <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
            </div>
        );
    }

    return (
        <div className="space-y-6">
            <div className="flex items-center justify-between">
                <div className="flex flex-col gap-2">
                    <h1 className="text-2xl font-bold tracking-tight">Stacks</h1>
                    <p className="text-muted-foreground">Manage your MCP Any configuration stacks.</p>
                </div>
                <Button onClick={() => setIsCreateOpen(true)}>
                    <Plus className="mr-2 h-4 w-4" /> Create Stack
                </Button>
            </div>

            {stacks.length === 0 ? (
                <div className="flex flex-col items-center justify-center h-64 border rounded-lg border-dashed bg-muted/10">
                    <Layers className="h-10 w-10 text-muted-foreground mb-4" />
                    <h3 className="text-lg font-medium">No stacks found</h3>
                    <p className="text-muted-foreground mb-4 text-center max-w-sm">
                        Create a stack to organize your upstream services into logical groups.
                    </p>
                    <Button variant="outline" onClick={() => setIsCreateOpen(true)}>
                        Create Stack
                    </Button>
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
                                        <div className="min-w-0">
                                            <div className="text-xl font-bold tracking-tight truncate" title={stack.name}>
                                                {stack.name}
                                            </div>
                                            <div className="text-xs text-muted-foreground font-mono truncate">
                                                v{stack.version || "0.0.1"}
                                            </div>
                                        </div>
                                    </div>

                                    {stack.description && (
                                        <p className="text-sm text-muted-foreground line-clamp-2 mb-4 h-10">
                                            {stack.description}
                                        </p>
                                    )}

                                    <div className="flex items-center justify-between text-xs text-muted-foreground mt-auto pt-4 border-t">
                                        <div className="flex items-center gap-1.5">
                                             {/* Placeholder logic for stack status - could be aggregated from services */}
                                            <span className="relative flex h-2 w-2">
                                                <span className="relative inline-flex rounded-full h-2 w-2 bg-green-500"></span>
                                            </span>
                                            Active
                                        </div>
                                        <div>
                                            {stack.services?.length || 0} Services
                                        </div>
                                    </div>
                                </CardContent>
                            </Card>
                        </Link>
                    ))}
                </div>
            )}

            <CreateStackDialog
                open={isCreateOpen}
                onOpenChange={setIsCreateOpen}
                onStackCreated={fetchStacks}
            />
        </div>
    );
}
