/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState } from "react";
import { Folder, Plus, Play, Trash2, ChevronRight, ChevronDown, Save, MoreHorizontal } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { ScrollArea } from "@/components/ui/scroll-area";
import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import {
    AlertDialog,
    AlertDialogAction,
    AlertDialogCancel,
    AlertDialogContent,
    AlertDialogDescription,
    AlertDialogFooter,
    AlertDialogHeader,
    AlertDialogTitle,
} from "@/components/ui/alert-dialog";
import { Badge } from "@/components/ui/badge";
import { cn } from "@/lib/utils";
import { useToast } from "@/hooks/use-toast";

/**
 * SavedRequest represents a saved tool execution request.
 */
export interface SavedRequest {
    /** Unique identifier for the saved request */
    id: string;
    /** Display name of the request */
    name: string;
    /** The tool name (e.g. "get_weather") */
    toolName: string;
    /** The arguments used for the tool */
    toolArgs: Record<string, unknown>;
}

/**
 * Collection represents a group of saved requests.
 */
export interface Collection {
    /** Unique identifier for the collection */
    id: string;
    /** Display name of the collection */
    name: string;
    /** List of requests in this collection */
    requests: SavedRequest[];
}

interface CollectionsSidebarProps {
    /** List of collections */
    collections: Collection[];
    /** Function to update collections */
    setCollections: (cols: Collection[]) => void;
    /** Callback to run a request */
    onRunRequest: (toolName: string, args: Record<string, unknown>) => void;
    /** Optional className */
    className?: string;
}

/**
 * CollectionsSidebar allows users to manage and run saved request collections.
 *
 * @param props - The component props.
 * @param props.collections - The list of collections.
 * @param props.setCollections - Function to update collections.
 * @param props.onRunRequest - Callback when a request is executed.
 * @param props.className - Optional class name.
 * @returns The rendered sidebar component.
 */
