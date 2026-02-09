/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState } from "react";
import { ServiceCollection } from "@/lib/marketplace-service";
import { Card, CardHeader, CardTitle, CardDescription, CardContent, CardFooter } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Layers, Edit, Play, Trash2, Box } from "lucide-react";
import { formatDistanceToNow } from "date-fns";

interface StackListProps {
    stacks: ServiceCollection[];
    onEdit: (stack: ServiceCollection) => void;
    onApply: (stack: ServiceCollection) => void;
    onDelete: (stack: ServiceCollection) => void;
    isLoading?: boolean;
}

export function StackList({ stacks, onEdit, onApply, onDelete, isLoading }: StackListProps) {
    if (isLoading) {
        return <div className="text-muted-foreground animate-pulse">Loading stacks...</div>;
    }

    if (stacks.length === 0) {
        return (
            <div className="flex flex-col items-center justify-center py-12 border-2 border-dashed rounded-lg bg-muted/10">
                <Box className="h-10 w-10 text-muted-foreground mb-4 opacity-50" />
                <h3 className="text-lg font-medium">No Stacks Found</h3>
                <p className="text-sm text-muted-foreground mb-4">Create a new stack to manage multiple services together.</p>
            </div>
        );
    }

    return (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            {stacks.map((stack) => (
                <Card key={stack.name} className="flex flex-col hover:shadow-md transition-all">
                    <CardHeader className="pb-2">
                        <div className="flex justify-between items-start">
                            <div className="flex items-center gap-2">
                                <div className="p-2 bg-primary/10 rounded-md">
                                    <Layers className="h-5 w-5 text-primary" />
                                </div>
                                <div>
                                    <CardTitle className="text-lg">{stack.name}</CardTitle>
                                    <CardDescription className="text-xs font-mono">{stack.version || "latest"}</CardDescription>
                                </div>
                            </div>
                        </div>
                    </CardHeader>
                    <CardContent className="flex-1 py-2">
                        <p className="text-sm text-muted-foreground line-clamp-2 mb-4 h-10">
                            {stack.description || "No description provided."}
                        </p>
                        <div className="flex flex-wrap gap-1">
                            <Badge variant="secondary" className="text-xs">
                                {stack.services?.length || 0} Services
                            </Badge>
                            {stack.author && (
                                <Badge variant="outline" className="text-xs">
                                    {stack.author}
                                </Badge>
                            )}
                        </div>
                    </CardContent>
                    <CardFooter className="pt-2 border-t bg-muted/10 flex justify-between gap-2">
                        <Button variant="ghost" size="sm" onClick={() => onDelete(stack)} className="text-destructive hover:text-destructive hover:bg-destructive/10">
                            <Trash2 className="h-4 w-4" />
                        </Button>
                        <div className="flex gap-2">
                            <Button variant="outline" size="sm" onClick={() => onEdit(stack)}>
                                <Edit className="h-4 w-4 mr-2" /> Edit
                            </Button>
                            <Button size="sm" onClick={() => onApply(stack)}>
                                <Play className="h-4 w-4 mr-2" /> Apply
                            </Button>
                        </div>
                    </CardFooter>
                </Card>
            ))}
        </div>
    );
}
