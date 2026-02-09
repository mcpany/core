/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import Link from "next/link";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Layers, Pencil, Trash2, Box } from "lucide-react";
import { formatDistanceToNow } from "date-fns";

export interface Stack {
  name: string;
  description: string;
  author: string;
  version: string;
  services: any[];
  createdAt?: string; // Optional if not returned by backend yet
}

interface StackListProps {
  stacks: Stack[];
  isLoading: boolean;
  onDelete: (name: string) => void;
}

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
        <div className="flex flex-col items-center justify-center py-12 text-center border-2 border-dashed rounded-lg bg-muted/20">
            <Layers className="h-10 w-10 text-muted-foreground mb-4 opacity-50" />
            <h3 className="text-lg font-medium">No stacks defined</h3>
            <p className="text-sm text-muted-foreground max-w-sm mt-1">
                Stacks allow you to group multiple services and deploy them together using a single configuration file.
            </p>
        </div>
      );
  }

  return (
    <div className="rounded-md border bg-background/50 backdrop-blur-sm">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>Name</TableHead>
            <TableHead>Description</TableHead>
            <TableHead>Services</TableHead>
            <TableHead>Version</TableHead>
            <TableHead>Author</TableHead>
            <TableHead className="text-right">Actions</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {stacks.map((stack) => (
            <TableRow key={stack.name}>
              <TableCell className="font-medium">
                <Link href={`/stacks/${stack.name}`} className="flex items-center gap-2 hover:underline">
                    <Layers className="h-4 w-4 text-primary" />
                    {stack.name}
                </Link>
              </TableCell>
              <TableCell className="max-w-xs truncate text-muted-foreground">
                {stack.description || "-"}
              </TableCell>
              <TableCell>
                <Badge variant="secondary" className="gap-1">
                    <Box className="h-3 w-3" />
                    {stack.services?.length || 0}
                </Badge>
              </TableCell>
              <TableCell>{stack.version}</TableCell>
              <TableCell className="text-muted-foreground text-xs">{stack.author}</TableCell>
              <TableCell className="text-right">
                <div className="flex justify-end gap-2">
                    <Button variant="ghost" size="icon" asChild>
                        <Link href={`/stacks/${stack.name}`}>
                            <Pencil className="h-4 w-4" />
                        </Link>
                    </Button>
                    <Button
                        variant="ghost"
                        size="icon"
                        className="text-destructive hover:text-destructive hover:bg-destructive/10"
                        onClick={() => onDelete(stack.name)}
                    >
                        <Trash2 className="h-4 w-4" />
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
