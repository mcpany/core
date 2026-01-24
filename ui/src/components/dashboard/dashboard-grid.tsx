/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useEffect } from "react";
import { DragDropContext, Droppable, Draggable, DropResult } from "@hello-pangea/dnd";
import { GripVertical, MoreHorizontal, Maximize, Columns, LayoutGrid, EyeOff, Settings2, Trash2, Plus } from "lucide-react";
import { cn } from "@/lib/utils";
import { AddWidgetSheet } from "@/components/dashboard/add-widget-sheet";
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
import { WIDGET_DEFINITIONS, WidgetType } from "@/components/dashboard/widget-registry";

export interface DashboardWidgetInstance {
    instanceId: string;
    type: string;
    title: string; // Override title or fallback to definition
    size: WidgetType;
    hidden?: boolean;
}

const DEFAULT_LAYOUT: DashboardWidgetInstance[] = [
    { instanceId: "metrics-default", type: "metrics", title: "Metrics Overview", size: "full" },
    { instanceId: "recent-activity-default", type: "recent-activity", title: "Recent Activity", size: "half" },
    { instanceId: "uptime-default", type: "uptime", title: "System Uptime", size: "half" },
    { instanceId: "failure-rate-default", type: "failure-rate", title: "Tool Failure Rates", size: "third" },
    { instanceId: "request-volume-default", type: "request-volume", title: "Request Volume", size: "half" },
    { instanceId: "top-tools-default", type: "top-tools", title: "Top Tools", size: "third" },
    { instanceId: "service-health-default", type: "service-health", title: "Service Health", size: "third" },
];

/**
 * DashboardGrid component.
 * Implements a draggable grid for dashboard widgets with resizing and visibility controls.
 * @returns The rendered component.
 */
export function DashboardGrid() {
    const [widgets, setWidgets] = useState<DashboardWidgetInstance[]>(DEFAULT_LAYOUT);
    const [isMounted, setIsMounted] = useState(false);
    const [isAddWidgetOpen, setIsAddWidgetOpen] = useState(false);

    useEffect(() => {
        setIsMounted(true);
        const saved = localStorage.getItem("dashboard-layout");
        if (saved) {
            try {
                const parsed = JSON.parse(saved);

                // Migration Strategy:
                // Old schema: { id: "metrics", type: "full", ... }
                // New schema: { instanceId: "uuid", type: "metrics", size: "full", ... }

                // Check if it's the old schema (has 'id' instead of 'instanceId' or 'type' as size)
                const isOldSchema = parsed.length > 0 && ('id' in parsed[0]);

                if (isOldSchema) {
                    console.log("Migrating dashboard layout...");
                    const migrated = parsed.map((w: any) => {
                         // Map old ID to type (assuming old IDs were effectively types like 'metrics')
                         // But wait, the old IDs matched the keys in my new registry!
                         // So w.id is the type.
                         const widgetType = w.id;
                         const def = WIDGET_DEFINITIONS[widgetType];

                         return {
                             instanceId: w.id + "-migrated", // Preserve ID for stability
                             type: widgetType,
                             title: w.title || (def ? def.title : "Unknown Widget"),
                             size: w.type === "wide" ? "full" : (["full", "half", "third", "two-thirds"].includes(w.type) ? w.type : "third"),
                             hidden: w.hidden ?? false
                         };
                    }).filter((w: any) => WIDGET_DEFINITIONS[w.type]); // Filter out unknown types

                    setWidgets(migrated);
                    // We don't save immediately to avoid overwriting before verification,
                    // but on next drag/change it will save new schema.
                } else {
                    // Assume valid new schema
                    setWidgets(parsed);
                }

            } catch (e) {
                console.error("Failed to load dashboard layout", e);
                setWidgets(DEFAULT_LAYOUT);
            }
        }
    }, []);

    const saveWidgets = (newWidgets: DashboardWidgetInstance[]) => {
        setWidgets(newWidgets);
        localStorage.setItem("dashboard-layout", JSON.stringify(newWidgets));
    };

    const onDragEnd = (result: DropResult) => {
        if (!result.destination) return;

        const visibleWidgets = widgets.filter(w => !w.hidden);
        const hiddenWidgets = widgets.filter(w => w.hidden);

        const items = Array.from(visibleWidgets);
        const [reorderedItem] = items.splice(result.source.index, 1);
        items.splice(result.destination.index, 0, reorderedItem);

        saveWidgets([...items, ...hiddenWidgets]);
    };

    const updateWidgetSize = (instanceId: string, newSize: WidgetType) => {
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
    }

    // Helper exposed for parent/sibling components if needed, or we move Add button here later.
    const addWidget = (type: string) => {
         const def = WIDGET_DEFINITIONS[type];
         if (!def) return;
         const newWidget: DashboardWidgetInstance = {
             instanceId: crypto.randomUUID(),
             type: type,
             title: def.title,
             size: def.defaultSize,
             hidden: false
         };
         saveWidgets([...widgets, newWidget]);
    }

    if (!isMounted) return null;

    const renderWidget = (widget: DashboardWidgetInstance) => {
        const def = WIDGET_DEFINITIONS[widget.type];
        if (!def) return <div className="p-4 border border-dashed text-muted-foreground">Unknown Widget Type: {widget.type}</div>;
        const Component = def.component;
        return <Component instanceId={widget.instanceId} />;
    };

    const getColSpan = (size: WidgetType) => {
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
            <AddWidgetSheet
                open={isAddWidgetOpen}
                onOpenChange={setIsAddWidgetOpen}
                onAddWidget={addWidget}
            />

            <div className="flex justify-end gap-2">
                <Button onClick={() => setIsAddWidgetOpen(true)} size="sm" className="h-8 shadow-sm">
                    <Plus className="mr-2 h-4 w-4" /> Add Widget
                </Button>

                <Popover>
                    <PopoverTrigger asChild>
                        <Button variant="outline" size="sm" className="h-8 border-dashed">
                            <Settings2 className="mr-2 h-4 w-4" />
                            Customize View
                        </Button>
                    </PopoverTrigger>
                    <PopoverContent className="w-56" align="end">
                        <div className="space-y-2">
                            <h4 className="font-medium leading-none mb-2">Active Widgets</h4>
                            {widgets.map((widget) => (
                                <div key={widget.instanceId} className="flex items-center justify-between space-x-2">
                                    <div className="flex items-center space-x-2 overflow-hidden">
                                        <Checkbox
                                            id={`show-${widget.instanceId}`}
                                            checked={!widget.hidden}
                                            onCheckedChange={() => toggleWidgetVisibility(widget.instanceId)}
                                        />
                                        <Label htmlFor={`show-${widget.instanceId}`} className="text-sm font-normal cursor-pointer w-full truncate">
                                            {widget.title}
                                        </Label>
                                    </div>
                                    <Button
                                        variant="ghost"
                                        size="icon"
                                        className="h-6 w-6 text-muted-foreground hover:text-destructive"
                                        onClick={() => removeWidget(widget.instanceId)}
                                    >
                                        <Trash2 className="h-3 w-3" />
                                    </Button>
                                </div>
                            ))}
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
                                                                <DropdownMenuRadioGroup value={widget.size} onValueChange={(v) => updateWidgetSize(widget.instanceId, v as WidgetType)}>
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
                                                        <DropdownMenuItem onClick={() => removeWidget(widget.instanceId)} className="text-destructive focus:text-destructive">
                                                            <Trash2 className="mr-2 h-4 w-4" />
                                                            Remove Widget
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
        </div>
    );
}
