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
import Link from "next/link";
import { Settings, Trash2, Layers, Search, Play } from "lucide-react";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"

interface Stack {
    name: string;
    description?: string;
    services?: any[];
    version?: string;
}

interface StackListProps {
  stacks: Stack[];
  isLoading?: boolean;
  onDelete?: (name: string) => void;
  onDeploy?: (name: string) => void;
}

export function StackList({ stacks, isLoading, onDelete, onDeploy }: StackListProps) {
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
      <div className="flex items-center justify-between">
          <div className="flex items-center space-x-2 w-full md:w-1/3">
            <Search className="h-4 w-4 text-muted-foreground" />
            <Input
              placeholder="Filter stacks..."
              value={filter}
              onChange={(e) => setFilter(e.target.value)}
              className="h-8"
            />
          </div>
      </div>

      <div className="rounded-md border">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Name</TableHead>
              <TableHead>Services</TableHead>
              <TableHead>Version</TableHead>
              <TableHead className="text-right">Actions</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {filteredStacks.map((stack) => (
               <TableRow key={stack.name}>
                 <TableCell className="font-medium">
                     <div className="flex items-center gap-2">
                         <Layers className="h-4 w-4 text-muted-foreground" />
                         <Link href={`/stacks/${stack.name}`} className="hover:underline font-semibold text-primary">
                            {stack.name}
                         </Link>
                     </div>
                     {stack.description && <div className="text-xs text-muted-foreground ml-6">{stack.description}</div>}
                 </TableCell>
                 <TableCell>
                     <Badge variant="secondary">{stack.services?.length || 0} Services</Badge>
                 </TableCell>
                 <TableCell>
                     {stack.version || "0.0.1"}
                 </TableCell>
                 <TableCell className="text-right">
                     <DropdownMenu>
                        <DropdownMenuTrigger asChild>
                            <Button variant="ghost" className="h-8 w-8 p-0">
                                <span className="sr-only">Open menu</span>
                                <Settings className="h-4 w-4" />
                            </Button>
                        </DropdownMenuTrigger>
                        <DropdownMenuContent align="end">
                            <DropdownMenuLabel>Actions</DropdownMenuLabel>
                            <DropdownMenuItem asChild>
                                <Link href={`/stacks/${stack.name}`}>
                                    <Settings className="mr-2 h-4 w-4" />
                                    Edit Configuration
                                </Link>
                            </DropdownMenuItem>
                            {onDeploy && (
                                <DropdownMenuItem onClick={() => onDeploy(stack.name)}>
                                    <Play className="mr-2 h-4 w-4 text-green-500" />
                                    Deploy Stack
                                </DropdownMenuItem>
                            )}
                            <DropdownMenuSeparator />
                            {onDelete && (
                                <DropdownMenuItem onClick={() => onDelete(stack.name)} className="text-destructive focus:text-destructive">
                                    <Trash2 className="mr-2 h-4 w-4" />
                                    Delete
                                </DropdownMenuItem>
                            )}
                        </DropdownMenuContent>
                     </DropdownMenu>
                 </TableCell>
               </TableRow>
            ))}
          </TableBody>
        </Table>
      </div>
    </div>
  );
}
