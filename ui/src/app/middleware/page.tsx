/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useEffect } from "react";
import { DragDropContext, Droppable, Draggable, DropResult } from "@hello-pangea/dnd";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Switch } from "@/components/ui/switch";
import { GripVertical, Plus, Trash2, Settings, Loader2 } from "lucide-react";
import { toast } from "sonner";

interface Middleware {
    id: string;
    name: string;
    type: string;
    enabled: boolean;
    priority: number;
}

/**
 * MiddlewarePage component.
 * @returns The rendered component.
 */
export default function MiddlewarePage() {
    const [middleware, setMiddleware] = useState<Middleware[]>([]);
    const [loading, setLoading] = useState(true);

    useEffect(() => {
        fetchMiddleware();
    }, []);

    const fetchMiddleware = async () => {
        try {
            const res = await fetch('/api/middleware');
            if (res.ok) {
                const data: any[] = await res.json();
                // Map backend proto to UI model
                // Backend: { name: string, priority: number, disabled: boolean }
                // Sort by priority (asc)
                data.sort((a, b) => (a.priority || 0) - (b.priority || 0));

                const mapped = data.map((m: any) => ({
                    id: m.name,
                    name: formatName(m.name),
                    type: m.name,
                    enabled: !m.disabled,
                    priority: m.priority || 0
                }));
                setMiddleware(mapped);
            } else {
                toast.error("Failed to load middleware");
            }
        } catch (err) {
            console.error(err);
            toast.error("Failed to connect to server");
        } finally {
            setLoading(false);
        }
    };

    const saveMiddleware = async (items: Middleware[]) => {
        // Map back to backend proto
        const payload = items.map((m, index) => ({
            name: m.type,
            priority: (index + 1) * 10, // Re-assign priorities based on order, spaced by 10
            disabled: !m.enabled
        }));

        try {
            const res = await fetch('/api/middleware', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(payload)
            });
            if (!res.ok) {
                throw new Error("Failed to save");
            }
            toast.success("Middleware pipeline updated");
            setMiddleware(items); // Optimistic update
        } catch (err) {
            toast.error("Failed to save changes");
            fetchMiddleware(); // Revert
        }
    };

    const formatName = (name: string) => {
        return name.split('_').map(w => w.charAt(0).toUpperCase() + w.slice(1)).join(' ');
    }

    const onDragEnd = (result: DropResult) => {
        if (!result.destination) {
            return;
        }

        const items = Array.from(middleware);
        const [reorderedItem] = items.splice(result.source.index, 1);
        items.splice(result.destination.index, 0, reorderedItem);

        // Update state immediately for UI responsiveness, then save
        setMiddleware(items);
        saveMiddleware(items);
    };

    const toggleMiddleware = (id: string) => {
        const updated = middleware.map(m => m.id === id ? { ...m, enabled: !m.enabled } : m);
        setMiddleware(updated);
        saveMiddleware(updated);
    };

    if (loading) {
        return (
            <div className="flex h-screen items-center justify-center">
                <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
            </div>
        );
    }

    return (
        <div className="flex-1 space-y-4 p-8 pt-6">
            <div className="flex items-center justify-between">
                <div>
                    <h1 className="text-3xl font-bold tracking-tight">Middleware Pipeline</h1>
                    <p className="text-muted-foreground">Drag and drop to reorder the request processing pipeline.</p>
                </div>
                {/* Adding new middleware is complex as it requires backend support for types.
                    For now, we only allow managing existing ones. */}
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
                                                            <Button variant="ghost" size="icon">
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
        </div>
    );
}
