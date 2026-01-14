/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState } from "react";
import { DragDropContext, Droppable, Draggable, DropResult } from "@hello-pangea/dnd";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Switch } from "@/components/ui/switch";
import { GripVertical, Plus, Trash2, Settings } from "lucide-react";
import { Sheet, SheetContent, SheetHeader, SheetTitle, SheetDescription } from "@/components/ui/sheet";

interface Middleware {
    id: string;
    name: string;
    type: string;
    enabled: boolean;
}

const INITIAL_MIDDLEWARE: Middleware[] = [
    { id: "mw-1", name: "Authentication", type: "auth", enabled: true },
    { id: "mw-2", name: "Rate Limiter", type: "rate_limit", enabled: true },
    { id: "mw-3", name: "Logging", type: "logger", enabled: true },
    { id: "mw-4", name: "Request Validation", type: "validator", enabled: false },
];

export default function MiddlewarePage() {
    const [middleware, setMiddleware] = useState<Middleware[]>(INITIAL_MIDDLEWARE);
    const [selectedMiddleware, setSelectedMiddleware] = useState<Middleware | null>(null);

    const onDragEnd = (result: DropResult) => {
        if (!result.destination) {
            return;
        }

        const items = Array.from(middleware);
        const [reorderedItem] = items.splice(result.source.index, 1);
        items.splice(result.destination.index, 0, reorderedItem);

        setMiddleware(items);
    };

    const toggleMiddleware = (id: string) => {
        setMiddleware(middleware.map(m => m.id === id ? { ...m, enabled: !m.enabled } : m));
    };

    return (
        <div className="flex-1 space-y-4 p-8 pt-6">
            <div className="flex items-center justify-between">
                <div>
                    <h2 className="text-3xl font-bold tracking-tight">Middleware Pipeline</h2>
                    <p className="text-muted-foreground">Drag and drop to reorder the request processing pipeline.</p>
                </div>
                <Button>
                    <Plus className="mr-2 h-4 w-4" /> Add Middleware
                </Button>
            </div>

            <div className="grid gap-6 md:grid-cols-2">
                <Card className="backdrop-blur-sm bg-background/50 h-fit">
                    <CardHeader>
                        <CardTitle>Active Pipeline</CardTitle>
                        <CardDescription>
                            Requests flow from top to bottom.
                        </CardDescription>
                    </CardHeader>
                    <CardContent>
                        <DragDropContext onDragEnd={onDragEnd}>
                            <Droppable droppableId="middleware-list">
                                {(provided) => (
                                    <div
                                        {...provided.droppableProps}
                                        ref={provided.innerRef}
                                        className="space-y-3"
                                    >
                                        {middleware.map((item, index) => (
                                            <Draggable key={item.id} draggableId={item.id} index={index}>
                                                {(provided, snapshot) => (
                                                    <div
                                                        ref={provided.innerRef}
                                                        {...provided.draggableProps}
                                                        className={`flex items-center justify-between p-4 rounded-lg border bg-card text-card-foreground shadow-sm transition-all ${snapshot.isDragging ? "shadow-lg ring-2 ring-primary ring-opacity-50" : "hover:bg-accent hover:text-accent-foreground"}`}
                                                    >
                                                        <div className="flex items-center gap-3">
                                                            <div {...provided.dragHandleProps} className="cursor-grab active:cursor-grabbing text-muted-foreground hover:text-foreground">
                                                                <GripVertical className="h-5 w-5" />
                                                            </div>
                                                            <div>
                                                                <p className="font-medium">{item.name}</p>
                                                                <p className="text-xs text-muted-foreground font-mono">{item.type}</p>
                                                            </div>
                                                        </div>
                                                        <div className="flex items-center gap-4">
                                                            <Switch
                                                                checked={item.enabled}
                                                                onCheckedChange={() => toggleMiddleware(item.id)}
                                                            />
                                                            <Button variant="ghost" size="icon" onClick={() => setSelectedMiddleware(item)}>
                                                                <Settings className="h-4 w-4" />
                                                            </Button>
                                                        </div>
                                                    </div>
                                                )}
                                            </Draggable>
                                        ))}
                                        {provided.placeholder}
                                    </div>
                                )}
                            </Droppable>
                        </DragDropContext>
                    </CardContent>
                </Card>

                <div className="space-y-6">
                    <Card>
                        <CardHeader>
                            <CardTitle>Pipeline Visualization</CardTitle>
                        </CardHeader>
                        <CardContent>
                            <div className="flex flex-col items-center space-y-2 py-4">
                                <div className="px-4 py-2 bg-muted rounded-full text-xs font-mono text-muted-foreground">Incoming Request</div>
                                <div className="h-6 w-0.5 bg-border"></div>
                                {middleware.filter(m => m.enabled).map((m, i) => (
                                    <div key={m.id} className="flex flex-col items-center w-full">
                                        <div className="w-3/4 p-3 border rounded-md text-center bg-card shadow-sm text-sm">
                                            {m.name}
                                        </div>
                                        {i < middleware.filter(x => x.enabled).length - 1 && (
                                             <div className="h-6 w-0.5 bg-border"></div>
                                        )}
                                    </div>
                                ))}
                                {middleware.filter(m => m.enabled).length === 0 && (
                                    <div className="p-4 border border-dashed rounded-md text-muted-foreground text-sm">No active middleware</div>
                                )}
                                <div className="h-6 w-0.5 bg-border"></div>
                                <div className="px-4 py-2 bg-primary text-primary-foreground rounded-full text-xs font-mono">Service</div>
                            </div>
                        </CardContent>
                    </Card>
                </div>
            </div>

            <Sheet open={!!selectedMiddleware} onOpenChange={(open) => !open && setSelectedMiddleware(null)}>
                <SheetContent>
                    <SheetHeader>
                        <SheetTitle>Configure Middleware</SheetTitle>
                        <SheetDescription>
                            Configure settings for {selectedMiddleware?.name}
                        </SheetDescription>
                    </SheetHeader>
                    <div className="grid gap-4 py-4">
                        <div className="text-sm text-muted-foreground">
                            Configuration options for {selectedMiddleware?.type} will appear here.
                        </div>
                    </div>
                </SheetContent>
            </Sheet>
        </div>
    );
}
