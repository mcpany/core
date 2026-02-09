/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import Link from "next/link";
import { Layers, Cuboid, Trash2, Play } from "lucide-react";
import { Card, CardHeader, CardTitle, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { ServiceCollection } from "@/lib/marketplace-service";

interface StackListProps {
    stacks: ServiceCollection[];
    onDelete: (name: string) => void;
    onDeploy: (name: string) => void;
}

export function StackList({ stacks, onDelete, onDeploy }: StackListProps) {
    if (stacks.length === 0) {
        return (
            <div className="text-center py-10 text-muted-foreground border-2 border-dashed rounded-lg">
                No stacks found. Create one to get started.
            </div>
        );
    }

    return (
        <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
            {stacks.map((stack) => (
                <Card key={stack.name} className="hover:shadow-md transition-all group border-transparent shadow-sm bg-card hover:bg-muted/50 flex flex-col">
                    <Link href={`/stacks/${stack.name}`} className="flex-1 cursor-pointer">
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
                                    <div className="text-2xl font-bold tracking-tight truncate" title={stack.name}>{stack.name}</div>
                                    <div className="text-xs text-muted-foreground font-mono truncate" title={stack.description}>{stack.description || "No description"}</div>
                                </div>
                            </div>

                            <div className="flex items-center justify-between text-xs text-muted-foreground mt-4 pt-4 border-t">
                                <div className="flex items-center gap-1.5">
                                    <Badge variant="secondary" className="font-normal">
                                        v{stack.version || "0.0.1"}
                                    </Badge>
                                </div>
                                <div>
                                    {stack.services?.length || 0} Services
                                </div>
                            </div>
                        </CardContent>
                    </Link>
                    <div className="px-6 pb-4 flex gap-2 justify-end opacity-0 group-hover:opacity-100 transition-opacity">
                         <Button
                            size="sm"
                            variant="outline"
                            className="h-8"
                            aria-label="Delete"
                            onClick={(e) => {
                                e.stopPropagation();
                                onDelete(stack.name);
                            }}
                        >
                            <Trash2 className="h-4 w-4 text-destructive" />
                        </Button>
                        <Button
                            size="sm"
                            className="h-8"
                            onClick={(e) => {
                                e.stopPropagation();
                                onDeploy(stack.name);
                            }}
                        >
                            <Play className="mr-2 h-3 w-3" /> Deploy
                        </Button>
                    </div>
                </Card>
            ))}
        </div>
    );
}
