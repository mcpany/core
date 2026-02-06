/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useEffect, useRef } from "react";
import { DragDropContext, Droppable, Draggable, DropResult } from "@hello-pangea/dnd";
import { GripVertical, MoreHorizontal, Maximize, Columns, LayoutGrid, EyeOff, Trash2, Settings2 } from "lucide-react";
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
 * @returns The rendered component.
 */
export function DashboardGrid() {
    const [widgets, setWidgets] = useState<WidgetInstance[]>([]);
    const [isMounted, setIsMounted] = useState(false);

    useEffect(() => {
        setIsMounted(true);
        const saved = localStorage.getItem("dashboard-layout");
        if (saved) {
            try {
                const parsed = JSON.parse(saved);

                // Migration Logic
                // Case 1: Legacy format (DashboardWidget[]) where id matches type
                if (parsed.length > 0 && !parsed[0].instanceId) {
                    interface LegacyWidget {
                        id: string;
                        title: string;
                        type: string; // Actually 'wide'|'half' etc in some cases, but mapped
                        hidden?: boolean;
                    }
                    const migrated: WidgetInstance[] = parsed.map((w: LegacyWidget) => ({
                        instanceId: crypto.randomUUID(),
                        type: w.id, // In legacy, id was effectively the type
                        title: WIDGET_DEFINITIONS.find(d => d.type === w.id)?.title || w.title,
                        size: (["full", "half", "third", "two-thirds"].includes(w.type) ? w.type : "third") as WidgetSize,
                        hidden: w.hidden ?? false
                    }));

                    // Filter out any invalid types
                    const validMigrated = migrated.filter(w => getWidgetDefinition(w.type));

                    // If migration resulted in empty or too few widgets, append defaults?
                    // No, respect user's (possibly empty) layout, but ensure at least we tried.
                    if (validMigrated.length === 0) {
                        setWidgets(DEFAULT_LAYOUT);
                    } else {
                        setWidgets(validMigrated);
                    }
                } else {
                    // Case 2: Already in new format
                    setWidgets(parsed);
                }
            } catch (e) {
                console.error("Failed to load dashboard layout", e);
                setWidgets(DEFAULT_LAYOUT);
            }
        } else {
            setWidgets(DEFAULT_LAYOUT);
        }
    }, []);

    const saveWidgets = (newWidgets: WidgetInstance[]) => {
        setWidgets(newWidgets);
    };

    // ⚡ BOLT: Debounce localStorage writes to prevent main thread blocking during drag/resize operations
    // Randomized Selection from Top 5 High-Impact Targets
    const isFirstRun = useRef(true);
    useEffect(() => {
        if (!isMounted) return;

        // Prevent saving the initial empty state if it's the very first mounted render
        // But we must allow saving if we just loaded/migrated data.
        // The issue is `isMounted` flips to true, and `widgets` might update in the same cycle or next.
        // If we simply rely on `widgets.length > 0`, we might miss a user clearing all widgets.
        // But for initial load, widgets is [].

        // Simplified approach: Just check if we have widgets or if we've passed the first "real" update.
        if (isFirstRun.current) {
            isFirstRun.current = false;
            // If widgets are empty on first run, it's likely the initial state.
            // If widgets are NOT empty on first run (e.g. migration happened fast?), we might want to save?
            // But `isMounted` gate likely delays this enough.
            return;
        }

        const timer = setTimeout(() => {
            localStorage.setItem("dashboard-layout", JSON.stringify(widgets));
        }, 500);

        return () => clearTimeout(timer);
    }, [widgets, isMounted]);

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

                {/* Legacy "Customize View" popover for quickly toggling hidden widgets could remain,
                    but "Add Widget" is cleaner. Let's keep a "View Options" for hidden widgets restoration?
                    Actually, if we support DELETE, hidden widgets are less useful unless it's a temp hide.
                    Let's keep the popover for recovering hidden widgets.
                */}
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
                <div className="flex flex-col items-center justify-center py-20 border-2 border-dashed rounded-lg bg-card/50 backdrop-blur-sm shadow-sm">
                    <div className="p-4 bg-primary/10 rounded-full mb-4">
                        <LayoutGrid className="h-10 w-10 text-primary" />
                    </div>
                    <h3 className="text-xl font-semibold mb-2">Your dashboard is empty</h3>
                    <p className="text-muted-foreground mb-6 max-w-sm text-center">
                        Add widgets to visualize metrics, monitor services, and manage your MCP ecosystem.
                    </p>
                    <AddWidgetSheet onAdd={addWidget} />
                </div>
            )}
        </div>
    );
}
