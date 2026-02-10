/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState } from "react";
import Link from "next/link";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Layers, Pencil, Trash2, Box, Download, MoreHorizontal } from "lucide-react";
import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";

export interface Stack {
    name: string;
    description?: string;
    version?: string;
    author?: string;
    services: any[];
}

interface StackListProps {
    stacks: Stack[];
    isLoading?: boolean;
    onDelete: (name: string) => void;
    onExport?: (stack: Stack) => void;
}

export function StackList({ stacks, isLoading, onDelete, onExport }: StackListProps) {
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
            <div className="flex flex-col items-center justify-center py-20 border-2 border-dashed rounded-lg bg-muted/20">
                <Layers className="h-10 w-10 text-muted-foreground mb-4 opacity-50" />
                <h3 className="text-lg font-medium">No stacks found</h3>
                <p className="text-sm text-muted-foreground mb-4">Create a new stack to manage a collection of services.</p>
                <Link href="/stacks/new">
                    <Button>Create Stack</Button>
                </Link>
            </div>
        );
    }

    return (
        <div className="rounded-md border bg-card">
            <Table>
                <TableHeader>
                    <TableRow>
                        <TableHead>Name</TableHead>
                        <TableHead>Services</TableHead>
                        <TableHead>Description</TableHead>
                        <TableHead className="text-right">Actions</TableHead>
                    </TableRow>
                </TableHeader>
                <TableBody>
                    {stacks.map((stack) => (
                        <TableRow key={stack.name}>
                            <TableCell className="font-medium">
                                <div className="flex items-center gap-2">
                                    <div className="p-2 bg-primary/10 rounded-md">
                                        <Layers className="h-4 w-4 text-primary" />
                                    </div>
                                    <Link href={`/stacks/${stack.name}`} className="hover:underline font-semibold">
                                        {stack.name}
                                    </Link>
                                </div>
                            </TableCell>
                            <TableCell>
                                <Badge variant="secondary" className="flex w-fit items-center gap-1">
                                    <Box className="h-3 w-3" />
                                    {stack.services?.length || 0}
                                </Badge>
                            </TableCell>
                            <TableCell className="text-muted-foreground max-w-md truncate">
                                {stack.description || "-"}
                            </TableCell>
                            <TableCell className="text-right">
                                <DropdownMenu>
                                    <DropdownMenuTrigger asChild>
                                        <Button variant="ghost" className="h-8 w-8 p-0">
                                            <span className="sr-only">Open menu</span>
                                            <MoreHorizontal className="h-4 w-4" />
                                        </Button>
                                    </DropdownMenuTrigger>
                                    <DropdownMenuContent align="end">
                                        <DropdownMenuItem asChild>
                                            <Link href={`/stacks/${stack.name}`}>
                                                <Pencil className="mr-2 h-4 w-4" />
                                                Edit Configuration
                                            </Link>
                                        </DropdownMenuItem>
                                        {onExport && (
                                            <DropdownMenuItem onClick={() => onExport(stack)}>
                                                <Download className="mr-2 h-4 w-4" />
                                                Export
                                            </DropdownMenuItem>
                                        )}
                                        <DropdownMenuItem onClick={() => onDelete(stack.name)} className="text-destructive focus:text-destructive">
                                            <Trash2 className="mr-2 h-4 w-4" />
                                            Delete
                                        </DropdownMenuItem>
                                    </DropdownMenuContent>
                                </DropdownMenu>
                            </TableCell>
                        </TableRow>
                    ))}
                </TableBody>
            </Table>
        </div>
    );
}
