/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useEffect } from "react";
import { Collection, CollectionService, TestCase } from "@/lib/collection-service";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { ScrollArea } from "@/components/ui/scroll-area";
import {
    Accordion,
    AccordionContent,
    AccordionItem,
    AccordionTrigger,
} from "@/components/ui/accordion";
import {
    Dialog,
    DialogContent,
    DialogHeader,
    DialogTitle,
    DialogFooter,
    DialogDescription
} from "@/components/ui/dialog";
import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Play, Plus, Trash2, Folder, MoreHorizontal, FileJson, Clock } from "lucide-react";
import { cn } from "@/lib/utils";
import { toast } from "@/hooks/use-toast";

interface CollectionsPanelProps {
    onRunTestCase: (toolName: string, args: Record<string, unknown>) => void;
    className?: string;
}

/**
 * Panel for managing and running test case collections.
 * Allows creating collections, adding/removing test cases, and executing them.
 *
 * @param props - The component props
 * @param props.onRunTestCase - Callback to execute a test case (tool execution)
 * @param props.className - Optional CSS class names
 */
export function CollectionsPanel({ onRunTestCase, className }: CollectionsPanelProps) {
    const [collections, setCollections] = useState<Collection[]>([]);
    const [isCreateOpen, setIsCreateOpen] = useState(false);
    const [newCollectionName, setNewCollectionName] = useState("");
    const [refreshTrigger, setRefreshTrigger] = useState(0);

    // Load collections on mount and refresh
    useEffect(() => {
        setCollections(CollectionService.list());

        const handleUpdate = () => {
            setCollections(CollectionService.list());
        };
        window.addEventListener("mcpany-collection-updated", handleUpdate);
        return () => window.removeEventListener("mcpany-collection-updated", handleUpdate);
    }, [refreshTrigger]);

    const refresh = () => {
        setRefreshTrigger(prev => prev + 1);
        window.dispatchEvent(new CustomEvent("mcpany-collection-updated"));
    };

    const handleCreateCollection = () => {
        if (!newCollectionName.trim()) return;
        const newCollection: Collection = {
            id: crypto.randomUUID(),
            name: newCollectionName.trim(),
            items: [],
            createdAt: Date.now()
        };
        CollectionService.save(newCollection);
        setNewCollectionName("");
        setIsCreateOpen(false);
        refresh();
        toast({ title: "Collection Created", description: newCollection.name });
    };

    const handleDeleteCollection = (id: string, name: string) => {
        if (confirm(`Delete collection "${name}" and all its test cases?`)) {
            CollectionService.delete(id);
            refresh();
            toast({ title: "Collection Deleted" });
        }
    };

    const handleDeleteTestCase = (collectionId: string, testCaseId: string) => {
        CollectionService.removeTestCase(collectionId, testCaseId);
        refresh();
        toast({ title: "Test Case Deleted" });
    };

    return (
        <div className={cn("flex flex-col h-full bg-muted/10 border-r", className)}>
            <div className="p-4 border-b flex items-center justify-between">
                <div className="flex items-center gap-2 text-sm font-semibold text-muted-foreground">
                    <Folder className="h-4 w-4" />
                    Collections
                </div>
                <Button variant="ghost" size="icon" className="h-8 w-8" onClick={() => setIsCreateOpen(true)}>
                    <Plus className="h-4 w-4" />
                </Button>
            </div>

            <ScrollArea className="flex-1">
                {collections.length === 0 ? (
                    <div className="text-center py-8 text-muted-foreground text-xs p-4">
                        No collections found. Create one to organize your test cases.
                    </div>
                ) : (
                    <Accordion type="multiple" className="w-full">
                        {collections.map((col) => (
                            <AccordionItem key={col.id} value={col.id} className="border-b-0 px-2">
                                <div className="flex items-center group">
                                    <AccordionTrigger className="flex-1 py-2 text-sm font-medium hover:no-underline px-2 rounded-md hover:bg-muted/50">
                                        <span className="truncate">{col.name}</span>
                                        <span className="ml-2 text-xs text-muted-foreground font-normal no-underline">
                                            ({col.items.length})
                                        </span>
                                    </AccordionTrigger>
                                    <DropdownMenu>
                                        <DropdownMenuTrigger asChild>
                                            <Button variant="ghost" size="icon" className="h-6 w-6 opacity-0 group-hover:opacity-100 transition-opacity">
                                                <MoreHorizontal className="h-3 w-3" />
                                            </Button>
                                        </DropdownMenuTrigger>
                                        <DropdownMenuContent align="end">
                                            <DropdownMenuItem onClick={() => handleDeleteCollection(col.id, col.name)} className="text-destructive">
                                                <Trash2 className="mr-2 h-4 w-4" /> Delete Collection
                                            </DropdownMenuItem>
                                        </DropdownMenuContent>
                                    </DropdownMenu>
                                </div>
                                <AccordionContent className="pb-2 pl-4">
                                    <div className="space-y-1">
                                        {col.items.length === 0 && (
                                            <div className="text-xs text-muted-foreground italic py-2">
                                                No test cases yet. Run a tool and click "Save to Collection".
                                            </div>
                                        )}
                                        {col.items.map((item) => (
                                            <div key={item.id} className="flex flex-col gap-1 p-2 rounded-md border bg-card hover:border-primary/30 group/item relative">
                                                <div className="flex items-center justify-between">
                                                    <span className="font-semibold text-xs truncate max-w-[150px]" title={item.name}>{item.name}</span>
                                                    <div className="flex items-center gap-1 opacity-0 group-hover/item:opacity-100 transition-opacity">
                                                        <Button
                                                            variant="ghost"
                                                            size="icon"
                                                            className="h-6 w-6 text-green-600 hover:text-green-700 hover:bg-green-100 dark:hover:bg-green-900/30"
                                                            onClick={() => onRunTestCase(item.toolName, item.args)}
                                                            title="Run"
                                                        >
                                                            <Play className="h-3 w-3" />
                                                        </Button>
                                                        <Button
                                                            variant="ghost"
                                                            size="icon"
                                                            className="h-6 w-6 text-muted-foreground hover:text-destructive"
                                                            onClick={() => handleDeleteTestCase(col.id, item.id)}
                                                            title="Delete"
                                                        >
                                                            <Trash2 className="h-3 w-3" />
                                                        </Button>
                                                    </div>
                                                </div>
                                                <div className="flex items-center gap-2 text-[10px] text-muted-foreground">
                                                    <span className="bg-primary/10 text-primary px-1 rounded">{item.toolName}</span>
                                                    <span className="flex items-center gap-0.5" title={new Date(item.createdAt).toLocaleString()}>
                                                        <Clock className="h-3 w-3" /> {new Date(item.createdAt).toLocaleDateString()}
                                                    </span>
                                                </div>
                                            </div>
                                        ))}
                                    </div>
                                </AccordionContent>
                            </AccordionItem>
                        ))}
                    </Accordion>
                )}
            </ScrollArea>

            <Dialog open={isCreateOpen} onOpenChange={setIsCreateOpen}>
                <DialogContent>
                    <DialogHeader>
                        <DialogTitle>Create Collection</DialogTitle>
                        <DialogDescription>
                            Group related tool executions into a collection.
                        </DialogDescription>
                    </DialogHeader>
                    <div className="py-4">
                        <Input
                            placeholder="Collection Name (e.g. Smoke Tests)"
                            value={newCollectionName}
                            onChange={(e) => setNewCollectionName(e.target.value)}
                            onKeyDown={(e) => e.key === 'Enter' && handleCreateCollection()}
                        />
                    </div>
                    <DialogFooter>
                        <Button variant="outline" onClick={() => setIsCreateOpen(false)}>Cancel</Button>
                        <Button onClick={handleCreateCollection}>Create</Button>
                    </DialogFooter>
                </DialogContent>
            </Dialog>
        </div>
    );
}
