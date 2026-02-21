/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useEffect } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { ArrowUp, ArrowDown, Save, Loader2 } from "lucide-react";
import { toast } from "sonner";

interface Middleware {
    name: string;
    priority: number;
    disabled?: boolean;
}

interface GlobalSettings {
    middlewares: Middleware[];
    [key: string]: any;
}

/**
 * PipelineVisualizer allows users to view and reorder the middleware processing pipeline.
 * @returns The rendered component.
 */
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

    const moveUp = (index: number) => {
        if (index === 0) return;
        const newList = [...middlewares];
        [newList[index - 1], newList[index]] = [newList[index], newList[index - 1]];
        updatePriorities(newList);
    };

    const moveDown = (index: number) => {
        if (index === middlewares.length - 1) return;
        const newList = [...middlewares];
        [newList[index + 1], newList[index]] = [newList[index], newList[index + 1]];
        updatePriorities(newList);
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
                <div className="space-y-2">
                    {middlewares.map((m, i) => (
                        <div key={m.name} className="flex items-center justify-between p-3 border rounded-lg bg-card hover:bg-accent/50 transition-colors">
                            <div className="flex items-center gap-4">
                                <Badge variant="outline" className="w-8 h-8 flex items-center justify-center rounded-full">
                                    {i + 1}
                                </Badge>
                                <div>
                                    <div className="font-medium">{m.name}</div>
                                    <div className="text-xs text-muted-foreground">Priority: {m.priority}</div>
                                </div>
                                {m.disabled && <Badge variant="destructive">Disabled</Badge>}
                            </div>
                            <div className="flex gap-1">
                                <Button variant="ghost" size="icon" disabled={i === 0} onClick={() => moveUp(i)}>
                                    <ArrowUp className="h-4 w-4" />
                                </Button>
                                <Button variant="ghost" size="icon" disabled={i === middlewares.length - 1} onClick={() => moveDown(i)}>
                                    <ArrowDown className="h-4 w-4" />
                                </Button>
                            </div>
                        </div>
                    ))}
                     {middlewares.length === 0 && <div className="text-center p-4 text-muted-foreground">No middlewares configured.</div>}
                </div>
            </CardContent>
        </Card>
    );
}
