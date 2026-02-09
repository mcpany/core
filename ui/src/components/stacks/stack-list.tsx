/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useMemo, useState } from "react";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Input } from "@/components/ui/input";
import { Trash2, PlayCircle, Edit, Layers, Filter } from "lucide-react";
import { Collection } from "@proto/config/v1/collection";

interface StackListProps {
    stacks: Collection[];
    isLoading?: boolean;
    onEdit: (stack: Collection) => void;
    onDelete: (name: string) => void;
    onApply: (name: string) => void;
}

export function StackList({ stacks, isLoading, onEdit, onDelete, onApply }: StackListProps) {
    const [filter, setFilter] = useState("");

    const filteredStacks = useMemo(() => {
        if (!filter) return stacks;
        return stacks.filter(s => s.name.toLowerCase().includes(filter.toLowerCase()));
    }, [stacks, filter]);

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
        return <div className="text-center py-10 text-muted-foreground">No stacks found. Create one to get started.</div>;
    }

    return (
        <div className="space-y-4">
            <div className="flex items-center space-x-2 w-full md:w-1/3">
                <Filter className="h-4 w-4 text-muted-foreground" />
                <Input
                    placeholder="Filter stacks..."
                    value={filter}
                    onChange={(e) => setFilter(e.target.value)}
                    className="h-8"
                />
            </div>

            <div className="rounded-md border">
                <Table>
                    <TableHeader>
                        <TableRow>
                            <TableHead>Name</TableHead>
                            <TableHead>Description</TableHead>
                            <TableHead>Version</TableHead>
                            <TableHead>Services</TableHead>
                            <TableHead className="text-right">Actions</TableHead>
                        </TableRow>
                    </TableHeader>
                    <TableBody>
                        {filteredStacks.map((stack) => (
                            <TableRow key={stack.name}>
                                <TableCell className="font-medium flex items-center gap-2">
                                    <Layers className="h-4 w-4 text-primary" />
                                    {stack.name}
                                </TableCell>
                                <TableCell className="text-muted-foreground">{stack.description}</TableCell>
                                <TableCell>{stack.version}</TableCell>
                                <TableCell>
                                    <Badge variant="secondary">{stack.services?.length || 0}</Badge>
                                </TableCell>
                                <TableCell className="text-right">
                                    <div className="flex justify-end gap-2">
                                        <Button variant="outline" size="icon" onClick={() => onApply(stack.name)} title="Deploy Stack">
                                            <PlayCircle className="h-4 w-4 text-green-600" />
                                        </Button>
                                        <Button variant="ghost" size="icon" onClick={() => onEdit(stack)}>
                                            <Edit className="h-4 w-4" />
                                        </Button>
                                        <Button variant="ghost" size="icon" className="text-destructive" onClick={() => onDelete(stack.name)}>
                                            <Trash2 className="h-4 w-4" />
                                        </Button>
                                    </div>
                                </TableCell>
                            </TableRow>
                        ))}
                    </TableBody>
                </Table>
            </div>
        </div>
    );
}