export function CollectionsSidebar({ collections, setCollections, onRunRequest, className }: CollectionsSidebarProps) {
    const [expanded, setExpanded] = useState<Set<string>>(new Set());
    const [isCreating, setIsCreating] = useState(false);
    const [newCollectionName, setNewCollectionName] = useState("");
    const [collectionToDelete, setCollectionToDelete] = useState<string | null>(null);
    const { toast } = useToast();

    const toggleExpand = (id: string) => {
        const newExpanded = new Set(expanded);
        if (newExpanded.has(id)) {
            newExpanded.delete(id);
        } else {
            newExpanded.add(id);
        }
        setExpanded(newExpanded);
    };

    const createCollection = () => {
        if (!newCollectionName.trim()) return;
        const newCol: Collection = {
            id: crypto.randomUUID(),
            name: newCollectionName.trim(),
            requests: []
        };
        setCollections([...collections, newCol]);
        setNewCollectionName("");
        setIsCreating(false);
        toggleExpand(newCol.id); // Auto expand
        toast({ title: "Collection created" });
    };

    const confirmDeleteCollection = () => {
        if (collectionToDelete) {
            setCollections(collections.filter(c => c.id !== collectionToDelete));
            setCollectionToDelete(null);
            toast({ title: "Collection deleted" });
        }
    };

    const deleteRequest = (colId: string, reqId: string) => {
        setCollections(collections.map(c => {
            if (c.id === colId) {
                return { ...c, requests: c.requests.filter(r => r.id !== reqId) };
            }
            return c;
        }));
    };

    return (
        <div className={cn("flex flex-col h-full bg-muted/10 border-r", className)}>
            <div className="p-4 border-b space-y-3">
                <div className="flex items-center justify-between">
                    <div className="flex items-center gap-2 text-sm font-semibold text-muted-foreground">
                        <Folder className="h-4 w-4" />
                        Collections
                    </div>
                    <Button variant="ghost" size="icon" className="h-6 w-6" onClick={() => setIsCreating(true)}>
                        <Plus className="h-4 w-4" />
                    </Button>
                </div>

                {isCreating && (
                    <div className="flex gap-2">
                        <Input
                            value={newCollectionName}
                            onChange={(e) => setNewCollectionName(e.target.value)}
                            placeholder="Name..."
                            className="h-8 text-xs"
                            autoFocus
                            onKeyDown={(e) => {
                                if (e.key === 'Enter') createCollection();
                                if (e.key === 'Escape') setIsCreating(false);
                            }}
                        />
                        <Button size="sm" className="h-8 w-8 p-0" onClick={createCollection} disabled={!newCollectionName.trim()}>
                            <Save className="h-3 w-3" />
                        </Button>
                    </div>
                )}
            </div>

            <ScrollArea className="flex-1">
                <div className="p-2 space-y-1">
                    {collections.length === 0 && !isCreating && (
                        <div className="text-center py-8 text-muted-foreground text-xs">
                            No collections yet.
                        </div>
                    )}

                    {collections.map((col) => (
                        <div key={col.id} className="border rounded-md bg-card overflow-hidden">
                            <div className="flex items-center justify-between p-2 hover:bg-muted/50 cursor-pointer group" onClick={() => toggleExpand(col.id)}>
                                <div className="flex items-center gap-2 overflow-hidden">
                                    {expanded.has(col.id) ? <ChevronDown className="h-3 w-3 text-muted-foreground" /> : <ChevronRight className="h-3 w-3 text-muted-foreground" />}
                                    <span className="text-sm font-medium truncate">{col.name}</span>
                                    <Badge variant="secondary" className="text-[9px] h-4 px-1">{col.requests.length}</Badge>
                                </div>
                                <div className="flex items-center opacity-0 group-hover:opacity-100 transition-opacity">
                                    <DropdownMenu>
                                        <DropdownMenuTrigger asChild>
                                            <Button variant="ghost" size="icon" className="h-6 w-6" onClick={(e) => e.stopPropagation()}>
                                                <MoreHorizontal className="h-3 w-3" />
                                            </Button>
                                        </DropdownMenuTrigger>
                                        <DropdownMenuContent align="end">
                                            <DropdownMenuItem onClick={() => setCollectionToDelete(col.id)} className="text-destructive">
                                                <Trash2 className="mr-2 h-4 w-4" /> Delete
                                            </DropdownMenuItem>
                                        </DropdownMenuContent>
                                    </DropdownMenu>
                                </div>
                            </div>

                            {expanded.has(col.id) && (
                                <div className="bg-muted/20 border-t">
                                    {col.requests.length === 0 ? (
                                        <div className="p-2 text-[10px] text-muted-foreground text-center italic">Empty</div>
                                    ) : (
                                        col.requests.map((req) => (
                                            <div key={req.id} className="flex items-center justify-between p-2 pl-6 hover:bg-muted/50 text-xs group border-b last:border-0">
                                                <div className="flex flex-col truncate cursor-pointer flex-1" onClick={() => onRunRequest(req.toolName, req.toolArgs)}>
                                                    <div className="font-medium truncate">{req.name || req.toolName}</div>
                                                    <div className="text-muted-foreground truncate font-mono text-[10px]">{req.toolName}</div>
                                                </div>
                                                <div className="flex items-center opacity-0 group-hover:opacity-100 transition-opacity">
                                                    <Button variant="ghost" size="icon" className="h-5 w-5" onClick={() => onRunRequest(req.toolName, req.toolArgs)} title="Run">
                                                        <Play className="h-3 w-3" />
                                                    </Button>
                                                    <Button variant="ghost" size="icon" className="h-5 w-5 text-muted-foreground hover:text-destructive" onClick={() => deleteRequest(col.id, req.id)} title="Remove">
                                                        <Trash2 className="h-3 w-3" />
                                                    </Button>
                                                </div>
                                            </div>
                                        ))
                                    )}
                                </div>
                            )}
                        </div>
                    ))}
                </div>
            </ScrollArea>

            <AlertDialog open={!!collectionToDelete} onOpenChange={(open) => !open && setCollectionToDelete(null)}>
                <AlertDialogContent>
                    <AlertDialogHeader>
                        <AlertDialogTitle>Are you sure?</AlertDialogTitle>
                        <AlertDialogDescription>
                            This action cannot be undone. This will permanently delete the collection and all its saved requests.
                        </AlertDialogDescription>
                    </AlertDialogHeader>
                    <AlertDialogFooter>
                        <AlertDialogCancel>Cancel</AlertDialogCancel>
                        <AlertDialogAction onClick={confirmDeleteCollection} className="bg-destructive text-destructive-foreground hover:bg-destructive/90">
                            Delete
                        </AlertDialogAction>
                    </AlertDialogFooter>
                </AlertDialogContent>
            </AlertDialog>
        </div>
    );
}
