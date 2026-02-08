/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { Plus, Layers, Cuboid, AlertCircle, RefreshCw } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card, CardHeader, CardTitle, CardContent, CardDescription, CardFooter } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { apiClient } from "@/lib/client";
import { ServiceCollection } from "@/lib/marketplace-service";

export function StackList() {
    const [stacks, setStacks] = useState<ServiceCollection[]>([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);

    const fetchStacks = async () => {
        setLoading(true);
        setError(null);
        try {
            const data = await apiClient.listCollections();
            // Handle both array and object response formats
            if (Array.isArray(data)) {
                setStacks(data);
            } else if (data && Array.isArray(data.collections)) {
                setStacks(data.collections);
            } else {
                setStacks([]);
            }
        } catch (err: any) {
            console.error("Failed to fetch stacks", err);
            setError(err.message || "Failed to load stacks.");
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        fetchStacks();
    }, []);

    if (loading) {
        return (
            <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
                {[...Array(3)].map((_, i) => (
                    <Skeleton key={i} className="h-[200px] w-full rounded-xl" />
                ))}
            </div>
        );
    }

    if (error) {
        return (
            <Alert variant="destructive">
                <AlertCircle className="h-4 w-4" />
                <AlertTitle>Error</AlertTitle>
                <AlertDescription className="flex items-center justify-between">
                    <span>{error}</span>
                    <Button variant="outline" size="sm" onClick={fetchStacks}>
                        <RefreshCw className="mr-2 h-4 w-4" /> Retry
                    </Button>
                </AlertDescription>
            </Alert>
        );
    }

    return (
        <div className="space-y-6">
            <div className="flex items-center justify-between">
                <div>
                    <h2 className="text-lg font-medium">Your Stacks</h2>
                    <p className="text-sm text-muted-foreground">
                        Manage and deploy collections of services.
                    </p>
                </div>
                <Link href="/stacks/create">
                    <Button>
                        <Plus className="mr-2 h-4 w-4" /> Create Stack
                    </Button>
                </Link>
            </div>

            {stacks.length === 0 ? (
                <div className="flex flex-col items-center justify-center p-12 border-2 border-dashed rounded-lg bg-muted/20 text-center">
                    <Layers className="h-12 w-12 text-muted-foreground opacity-50 mb-4" />
                    <h3 className="text-lg font-medium">No Stacks Found</h3>
                    <p className="text-sm text-muted-foreground max-w-sm mb-6">
                        Create a stack to group multiple services together for easy deployment and management.
                    </p>
                    <Link href="/stacks/create">
                        <Button variant="outline">Create your first Stack</Button>
                    </Link>
                </div>
            ) : (
                <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
                    {stacks.map((stack) => (
                        <Link key={stack.name} href={`/stacks/${stack.name}`}>
                            <Card className="h-full hover:shadow-md transition-all cursor-pointer group border-transparent shadow-sm bg-card hover:bg-muted/50">
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
                                            <div className="text-xl font-bold tracking-tight truncate" title={stack.name}>{stack.name}</div>
                                            <div className="text-xs text-muted-foreground truncate" title={stack.description}>{stack.description || "No description"}</div>
                                        </div>
                                    </div>
                                </CardContent>
                                <CardFooter className="pt-0 border-t mt-auto">
                                    <div className="flex items-center justify-between w-full text-xs text-muted-foreground pt-4">
                                        <div className="flex items-center gap-2">
                                            <Badge variant="secondary" className="font-normal">
                                                v{stack.version || "1.0.0"}
                                            </Badge>
                                        </div>
                                        <div>
                                            {stack.services?.length || 0} Services
                                        </div>
                                    </div>
                                </CardFooter>
                            </Card>
                        </Link>
                    ))}
                </div>
            )}
        </div>
    );
}
