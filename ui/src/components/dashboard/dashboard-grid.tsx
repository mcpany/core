/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useEffect, useRef } from "react";
import { DragDropContext, Droppable, Draggable, DropResult } from "@hello-pangea/dnd";
import { GripVertical, MoreHorizontal, Maximize, Columns, LayoutGrid, EyeOff, Trash2, Settings2, Loader2 } from "lucide-react";
import { cn } from "@/lib/utils";
import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuLabel,
    DropdownMenuSeparator,
    DropdownMenuTrigger,
    DropdownMenuSub,
    DropdownMenuSubTrigger,
    DropdownMenuSubContent,
    DropdownMenuRadioGroup,
    DropdownMenuRadioItem,
} from "@/components/ui/dropdown-menu";
import { Button } from "@/components/ui/button";
import {
    Popover,
    PopoverContent,
    PopoverTrigger,
} from "@/components/ui/popover";
import { Checkbox } from "@/components/ui/checkbox";
import { Label } from "@/components/ui/label";
import { WIDGET_DEFINITIONS, getWidgetDefinition, WidgetSize } from "@/components/dashboard/widget-registry";
import { AddWidgetSheet } from "@/components/dashboard/add-widget-sheet";
import { apiClient } from "@/lib/client";
import { useToast } from "@/hooks/use-toast";

/**
 * Represents a specific instance of a widget on the dashboard.
 */
export interface WidgetInstance {
    /** Unique ID for this instance (allows multiple widgets of same type). */
    instanceId: string;
    /** The type of widget (must match a type in WIDGET_DEFINITIONS). */
    type: string;
    /** The title to display for this instance. */
    title: string;
    /** The current size of the widget. */
    size: WidgetSize;
    /** Whether the widget is currently hidden from view. */
    hidden?: boolean;
}

// Default widgets for a fresh dashboard
const DEFAULT_LAYOUT: WidgetInstance[] = WIDGET_DEFINITIONS.map(def => ({
    instanceId: crypto.randomUUID(),
    type: def.type,
    title: def.title,
    size: def.defaultSize,
    hidden: false
}));

/**
 * DashboardGrid component.
 * Implements a draggable grid for dashboard widgets with resizing and dynamic layout controls.
 * Persists layout to backend user preferences.
 * @returns The rendered component.
 */
