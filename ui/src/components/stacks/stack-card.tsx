/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import Link from "next/link";
import { Layers, Cuboid, Trash2, Edit, Activity } from "lucide-react";
import { Card, CardHeader, CardTitle, CardContent, CardFooter } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import {
    AlertDialog,
    AlertDialogAction,
    AlertDialogCancel,
    AlertDialogContent,
    AlertDialogDescription,
    AlertDialogFooter,
    AlertDialogHeader,
    AlertDialogTitle,
    AlertDialogTrigger,
  } from "@/components/ui/alert-dialog";

/**
 * Stack interface defining the structure of a stack object.
 */
export interface Stack {
    name: string;
    description?: string;
    services?: unknown[];
}

interface StackCardProps {
    stack: Stack;
    onDelete: (name: string) => void;
}

/**
 * StackCard component displays a single stack's details in a card format.
 *
 * @param props - The component props.
 * @param props.stack - The stack object to display.
 * @param props.onDelete - Callback function to handle stack deletion.
 * @returns The rendered card component.
 */
export function StackCard({ stack, onDelete }: StackCardProps) {
    return (
        <Card className="hover:shadow-md transition-all border-transparent shadow-sm bg-card hover:bg-muted/50 flex flex-col h-full group">
            <Link href={`/stacks/${stack.name}`} className="flex-1">
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
                            <div className="text-2xl font-bold tracking-tight truncate" title={stack.name}>{stack.name}</div>
                            {stack.description && (
                                <div className="text-xs text-muted-foreground truncate" title={stack.description}>
                                    {stack.description}
                                </div>
                            )}
                        </div>
                    </div>

                    <div className="flex items-center justify-between text-xs text-muted-foreground mt-4 pt-4 border-t">
                        <div className="flex items-center gap-1.5">
                            <Activity className="h-3 w-3" />
                            <span>
                                {stack.services?.length || 0} Services
                            </span>
                        </div>
                    </div>
                </CardContent>
            </Link>
            <CardFooter className="pt-0 pb-4 px-6 flex justify-between gap-2 opacity-0 group-hover:opacity-100 transition-opacity">
                <Button variant="ghost" size="sm" asChild className="h-8 px-2 text-xs">
                     <Link href={`/stacks/${stack.name}`}>
                        <Edit className="mr-1 h-3 w-3" /> Edit
                     </Link>
                </Button>
                <AlertDialog>
                    <AlertDialogTrigger asChild>
                         <Button variant="ghost" size="sm" className="h-8 px-2 text-xs text-destructive hover:text-destructive hover:bg-destructive/10">
                            <Trash2 className="mr-1 h-3 w-3" /> Delete
                        </Button>
                    </AlertDialogTrigger>
                    <AlertDialogContent>
                        <AlertDialogHeader>
                            <AlertDialogTitle>Are you absolutely sure?</AlertDialogTitle>
                            <AlertDialogDescription>
                                This will permanently delete the stack <strong>{stack.name}</strong>.
                                Access to services defined in this stack may be lost.
                            </AlertDialogDescription>
                        </AlertDialogHeader>
                        <AlertDialogFooter>
                            <AlertDialogCancel>Cancel</AlertDialogCancel>
                            <AlertDialogAction onClick={() => onDelete(stack.name)} className="bg-destructive text-destructive-foreground hover:bg-destructive/90">
                                Delete Stack
                            </AlertDialogAction>
                        </AlertDialogFooter>
                    </AlertDialogContent>
                </AlertDialog>
            </CardFooter>
        </Card>
    );
}
