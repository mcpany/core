/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useEffect } from "react";
import { DragDropContext, Droppable, Draggable, DropResult } from "@hello-pangea/dnd";
import { MetricsOverview } from "@/components/dashboard/metrics-overview";
import { ServiceHealthWidget } from "@/components/dashboard/service-health-widget";
import { LazyRequestVolumeChart, LazyTopToolsWidget, LazyHealthHistoryChart, LazyRecentActivityWidget } from "@/components/dashboard/lazy-charts";
import { ToolFailureRateWidget } from "@/components/dashboard/tool-failure-rate-widget";
import { GripVertical, MoreHorizontal, Maximize, Columns, LayoutGrid, EyeOff, Settings2 } from "lucide-react";
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

type WidgetType = "full" | "half" | "third" | "two-thirds";

interface DashboardWidget {
    id: string;
    title: string;
    type: WidgetType;
    hidden?: boolean;
}

const DEFAULT_WIDGETS: DashboardWidget[] = [
    { id: "metrics", title: "Metrics Overview", type: "full" },
    { id: "recent-activity", title: "Recent Activity", type: "half" },
    { id: "uptime", title: "System Uptime", type: "half" },
    { id: "failure-rate", title: "Tool Failure Rates", type: "third" },
    { id: "request-volume", title: "Request Volume", type: "half" },
    { id: "top-tools", title: "Top Tools", type: "third" },
    { id: "service-health", title: "Service Health", type: "third" },
];

/**
 * DashboardGrid component.
 * Implements a draggable grid for dashboard widgets with resizing and visibility controls.
 * @returns The rendered component.
 */
export function DashboardGrid() {
    const [widgets, setWidgets] = useState<DashboardWidget[]>(DEFAULT_WIDGETS);
    const [isMounted, setIsMounted] = useState(false);

    useEffect(() => {
        setIsMounted(true);
        const saved = localStorage.getItem("dashboard-layout");
        if (saved) {
            try {
                const parsed = JSON.parse(saved);
                // Migration logic: Ensure all widgets have valid types and hidden property
                const migrated = parsed.map((w: any) => ({
                    ...w,
                    type: w.type === "wide" ? "full" : (["full", "half", "third", "two-thirds"].includes(w.type) ? w.type : "third"),
                    hidden: w.hidden ?? false,
                    // Fix stale titles: Find the current title from DEFAULT_WIDGETS
                    title: DEFAULT_WIDGETS.find(d => d.id === w.id)?.title || w.title
                }));

                // Merge with default widgets to catch any new ones added to the codebase
                const currentIds = new Set(migrated.map((w: any) => w.id));
                const missingWidgets = DEFAULT_WIDGETS.filter(w => !currentIds.has(w.id));

                setWidgets([...migrated, ...missingWidgets]);
            } catch (e) {
                console.error("Failed to load dashboard layout", e);
                setWidgets(DEFAULT_WIDGETS);
            }
        }
    }, []);

    const saveWidgets = (newWidgets: DashboardWidget[]) => {
        setWidgets(newWidgets);
        localStorage.setItem("dashboard-layout", JSON.stringify(newWidgets));
    };

    const onDragEnd = (result: DropResult) => {
        if (!result.destination) return;

        // We are dragging within the filtered list (visible widgets only)
        // But we need to update the full list order.
        // Actually, reordering hidden widgets doesn't make much sense, so we effectively reorder the visible subset
        // and reconstruct the full list.

        const visibleWidgets = widgets.filter(w => !w.hidden);
        const hiddenWidgets = widgets.filter(w => w.hidden);

        const items = Array.from(visibleWidgets);
        const [reorderedItem] = items.splice(result.source.index, 1);
        items.splice(result.destination.index, 0, reorderedItem);

        // Ideally we keep hidden widgets at the end or intermixed?
        // For simplicity, let's append hidden widgets at the end to avoid complex splicing
        saveWidgets([...items, ...hiddenWidgets]);
    };

    const updateWidgetType = (id: string, newType: WidgetType) => {
        const updated = widgets.map(w => w.id === id ? { ...w, type: newType } : w);
        saveWidgets(updated);
    };

    const toggleWidgetVisibility = (id: string) => {
        const updated = widgets.map(w => w.id === id ? { ...w, hidden: !w.hidden } : w);
        saveWidgets(updated);
    };

    if (!isMounted) return null;

    const renderWidget = (id: string) => {
        switch (id) {
            case "metrics": return <MetricsOverview />;
            case "recent-activity": return <LazyRecentActivityWidget />;
            case "uptime": return <LazyHealthHistoryChart />;
            case "failure-rate": return <ToolFailureRateWidget />;
            case "request-volume": return <LazyRequestVolumeChart />;
            case "top-tools": return <LazyTopToolsWidget />;
            case "service-health": return <ServiceHealthWidget />;
            default: return null;
        }
    };

    const getColSpan = (type: WidgetType) => {
        switch (type) {
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
            <div className="flex justify-end">
                <Popover>
                    <PopoverTrigger asChild>
                        <Button variant="outline" size="sm" className="h-8 border-dashed">
                            <Settings2 className="mr-2 h-4 w-4" />
                            Customize View
                        </Button>
                    </PopoverTrigger>
                    <PopoverContent className="w-56" align="end">
                        <div className="space-y-2">
                            <h4 className="font-medium leading-none mb-2">Toggle Widgets</h4>
                            {widgets.map((widget) => (
                                <div key={widget.id} className="flex items-center space-x-2">
                                    <Checkbox
                                        id={`show-${widget.id}`}
                                        checked={!widget.hidden}
                                        onCheckedChange={() => toggleWidgetVisibility(widget.id)}
                                    />
                                    <Label htmlFor={`show-${widget.id}`} className="text-sm font-normal cursor-pointer w-full">
                                        {widget.title}
                                    </Label>
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
                                <Draggable key={widget.id} draggableId={widget.id} index={index}>
                                    {(provided, snapshot) => (
                                        <div
                                            ref={provided.innerRef}
                                            {...provided.draggableProps}
                                            className={cn(
                                                "relative group/widget rounded-lg transition-all duration-200",
                                                getColSpan(widget.type),
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
                                                                <DropdownMenuRadioGroup value={widget.type} onValueChange={(v) => updateWidgetType(widget.id, v as WidgetType)}>
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
                                                        <DropdownMenuItem onClick={() => toggleWidgetVisibility(widget.id)}>
                                                            <EyeOff className="mr-2 h-4 w-4" />
                                                            Hide Widget
                                                        </DropdownMenuItem>
                                                    </DropdownMenuContent>
                                                </DropdownMenu>
                                            </div>

                                            {renderWidget(widget.id)}
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