export function DashboardGrid() {
    const [widgets, setWidgets] = useState<WidgetInstance[]>([]);
    const [isMounted, setIsMounted] = useState(false);
    const [isLoading, setIsLoading] = useState(true);
    const { toast } = useToast();

    // Fetch layout from backend on mount
    useEffect(() => {
        setIsMounted(true);
        const fetchLayout = async () => {
            try {
                const res = await apiClient.getUserPreferences();
                const savedLayout = res.preferences?.["dashboard_layout"];

                if (savedLayout) {
                    try {
                        const parsed = JSON.parse(savedLayout);
                        // Validate format loosely
                        if (Array.isArray(parsed)) {
                            setWidgets(parsed);
                        } else {
                            setWidgets(DEFAULT_LAYOUT);
                        }
                    } catch (e) {
                        console.error("Failed to parse dashboard layout", e);
                        setWidgets(DEFAULT_LAYOUT);
                    }
                } else {
                    // Try legacy localStorage migration
                    const localSaved = localStorage.getItem("dashboard-layout");
                    if (localSaved) {
                        try {
                            const parsed = JSON.parse(localSaved);
                            // Simple validation
                            if (Array.isArray(parsed)) {
                                setWidgets(parsed);
                                // Trigger a save to backend immediately to migrate?
                                // We'll let the next effect handle it if we mark it dirty, but effect skips first run.
                                // Let's just set it and let user save on next interaction, or we could force save.
                            } else {
                                setWidgets(DEFAULT_LAYOUT);
                            }
                        } catch {
                            setWidgets(DEFAULT_LAYOUT);
                        }
                    } else {
                        setWidgets(DEFAULT_LAYOUT);
                    }
                }
            } catch (e) {
                console.error("Failed to fetch user preferences", e);
                // Fallback to defaults or localStorage if offline?
                // For "Portainer" feel, we might want to alert, but for resilience, fallback to local is kind.
                const localSaved = localStorage.getItem("dashboard-layout");
                if (localSaved) {
                    try {
                        setWidgets(JSON.parse(localSaved));
                    } catch {
                        setWidgets(DEFAULT_LAYOUT);
                    }
                } else {
                    setWidgets(DEFAULT_LAYOUT);
                }
                toast({
                    variant: "destructive",
                    title: "Sync Error",
                    description: "Could not load cloud layout. Using local/default.",
                });
            } finally {
                setIsLoading(false);
            }
        };

        fetchLayout();
    }, [toast]);

    const saveWidgets = (newWidgets: WidgetInstance[]) => {
        setWidgets(newWidgets);
    };

    // ⚡ BOLT: Debounce API writes to prevent main thread blocking and network spam during drag/resize operations
    // Randomized Selection from Top 5 High-Impact Targets
    const isFirstRun = useRef(true);
    useEffect(() => {
        if (!isMounted || isLoading) return;

        // Prevent saving the initial empty/default state if we haven't touched it.
        // The fetch effect sets widgets, which triggers this effect.
        // We want to skip the save triggered by the initial fetch.
        if (isFirstRun.current) {
            isFirstRun.current = false;
            return;
        }

        const timer = setTimeout(async () => {
            try {
                const layoutJson = JSON.stringify(widgets);
                // Save to localStorage as backup/cache
                localStorage.setItem("dashboard-layout", layoutJson);

                // Save to backend
                await apiClient.updateUserPreferences({
                    "dashboard_layout": layoutJson
                });
            } catch (e) {
                console.error("Failed to save dashboard layout", e);
                // Silent fail or toast? Silent for auto-save usually better, unless persistent failure.
            }
        }, 1000); // Increased debounce to 1s for network calls

        return () => clearTimeout(timer);
    }, [widgets, isMounted, isLoading]);

    const onDragEnd = (result: DropResult) => {
        if (!result.destination) return;

        const visibleWidgets = widgets.filter(w => !w.hidden);
        const hiddenWidgets = widgets.filter(w => w.hidden);

        const items = Array.from(visibleWidgets);
        const [reorderedItem] = items.splice(result.source.index, 1);
        items.splice(result.destination.index, 0, reorderedItem);

        saveWidgets([...items, ...hiddenWidgets]);
    };

    const updateWidgetSize = (instanceId: string, newSize: WidgetSize) => {
        const updated = widgets.map(w => w.instanceId === instanceId ? { ...w, size: newSize } : w);
        saveWidgets(updated);
    };

    const toggleWidgetVisibility = (instanceId: string) => {
        const updated = widgets.map(w => w.instanceId === instanceId ? { ...w, hidden: !w.hidden } : w);
        saveWidgets(updated);
    };

    const removeWidget = (instanceId: string) => {
        const updated = widgets.filter(w => w.instanceId !== instanceId);
        saveWidgets(updated);
    };

    const addWidget = (type: string) => {
        const def = getWidgetDefinition(type);
        if (!def) return;

        const newWidget: WidgetInstance = {
            instanceId: crypto.randomUUID(),
            type: def.type,
            title: def.title,
            size: def.defaultSize,
            hidden: false
        };

        // Add to the top
        saveWidgets([newWidget, ...widgets]);
    };

    if (!isMounted) return null;

    if (isLoading && widgets.length === 0) {
        return (
            <div className="flex items-center justify-center h-64">
                <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
                <span className="ml-2 text-muted-foreground">Loading dashboard...</span>
            </div>
        );
    }

    const renderWidget = (widget: WidgetInstance) => {
        const def = getWidgetDefinition(widget.type);
        if (!def) return <div className="p-4 border border-dashed text-muted-foreground">Unknown Widget Type: {widget.type}</div>;

        const Component = def.component;
        return <Component />;
    };

    const getColSpan = (size: WidgetSize) => {
        switch (size) {
            case "full": return "col-span-12";
            case "two-thirds": return "col-span-12 lg:col-span-8";
            case "half": return "col-span-12 lg:col-span-6";
            case "third": return "col-span-12 lg:col-span-4";
            default: return "col-span-12 lg:col-span-4";
        }
    };

    const visibleWidgets = widgets.filter(w => !w.hidden);

    return (
        <div className="space-y-4">
            <div className="flex justify-end gap-2">
                <AddWidgetSheet onAdd={addWidget} />

                <Popover>
                    <PopoverTrigger asChild>
                        <Button variant="outline" size="sm" className="h-8 border-dashed">
                            <Settings2 className="mr-2 h-4 w-4" />
                            Layout
                        </Button>
                    </PopoverTrigger>
                    <PopoverContent className="w-56" align="end">
                        <div className="space-y-2">
                            <h4 className="font-medium leading-none mb-2">Visible Widgets</h4>
                            {widgets.map((widget) => (
                                <div key={widget.instanceId} className="flex items-center space-x-2">
                                    <Checkbox
                                        id={`show-${widget.instanceId}`}
                                        checked={!widget.hidden}
                                        onCheckedChange={() => toggleWidgetVisibility(widget.instanceId)}
                                    />
                                    <Label htmlFor={`show-${widget.instanceId}`} className="text-sm font-normal cursor-pointer w-full truncate">
                                        {widget.title}
                                    </Label>
                                </div>
                            ))}
                             {widgets.length === 0 && <p className="text-xs text-muted-foreground">No widgets added.</p>}
                             <div className="pt-2">
                                <Button variant="ghost" size="sm" className="w-full text-xs text-destructive" onClick={() => saveWidgets([])}>
                                    Clear All
                                </Button>
                             </div>
                        </div>
                    </PopoverContent>
                </Popover>
            </div>

            <DragDropContext onDragEnd={onDragEnd}>
                <Droppable droppableId="dashboard-widgets" direction="horizontal">
                    {(provided) => (
                        <div
                            {...provided.droppableProps}
                            ref={provided.innerRef}
                            className="grid grid-cols-12 gap-4"
                        >
                            {visibleWidgets.map((widget, index) => (
                                <Draggable key={widget.instanceId} draggableId={widget.instanceId} index={index}>
                                    {(provided, snapshot) => (
                                        <div
                                            ref={provided.innerRef}
                                            {...provided.draggableProps}
                                            className={cn(
                                                "relative group/widget rounded-lg transition-all duration-200",
                                                getColSpan(widget.size),
                                                snapshot.isDragging && "z-50 shadow-2xl scale-[1.02] opacity-90"
                                            )}
                                        >
                                            <div className="absolute top-2 right-2 flex items-center space-x-1 opacity-0 group-hover/widget:opacity-100 transition-opacity z-20">
                                                 <div
                                                    {...provided.dragHandleProps}
                                                    className="p-1 hover:bg-muted/80 bg-background/50 backdrop-blur-sm rounded cursor-grab active:cursor-grabbing border border-transparent hover:border-border"
                                                >
                                                    <GripVertical className="h-4 w-4 text-muted-foreground" />
                                                </div>

                                                <DropdownMenu>
                                                    <DropdownMenuTrigger asChild>
                                                        <div className="p-1 hover:bg-muted/80 bg-background/50 backdrop-blur-sm rounded cursor-pointer border border-transparent hover:border-border">
                                                            <MoreHorizontal className="h-4 w-4 text-muted-foreground" />
                                                        </div>
                                                    </DropdownMenuTrigger>
                                                    <DropdownMenuContent align="end">
                                                        <DropdownMenuLabel>Widget Options</DropdownMenuLabel>
                                                        <DropdownMenuSeparator />
                                                        <DropdownMenuSub>
                                                            <DropdownMenuSubTrigger>
                                                                <Maximize className="mr-2 h-4 w-4" />
                                                                <span>Size</span>
                                                            </DropdownMenuSubTrigger>
                                                            <DropdownMenuSubContent>
                                                                <DropdownMenuRadioGroup value={widget.size} onValueChange={(v) => updateWidgetSize(widget.instanceId, v as WidgetSize)}>
                                                                    <DropdownMenuRadioItem value="full">
                                                                        <LayoutGrid className="mr-2 h-4 w-4" /> Full Width
                                                                    </DropdownMenuRadioItem>
                                                                    <DropdownMenuRadioItem value="two-thirds">
                                                                        <Columns className="mr-2 h-4 w-4" /> 2/3 Width
                                                                    </DropdownMenuRadioItem>
                                                                    <DropdownMenuRadioItem value="half">
                                                                        <Columns className="mr-2 h-4 w-4" /> 1/2 Width
                                                                    </DropdownMenuRadioItem>
                                                                    <DropdownMenuRadioItem value="third">
                                                                        <Columns className="mr-2 h-4 w-4" /> 1/3 Width
                                                                    </DropdownMenuRadioItem>
                                                                </DropdownMenuRadioGroup>
                                                            </DropdownMenuSubContent>
                                                        </DropdownMenuSub>
                                                        <DropdownMenuItem onClick={() => toggleWidgetVisibility(widget.instanceId)}>
                                                            <EyeOff className="mr-2 h-4 w-4" />
                                                            Hide Widget
                                                        </DropdownMenuItem>
                                                        <DropdownMenuSeparator />
                                                        <DropdownMenuItem onClick={() => removeWidget(widget.instanceId)} className="text-red-600 focus:text-red-600">
                                                            <Trash2 className="mr-2 h-4 w-4" />
                                                            Remove
                                                        </DropdownMenuItem>
                                                    </DropdownMenuContent>
                                                </DropdownMenu>
                                            </div>

                                            {renderWidget(widget)}
                                        </div>
                                    )}
                                </Draggable>
                            ))}
                            {provided.placeholder}
                        </div>
                    )}
                </Droppable>
            </DragDropContext>

            {/* Empty State / Onboarding */}
            {visibleWidgets.length === 0 && (
                <div className="flex flex-col items-center justify-center py-20 border-2 border-dashed rounded-lg bg-muted/20">
                    <LayoutGrid className="h-10 w-10 text-muted-foreground mb-4 opacity-50" />
                    <h3 className="text-lg font-medium">Your dashboard is empty</h3>
                    <p className="text-sm text-muted-foreground mb-4">Add widgets to customize your view.</p>
                    <AddWidgetSheet onAdd={addWidget} />
                </div>
            )}
        </div>
    );
}
