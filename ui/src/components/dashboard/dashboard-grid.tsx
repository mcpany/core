/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useEffect } from "react";
import { DragDropContext, Droppable, Draggable, DropResult } from "@hello-pangea/dnd";
import { MetricsOverview } from "@/components/dashboard/metrics-overview";
import { ServiceHealthWidget } from "@/components/dashboard/service-health-widget";
import { LazyRequestVolumeChart, LazyTopToolsWidget, LazyHealthHistoryChart } from "@/components/dashboard/lazy-charts";
import { ToolFailureRateWidget } from "@/components/dashboard/tool-failure-rate-widget";
import { GripVertical } from "lucide-react";
import { cn } from "@/lib/utils";

const DEFAULT_WIDGETS = [
    { id: "metrics", title: "Metrics Overview", type: "wide" },
    { id: "uptime", title: "System Uptime", type: "half" }, // 4 cols
    { id: "failure-rate", title: "Tool Failure Rates", type: "third" }, // 3 cols
    { id: "request-volume", title: "Request Volume", type: "half" },
    { id: "top-tools", title: "Top Tools", type: "third" },
    { id: "service-health", title: "Service Health", type: "third" },
];

/**
 * DashboardGrid component.
 * Implements a draggable grid for dashboard widgets.
 * @returns The rendered component.
 */
export function DashboardGrid() {
    const [widgets, setWidgets] = useState(DEFAULT_WIDGETS);
    const [isMounted, setIsMounted] = useState(false);

    useEffect(() => {
        setIsMounted(true);
        const saved = localStorage.getItem("dashboard-layout");
        if (saved) {
            try {
                setWidgets(JSON.parse(saved));
            } catch (e) {
                console.error("Failed to load dashboard layout", e);
            }
        }
    }, []);

    const onDragEnd = (result: DropResult) => {
        if (!result.destination) return;

        const items = Array.from(widgets);
        const [reorderedItem] = items.splice(result.source.index, 1);
        items.splice(result.destination.index, 0, reorderedItem);

        setWidgets(items);
        localStorage.setItem("dashboard-layout", JSON.stringify(items));
    };

    if (!isMounted) return null;

    const renderWidget = (id: string) => {
        switch (id) {
            case "metrics": return <MetricsOverview />;
            case "uptime": return <LazyHealthHistoryChart />;
            case "failure-rate": return <ToolFailureRateWidget />;
            case "request-volume": return <LazyRequestVolumeChart />;
            case "top-tools": return <LazyTopToolsWidget />;
            case "service-health": return <ServiceHealthWidget />;
            default: return null;
        }
    };

    return (
        <DragDropContext onDragEnd={onDragEnd}>
            <Droppable droppableId="dashboard-widgets" direction="vertical">
                {(provided) => (
                    <div
                        {...provided.droppableProps}
                        ref={provided.innerRef}
                        className="space-y-4"
                    >
                        {/* Metrics is always on top for now as it's full width and special */}
                        {widgets.filter(w => w.id === "metrics").map((widget, index) => (
                             <div key={widget.id} className="w-full">
                                {renderWidget(widget.id)}
                             </div>
                        ))}

                        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-7">
                            {widgets.filter(w => w.id !== "metrics").map((widget, index) => (
                                <Draggable key={widget.id} draggableId={widget.id} index={index}>
                                    {(provided, snapshot) => (
                                        <div
                                            ref={provided.innerRef}
                                            {...provided.draggableProps}
                                            className={cn(
                                                "relative group/widget",
                                                widget.type === "half" ? "lg:col-span-4" : "lg:col-span-3",
                                                snapshot.isDragging && "z-50 shadow-2xl scale-[1.02] transition-transform"
                                            )}
                                        >
                                            <div
                                                {...provided.dragHandleProps}
                                                className="absolute top-3 right-3 opacity-0 group-hover/widget:opacity-100 transition-opacity z-10 cursor-grab active:cursor-grabbing p-1 hover:bg-muted rounded"
                                            >
                                                <GripVertical className="h-4 w-4 text-muted-foreground" />
                                            </div>
                                            {renderWidget(widget.id)}
                                        </div>
                                    )}
                                </Draggable>
                            ))}
                        </div>
                        {provided.placeholder}
                    </div>
                )}
            </Droppable>
        </DragDropContext>
    );
}
