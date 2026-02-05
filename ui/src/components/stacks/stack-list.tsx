/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useMemo } from "react";
import Link from "next/link";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Layers, Trash2, ArrowRight } from "lucide-react";

/**
 * Collection represents a stack of services.
 */
export interface Collection {
    name: string;
    services?: any[];
    created_at?: string; // Assuming we might have this, if not we'll handle gracefully
}

interface StackListProps {
    stacks: Collection[];
    isLoading: boolean;
    onDelete: (name: string) => void;
}

/**
 * StackList displays a table of available stacks.
 * @param props The component props.
 * @param props.stacks The list of stacks to display.
 * @param props.isLoading Whether the data is loading.
 * @param props.onDelete Callback when a stack is deleted.
 * @returns The rendered component.
 */
export function StackList({ stacks, isLoading, onDelete }: StackListProps) {
    if (isLoading) {
        return (
            <div className="space-y-4">
                 {[...Array(3)].map((_, i) => (
                    <div key={i} className="w-full h-12 bg-muted animate-pulse rounded-md" />
                 ))}
            </div>
        );
    }

    if (stacks.length === 0) {
        return (
            <div className="flex flex-col items-center justify-center py-10 text-muted-foreground border rounded-md border-dashed">
                <Layers className="h-10 w-10 mb-4 opacity-50" />
                <p>No stacks found.</p>
                <p className="text-sm">Create a new stack to get started.</p>
            </div>
        );
    }

    return (
        <div className="rounded-md border">
            <Table>
                <TableHeader>
                    <TableRow>
                        <TableHead>Name</TableHead>
                        <TableHead>Services</TableHead>
                        <TableHead className="text-right">Actions</TableHead>
                    </TableRow>
                </TableHeader>
                <TableBody>
                    {stacks.map((stack) => (
                        <TableRow key={stack.name}>
                            <TableCell className="font-medium">
                                <Link href={`/stacks/${stack.name}`} className="flex items-center gap-2 hover:underline decoration-primary underline-offset-4">
                                    <Layers className="h-4 w-4 text-muted-foreground" />
                                    {stack.name}
                                </Link>
                            </TableCell>
                            <TableCell>
                                <Badge variant="secondary">
                                    {stack.services?.length || 0} Services
                                </Badge>
                            </TableCell>
                            <TableCell className="text-right">
                                <div className="flex items-center justify-end gap-2">
                                    <Button variant="ghost" size="icon" asChild>
                                        <Link href={`/stacks/${stack.name}`}>
                                            <ArrowRight className="h-4 w-4" />
                                            <span className="sr-only">View</span>
                                        </Link>
                                    </Button>
                                    <Button
                                        variant="ghost"
                                        size="icon"
                                        className="text-destructive hover:text-destructive hover:bg-destructive/10"
                                        onClick={() => onDelete(stack.name)}
                                    >
                                        <Trash2 className="h-4 w-4" />
                                        <span className="sr-only">Delete</span>
                                    </Button>
                                </div>
                            </TableCell>
                        </TableRow>
                    ))}
                </TableBody>
            </Table>
        </div>
    );
}
