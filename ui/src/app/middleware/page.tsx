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
import { GripVertical, Plus, Settings, Save, RefreshCw } from "lucide-react";
import { toast } from "sonner";

interface Middleware {
    name: string;
    priority: number;
    disabled: boolean;
}

// Helper for UI representation
interface UIMiddleware extends Middleware {
    id: string; // We use name as ID
}

/**
 * MiddlewarePage component.
 * @returns The rendered component.
 */
export default function MiddlewarePage() {
    const [middleware, setMiddleware] = useState<UIMiddleware[]>([]);
    const [loading, setLoading] = useState(true);
    const [saving, setSaving] = useState(false);

    useEffect(() => {
        fetchMiddleware();
    }, []);

    const fetchMiddleware = async () => {
        setLoading(true);
        try {
            const res = await fetch('/api/middleware');
            if (!res.ok) throw new Error("Failed to fetch");
            const data: Middleware[] = await res.json();
            // Sort by priority (asc) for display (Top = runs first = lower priority)
            const sorted = data.sort((a, b) => a.priority - b.priority).map(m => ({
                ...m,
                id: m.name
            }));
            setMiddleware(sorted);
        } catch (err) {
            console.error(err);
            toast.error("Failed to load middleware");
        } finally {
            setLoading(false);
        }
    };

    const saveMiddleware = async (items: UIMiddleware[]) => {
        setSaving(true);
        try {
            // Re-assign priorities based on index
            // Index 0 = Priority 10, Index 1 = Priority 20... (spacing allows insertion)
            const payload: Middleware[] = items.map((m, index) => ({
                name: m.name,
                disabled: m.disabled,
                priority: (index + 1) * 10
            }));

            const res = await fetch('/api/middleware', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(payload)
            });

            if (!res.ok) throw new Error("Failed to save");
            toast.success("Middleware updated");
            fetchMiddleware(); // Reload to confirm
        } catch (err) {
            console.error(err);
            toast.error("Failed to save middleware");
        } finally {
            setSaving(false);
        }
    };

    const onDragEnd = (result: DropResult) => {
        if (!result.destination) {
            return;
        }

        const items = Array.from(middleware);
        const [reorderedItem] = items.splice(result.source.index, 1);
        items.splice(result.destination.index, 0, reorderedItem);

        setMiddleware(items);
        // Auto-save on drag end? Or explicit save?
        // Let's do explicit save button to avoid accidental heavy reloads
    };

    const toggleMiddleware = (name: string) => {
        const updated = middleware.map(m => m.name === name ? { ...m, disabled: !m.disabled } : m);
        setMiddleware(updated);
    };

    return (
        <div className="flex-1 space-y-4 p-8 pt-6">
            <div className="flex items-center justify-between">
                <div>
                    <h1 className="text-3xl font-bold tracking-tight">Middleware Pipeline</h1>
                    <p className="text-muted-foreground">Drag and drop to reorder the request processing pipeline.</p>
                </div>
                <div className="flex gap-2">
                    <Button variant="outline" onClick={fetchMiddleware} disabled={loading || saving}>
                        <RefreshCw className={`mr-2 h-4 w-4 ${loading ? 'animate-spin' : ''}`} /> Refresh
                    </Button>
                    <Button onClick={() => saveMiddleware(middleware)} disabled={loading || saving}>
                        <Save className="mr-2 h-4 w-4" /> Save Changes
                    </Button>
                </div>
            </div>

            <div className="grid gap-6 md:grid-cols-2">
                <Card className="backdrop-blur-sm bg-background/50 h-fit">
                    <CardHeader>
                        <CardTitle>Active Pipeline</CardTitle>
                        <CardDescription>
                            Requests flow from top to bottom (Priority Low to High).
                        </CardDescription>
                    </CardHeader>
                    <CardContent>
                        {loading && middleware.length === 0 ? (
                            <div className="flex justify-center p-4">Loading...</div>
                        ) : (
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
                                                                <p className="text-xs text-muted-foreground font-mono">Priority: {item.priority || (index + 1) * 10}</p>
                                                            </div>
                                                        </div>
                                                        <div className="flex items-center gap-4">
                                                            <Switch
                                                                checked={!item.disabled}
                                                                onCheckedChange={() => toggleMiddleware(item.name)}
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
                        )}
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
                                {middleware.filter(m => !m.disabled).map((m, i) => (
                                    <div key={m.id} className="flex flex-col items-center w-full">
                                        <div className="w-3/4 p-3 border rounded-md text-center bg-card shadow-sm text-sm">
                                            {m.name}
                                        </div>
                                        {i < middleware.filter(x => !x.disabled).length - 1 && (
                                             <div className="h-6 w-0.5 bg-border"></div>
                                        )}
                                    </div>
                                ))}
                                {middleware.filter(m => !m.disabled).length === 0 && (
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
