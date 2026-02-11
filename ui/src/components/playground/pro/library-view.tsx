/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState } from "react";
import { usePlaygroundLibrary } from "@/hooks/use-playground-library";
import { Accordion, AccordionContent, AccordionItem, AccordionTrigger } from "@/components/ui/accordion";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Plus, Play, Trash2, Folder, FileJson, MoreVertical } from "lucide-react";
import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
    DialogTrigger,
} from "@/components/ui/dialog";
import { SavedRequest } from "@/types/playground-library";
import { ScrollArea } from "@/components/ui/scroll-area";

interface LibraryViewProps {
    onRunRequest: (toolName: string, args: Record<string, unknown>) => void;
}

export function LibraryView({ onRunRequest }: LibraryViewProps) {
    const { collections, createCollection, deleteCollection, deleteRequestFromCollection } = usePlaygroundLibrary();
    const [isCreateOpen, setIsCreateOpen] = useState(false);
    const [newCollectionName, setNewCollectionName] = useState("");

    const handleCreateCollection = () => {
        if (!newCollectionName.trim()) return;
        createCollection(newCollectionName);
        setNewCollectionName("");
        setIsCreateOpen(false);
    };

    return (
        <div className="flex flex-col h-full">
            <div className="p-4 border-b flex items-center justify-between bg-muted/10">
                <h3 className="font-semibold text-sm flex items-center gap-2">
                    <Folder className="h-4 w-4" /> Library
                </h3>
                <Dialog open={isCreateOpen} onOpenChange={setIsCreateOpen}>
                    <DialogTrigger asChild>
                        <Button variant="ghost" size="icon" className="h-8 w-8">
                            <Plus className="h-4 w-4" />
                        </Button>
                    </DialogTrigger>
                    <DialogContent>
                        <DialogHeader>
                            <DialogTitle>New Collection</DialogTitle>
                            <DialogDescription>
                                Create a collection to organize your saved requests.
                            </DialogDescription>
                        </DialogHeader>
                        <div className="py-4">
                            <Input
                                placeholder="Collection Name"
                                value={newCollectionName}
                                onChange={(e) => setNewCollectionName(e.target.value)}
                                onKeyDown={(e) => e.key === 'Enter' && handleCreateCollection()}
                            />
                        </div>
                        <DialogFooter>
                            <Button onClick={handleCreateCollection} disabled={!newCollectionName.trim()}>Create</Button>
                        </DialogFooter>
                    </DialogContent>
                </Dialog>
            </div>

            <ScrollArea className="flex-1">
                {collections.length === 0 ? (
                    <div className="p-8 text-center text-muted-foreground text-sm flex flex-col items-center gap-2">
                        <Folder className="h-8 w-8 opacity-20" />
                        <p>No collections found.</p>
                        <Button variant="outline" size="sm" onClick={() => setIsCreateOpen(true)}>Create One</Button>
                    </div>
                ) : (
                    <Accordion type="multiple" className="w-full">
                        {collections.map((collection) => (
                            <AccordionItem key={collection.id} value={collection.id}>
                                <div className="flex items-center justify-between pr-4 hover:bg-muted/50 group/item">
                                    <AccordionTrigger className="px-4 py-3 text-sm font-medium hover:no-underline flex-1">
                                        {collection.name}
                                        <span className="ml-2 text-xs text-muted-foreground font-normal">
                                            ({collection.requests.length})
                                        </span>
                                    </AccordionTrigger>
                                    <DropdownMenu>
                                        <DropdownMenuTrigger asChild>
                                            <Button variant="ghost" size="icon" className="h-6 w-6 opacity-0 group-hover/item:opacity-100 transition-opacity">
                                                <MoreVertical className="h-3 w-3" />
                                            </Button>
                                        </DropdownMenuTrigger>
                                        <DropdownMenuContent align="end">
                                            <DropdownMenuItem className="text-destructive focus:text-destructive" onClick={() => deleteCollection(collection.id)}>
                                                <Trash2 className="mr-2 h-3 w-3" /> Delete Collection
                                            </DropdownMenuItem>
                                        </DropdownMenuContent>
                                    </DropdownMenu>
                                </div>
                                <AccordionContent className="px-0 pb-0">
                                    <div className="flex flex-col divide-y border-t bg-muted/10">
                                        {collection.requests.length === 0 ? (
                                            <div className="p-4 text-xs text-muted-foreground text-center italic">
                                                No saved requests.
                                            </div>
                                        ) : (
                                            collection.requests.map((req) => (
                                                <div key={req.id} className="flex items-center justify-between p-2 pl-6 hover:bg-accent/50 group/req transition-colors">
                                                    <div className="flex-1 min-w-0 pr-2">
                                                        <div className="text-sm font-medium truncate">{req.name}</div>
                                                        <div className="text-[10px] text-muted-foreground truncate flex items-center gap-1">
                                                            <FileJson className="h-3 w-3" /> {req.toolName}
                                                        </div>
                                                    </div>
                                                    <div className="flex items-center gap-1 opacity-0 group-hover/req:opacity-100 transition-opacity">
                                                        <Button
                                                            variant="ghost"
                                                            size="icon"
                                                            className="h-6 w-6 text-green-600 hover:text-green-700 hover:bg-green-100 dark:hover:bg-green-900/30"
                                                            onClick={() => onRunRequest(req.toolName, req.args)}
                                                            title="Run"
                                                        >
                                                            <Play className="h-3 w-3" />
                                                        </Button>
                                                        <Button
                                                            variant="ghost"
                                                            size="icon"
                                                            className="h-6 w-6 text-muted-foreground hover:text-destructive hover:bg-destructive/10"
                                                            onClick={() => deleteRequestFromCollection(collection.id, req.id)}
                                                            title="Delete"
                                                        >
                                                            <Trash2 className="h-3 w-3" />
                                                        </Button>
                                                    </div>
                                                </div>
                                            ))
                                        )}
                                    </div>
                                </AccordionContent>
                            </AccordionItem>
                        ))}
                    </Accordion>
                )}
            </ScrollArea>
        </div>
    );
}
