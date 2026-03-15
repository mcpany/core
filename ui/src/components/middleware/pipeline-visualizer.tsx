/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useEffect } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Switch } from "@/components/ui/switch";
import { GripVertical, Save, Loader2 } from "lucide-react";
import { toast } from "sonner";
import { DragDropContext, Droppable, Draggable, DropResult } from "@hello-pangea/dnd";

interface Middleware {
    name: string;
    priority: number;
    disabled?: boolean;
}

interface GlobalSettings {
    middlewares: Middleware[];
    [key: string]: any;
}

/** PipelineVisualizer allows verifying and modifying the middleware pipeline order. */
export function PipelineVisualizer() {
    const [middlewares, setMiddlewares] = useState<Middleware[]>([]);
    const [settings, setSettings] = useState<GlobalSettings | null>(null);
    const [loading, setLoading] = useState(true);

    const fetchSettings = async () => {
        try {
            const res = await fetch("/api/v1/settings");
            if (res.ok) {
                const data = await res.json();
                setSettings(data);
                // Sort by priority
                const sorted = (data.middlewares || []).sort((a: Middleware, b: Middleware) => a.priority - b.priority);
                setMiddlewares(sorted);
            } else {
                toast.error("Failed to load settings");
            }
        } catch (e) {
            toast.error("Failed to load settings");
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        fetchSettings();
    }, []);

    const updatePriorities = (list: Middleware[]) => {
        // Reassign priorities: 10, 20, 30...
        const updated = list.map((m, i) => ({ ...m, priority: (i + 1) * 10 }));
        setMiddlewares(updated);
    };

    const onDragEnd = (result: DropResult) => {
        if (!result.destination) return;
        const newList = Array.from(middlewares);
        const [reorderedItem] = newList.splice(result.source.index, 1);
        newList.splice(result.destination.index, 0, reorderedItem);
        updatePriorities(newList);
    };

    const toggleDisabled = (index: number) => {
        const newList = [...middlewares];
        newList[index] = { ...newList[index], disabled: !newList[index].disabled };
        setMiddlewares(newList);
    };

    const saveOrder = async () => {
        if (!settings) return;

        const loadingToast = toast.loading("Saving pipeline...");
        try {
            const newSettings = { ...settings, middlewares };
            const res = await fetch("/api/v1/settings", {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify(newSettings)
            });

            toast.dismiss(loadingToast);
            if (res.ok) {
                toast.success("Pipeline updated");
            } else {
                toast.error("Failed to save pipeline");
            }
        } catch (e) {
            toast.dismiss(loadingToast);
            toast.error("Error saving pipeline");
        }
    };

    if (loading) return <Loader2 className="animate-spin" />;

    return (
        <Card>
            <CardHeader className="flex flex-row items-center justify-between">
                <CardTitle>Processing Order</CardTitle>
                <Button onClick={saveOrder}><Save className="mr-2 h-4 w-4"/> Save Changes</Button>
            </CardHeader>
            <CardContent>
                <DragDropContext onDragEnd={onDragEnd}>
                    <Droppable droppableId="middlewares">
                        {(provided) => (
                            <div
                                {...provided.droppableProps}
                                ref={provided.innerRef}
                                className="space-y-2"
                            >
                                {middlewares.map((m, i) => (
                                    <Draggable key={m.name} draggableId={m.name} index={i}>
                                        {(provided) => (
                                            <div
                                                ref={provided.innerRef}
                                                {...provided.draggableProps}
                                                className={`flex items-center justify-between p-3 border rounded-lg transition-colors ${m.disabled ? 'bg-muted/50 opacity-60' : 'bg-card hover:bg-accent/50'}`}
                                            >
                                                <div className="flex items-center gap-4">
                                                    <div {...provided.dragHandleProps} className="cursor-grab hover:text-primary">
                                                        <GripVertical className="h-5 w-5 text-muted-foreground" />
                                                    </div>
                                                    <Badge variant="outline" className="w-8 h-8 flex items-center justify-center rounded-full">
                                                        {i + 1}
                                                    </Badge>
                                                    <div>
                                                        <div className="font-medium">{m.name}</div>
                                                        <div className="text-xs text-muted-foreground">Priority: {m.priority}</div>
                                                    </div>
                                                </div>
                                                <div className="flex items-center gap-2">
                                                    <Switch
                                                        checked={!m.disabled}
                                                        onCheckedChange={() => toggleDisabled(i)}
                                                    />
                                                    <span className="text-sm text-muted-foreground w-12 text-right">
                                                        {m.disabled ? "Off" : "On"}
                                                    </span>
                                                </div>
                                            </div>
                                        )}
                                    </Draggable>
                                ))}
                                {provided.placeholder}
                                {middlewares.length === 0 && <div className="text-center p-4 text-muted-foreground">No middlewares configured.</div>}
                            </div>
                        )}
                    </Droppable>
                </DragDropContext>
            </CardContent>
        </Card>
    );
}
