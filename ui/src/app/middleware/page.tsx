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
import { apiClient } from "@/lib/client";
import { toast } from "sonner";

interface Middleware {
    id: string;
    name: string;
    type: string;
    enabled: boolean;
    priority?: number;
}

/**
 * MiddlewarePage component.
 * @returns The rendered component.
 */
export default function MiddlewarePage() {
    const [middleware, setMiddleware] = useState<Middleware[]>([]);
    const [isLoading, setIsLoading] = useState(true);
    const [isSaving, setIsSaving] = useState(false);

    useEffect(() => {
        fetchMiddleware();
    }, []);

    const fetchMiddleware = async () => {
        setIsLoading(true);
        try {
            const settings = await apiClient.getGlobalSettings();
            if (settings.middlewares) {
                const mws: Middleware[] = settings.middlewares.map((m: any) => ({
                    id: m.name,
                    name: m.name,
                    type: m.name.toLowerCase().replace(/ /g, "_"),
                    enabled: !m.disabled,
                    priority: m.priority || 0
                }));
                mws.sort((a, b) => (a.priority || 0) - (b.priority || 0));
                setMiddleware(mws);
            } else {
                setMiddleware([]);
            }
        } catch (error) {
            console.error("Failed to load middleware", error);
            toast.error("Failed to load middleware configuration");
        } finally {
            setIsLoading(false);
        }
    };

    const handleSave = async () => {
        setIsSaving(true);
        try {
            // Fetch latest settings to avoid overwriting other fields
            const currentSettings = await apiClient.getGlobalSettings();

            // Update middlewares with new order (priority) and status
            const updatedMiddlewares = middleware.map((m, index) => ({
                name: m.name,
                priority: index, // Priority is based on list order
                disabled: !m.enabled
            }));

            currentSettings.middlewares = updatedMiddlewares;

            await apiClient.saveGlobalSettings(currentSettings);
            toast.success("Middleware pipeline saved successfully");

            // Refresh to ensure sync
            fetchMiddleware();
        } catch (error) {
            console.error("Failed to save middleware", error);
            toast.error("Failed to save middleware configuration");
        } finally {
            setIsSaving(false);
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
    };

    const toggleMiddleware = (id: string) => {
        setMiddleware(middleware.map(m => m.id === id ? { ...m, enabled: !m.enabled } : m));
    };

    return (
        <div className="flex-1 space-y-4 p-8 pt-6">
            <div className="flex items-center justify-between">
                <div>
                    <h1 className="text-3xl font-bold tracking-tight">Middleware Pipeline</h1>
                    <p className="text-muted-foreground">Drag and drop to reorder the request processing pipeline.</p>
                </div>
                <div className="flex items-center gap-2">
                    <Button variant="outline" onClick={fetchMiddleware} disabled={isLoading || isSaving}>
                        <RefreshCw className={`mr-2 h-4 w-4 ${isLoading ? 'animate-spin' : ''}`} /> Refresh
                    </Button>
                    <Button onClick={handleSave} disabled={isLoading || isSaving}>
                        <Save className="mr-2 h-4 w-4" />
                        {isSaving ? "Saving..." : "Save Changes"}
                    </Button>
                </div>
            </div>

            {isLoading ? (
                <div className="flex items-center justify-center h-64 text-muted-foreground">
                    Loading pipeline configuration...
                </div>
            ) : (
                <div className="grid gap-6 md:grid-cols-2">
                    <Card className="backdrop-blur-sm bg-background/50 h-fit">
                        <CardHeader>
                            <CardTitle>Active Pipeline</CardTitle>
                            <CardDescription>
                                Requests flow from top to bottom. (Lower priority first)
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
                            {middleware.length === 0 && (
                                <div className="p-8 text-center text-muted-foreground border border-dashed rounded-lg">
                                    No middleware configured.
                                </div>
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
            )}
        </div>
    );
}
