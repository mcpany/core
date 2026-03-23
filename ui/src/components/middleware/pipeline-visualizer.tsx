/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */



import { useState, useEffect } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Save, Loader2, GripVertical } from "lucide-react";
import { toast } from "sonner";
import { Switch } from "@/components/ui/switch";
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

    const onDragEnd = (result: DropResult) => {
        if (!result.destination) return;

        const sourceIndex = result.source.index;
        const destinationIndex = result.destination.index;

        if (sourceIndex === destinationIndex) return;

        const newList = Array.from(middlewares);
        const [reorderedItem] = newList.splice(sourceIndex, 1);
        newList.splice(destinationIndex, 0, reorderedItem);

        updatePriorities(newList);
    };

    const toggleDisabled = (index: number, disabled: boolean) => {
        const newList = [...middlewares];
        newList[index].disabled = disabled;
        setMiddlewares(newList);
    };

    const updatePriorities = (list: Middleware[]) => {
        // Reassign priorities: 10, 20, 30...
        const updated = list.map((m, i) => ({ ...m, priority: (i + 1) * 10 }));
        setMiddlewares(updated);
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
                    <Droppable droppableId="middleware-pipeline">
                        {(provided) => (
                            <div {...provided.droppableProps} ref={provided.innerRef} className="space-y-2">
                                {middlewares.map((m, i) => (
                                    <Draggable key={m.name} draggableId={m.name} index={i}>
                                        {(provided) => (
                                            <div
                                                ref={provided.innerRef}
                                                {...provided.draggableProps}
                                                className="flex items-center justify-between p-3 border rounded-lg bg-card hover:bg-accent/50 transition-colors"
                                            >
                                                <div className="flex items-center gap-4">
                                                    <div {...provided.dragHandleProps} className="cursor-grab hover:text-accent-foreground text-muted-foreground">
                                                        <GripVertical className="h-5 w-5" />
                                                    </div>
                                                    <Badge variant="outline" className="w-8 h-8 flex items-center justify-center rounded-full">
                                                        {i + 1}
                                                    </Badge>
                                                    <div>
                                                        <div className="font-medium">{m.name}</div>
                                                        <div className="text-xs text-muted-foreground">Priority: {m.priority}</div>
                                                    </div>
                                                    {m.disabled && <Badge variant="destructive">Disabled</Badge>}
                                                </div>
                                                <div className="flex gap-1 items-center">
                                                    <Switch
                                                        checked={!m.disabled}
                                                        onCheckedChange={(checked) => toggleDisabled(i, !checked)}
                                                    />
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
