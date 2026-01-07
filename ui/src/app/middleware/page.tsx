
"use client";

import { useState, useEffect } from "react";
import { apiClient } from "@/lib/client";
import { GlassCard } from "@/components/layout/glass-card";
import { CardHeader, CardTitle, CardDescription, CardContent } from "@/components/ui/card";
import { DragDropContext, Droppable, Draggable } from "@hello-pangea/dnd";
import { GripVertical, ToggleLeft, ToggleRight } from "lucide-react";
import { Switch } from "@/components/ui/switch";
import { cn } from "@/lib/utils";
import { useToast } from "@/hooks/use-toast";

export default function MiddlewarePage() {
    const [middlewares, setMiddlewares] = useState<any[]>([]);
    const [loading, setLoading] = useState(true);
    const { toast } = useToast();

    useEffect(() => {
        loadMiddleware();
    }, []);

    const loadMiddleware = () => {
        setLoading(true);
        apiClient.listMiddleware().then(res => {
            // Sort by priority (asc)
            const sorted = [...res].sort((a, b) => a.priority - b.priority);
            setMiddlewares(sorted);
            setLoading(false);
        }).catch(err => {
            console.error("Failed to load middleware", err);
            setLoading(false);
            toast({
                variant: "destructive",
                title: "Error",
                description: "Failed to load middleware."
            });
        });
    };

    const saveMiddleware = async (updatedMiddlewares: any[]) => {
        try {
            await apiClient.saveMiddleware(updatedMiddlewares);
            toast({
                title: "Saved",
                description: "Middleware configuration updated."
            });
        } catch (e) {
            console.error("Failed to save middleware", e);
            toast({
                variant: "destructive",
                title: "Error",
                description: "Failed to save middleware configuration."
            });
            loadMiddleware(); // Revert on error
        }
    }

    const onDragEnd = (result: any) => {
        if (!result.destination) return;

        const items = Array.from(middlewares);
        const [reorderedItem] = items.splice(result.source.index, 1);
        items.splice(result.destination.index, 0, reorderedItem);

        // Update priorities based on new index
        const updated = items.map((item, index) => ({
            ...item,
            priority: index + 1
        }));

        setMiddlewares(updated);
        saveMiddleware(updated);
    };

    const toggleMiddleware = (id: string) => {
        const updated = middlewares.map(m =>
            m.id === id ? { ...m, enabled: !m.enabled } : m
        );
        setMiddlewares(updated);
        saveMiddleware(updated);
    };

    return (
        <div className="flex-1 space-y-4 p-8 pt-6">
            <div className="flex items-center justify-between">
                <h2 className="text-3xl font-bold tracking-tight">Middleware</h2>
            </div>
            <GlassCard>
                <CardHeader>
                    <CardTitle>Request Pipeline</CardTitle>
                    <CardDescription>Drag to reorder middleware processing steps.</CardDescription>
                </CardHeader>
                <CardContent>
                    <DragDropContext onDragEnd={onDragEnd}>
                        <Droppable droppableId="middleware-list">
                            {(provided) => (
                                <div
                                    {...provided.droppableProps}
                                    ref={provided.innerRef}
                                    className="space-y-2"
                                >
                                    {middlewares.map((item, index) => (
                                        <Draggable key={item.id} draggableId={item.id} index={index}>
                                            {(provided) => (
                                                <div
                                                    ref={provided.innerRef}
                                                    {...provided.draggableProps}
                                                    className={cn(
                                                        "flex items-center justify-between p-4 rounded-lg border bg-card/50 backdrop-blur-sm transition-all hover:bg-accent/50",
                                                        !item.enabled && "opacity-60 grayscale"
                                                    )}
                                                >
                                                    <div className="flex items-center gap-4">
                                                        <div {...provided.dragHandleProps} className="cursor-grab text-muted-foreground hover:text-foreground">
                                                            <GripVertical className="h-5 w-5" />
                                                        </div>
                                                        <div>
                                                            <h4 className="font-semibold">{item.name}</h4>
                                                            <p className="text-sm text-muted-foreground">Priority: {item.priority}</p>
                                                        </div>
                                                    </div>
                                                    <div className="flex items-center gap-2">
                                                        <span className="text-sm text-muted-foreground">{item.enabled ? "Enabled" : "Disabled"}</span>
                                                        <Switch
                                                            checked={item.enabled}
                                                            onCheckedChange={() => toggleMiddleware(item.id)}
                                                        />
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
            </GlassCard>
        </div>
    );
}
