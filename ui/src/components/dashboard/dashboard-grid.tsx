/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useEffect, useRef, useMemo } from "react";
// @ts-expect-error - WidthProvider is missing from main export in v2, using legacy which includes it
import { Responsive, WidthProvider } from "react-grid-layout/legacy";
import "react-grid-layout/css/styles.css";
import "react-resizable/css/styles.css";

import { MoreHorizontal, Maximize, Columns, LayoutGrid, EyeOff, Trash2, Settings2, Loader2 } from "lucide-react";
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

const ResponsiveGridLayout = WidthProvider(Responsive);

/**
 * Represents the layout of a widget in the grid.
 */
export interface Layout {
    x: number;
    y: number;
    w: number;
    h: number;
}

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
    /** The layout of the widget in the grid. */
    layout: Layout;
    /** Whether the widget is currently hidden from view. */
    hidden?: boolean;
}

// Default widgets for a fresh dashboard
const DEFAULT_LAYOUT: WidgetInstance[] = WIDGET_DEFINITIONS.map((def, index) => ({
    instanceId: crypto.randomUUID(),
    type: def.type,
    title: def.title,
    size: def.defaultSize,
    layout: { x: (index % 3) * 4, y: Math.floor(index / 3) * 4, w: 4, h: 4 }, // Basic default layout
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
    const [loading, setLoading] = useState(true);

    const migrateLayout = (parsed: any): WidgetInstance[] => {
        let widgets: Partial<WidgetInstance>[] = [];

        // Case 1: Legacy format (DashboardWidget[]) where id matches type
        if (Array.isArray(parsed) && parsed.length > 0 && !parsed[0].instanceId) {
            interface LegacyWidget {
                id: string;
                title: string;
                type: string;
                hidden?: boolean;
            }
            widgets = parsed.map((w: LegacyWidget) => ({
                instanceId: crypto.randomUUID(),
                type: w.id,
                title: WIDGET_DEFINITIONS.find(d => d.type === w.id)?.title || w.title,
                size: (["full", "half", "third", "two-thirds"].includes(w.type) ? w.type : "third") as WidgetSize,
                hidden: w.hidden ?? false
            }));
        } else if (Array.isArray(parsed)) {
            // Case 2: Already in intermediate format (might lack layout)
            widgets = parsed;
        } else {
            return DEFAULT_LAYOUT;
        }

        // Assign Layout if missing
        let currentY = 0;
        let currentX = 0;
        const ROW_HEIGHT = 4;

        const getSizeW = (size: WidgetSize) => {
            switch (size) {
                case "full": return 12;
                case "two-thirds": return 8;
                case "half": return 6;
                case "third": return 4;
                default: return 4;
            }
        };

        const migrated: WidgetInstance[] = widgets.map((w) => {
            // If already has layout, keep it
            if ((w as any).layout) return w as WidgetInstance;

            const width = getSizeW(w.size || "third");
            const height = w.type === "metrics" ? 3 : ROW_HEIGHT; // Metrics overview can be shorter

            // Simple flow layout
            if (currentX + width > 12) {
                currentX = 0;
                currentY += ROW_HEIGHT; // Move to next row approximation
            }

            const layout = { x: currentX, y: currentY, w: width, h: height };

            // Advance cursor
            currentX += width;

            return {
                instanceId: w.instanceId || crypto.randomUUID(),
                type: w.type || "unknown",
                title: w.title || "Widget",
                size: w.size || "third",
                layout: layout,
                hidden: w.hidden ?? false
            };
        }).filter(w => getWidgetDefinition(w.type));

        if (migrated.length === 0) return DEFAULT_LAYOUT;
        return migrated;
    }

    useEffect(() => {
        setIsMounted(true);

        const loadLayout = async () => {
            try {
                // Fetch from API
                const res = await fetch('/api/v1/user/preferences');
                if (res.ok) {
                    const data = await res.json();
                    if (data && data['dashboard-layout']) {
                         try {
                            const parsed = JSON.parse(data['dashboard-layout']);
                            setWidgets(migrateLayout(parsed));
                         } catch (e) {
                            console.error("Failed to parse remote layout", e);
                            setWidgets(DEFAULT_LAYOUT);
                         }
                    } else {
                         // No layout saved in backend, check local storage for migration
                         const local = localStorage.getItem("dashboard-layout");
                         if (local) {
                             try {
                                const parsed = JSON.parse(local);
                                const migrated = migrateLayout(parsed);
                                setWidgets(migrated);
                                // We rely on the save effect to sync this to backend
                             } catch (e) {
                                console.error("Failed to parse local layout", e);
                                setWidgets(DEFAULT_LAYOUT);
                             }
                         } else {
                             setWidgets(DEFAULT_LAYOUT);
                         }
                    }
                } else {
                     console.warn("Failed to fetch preferences, falling back to local/default");
                     // Fallback to local storage or default
                     const local = localStorage.getItem("dashboard-layout");
                     if (local) {
                        try {
                            setWidgets(migrateLayout(JSON.parse(local)));
                        } catch {
                            setWidgets(DEFAULT_LAYOUT);
                        }
                     } else {
                        setWidgets(DEFAULT_LAYOUT);
                     }
                }
            } catch (err) {
                 console.error("Failed to load layout", err);
                 setWidgets(DEFAULT_LAYOUT);
            } finally {
                setLoading(false);
            }
        };

        loadLayout();
    }, []);

    const saveWidgets = (newWidgets: WidgetInstance[]) => {
        setWidgets(newWidgets);
    };

    // ⚡ BOLT: Debounce API writes to prevent server spam during drag/resize operations
    // Randomized Selection from Top 5 High-Impact Targets
    const isFirstRun = useRef(true);
    useEffect(() => {
        if (!isMounted || loading) return;

        // Prevent saving the initial empty state if it's the very first mounted render
        // But we must allow saving if we just loaded/migrated data.
        if (isFirstRun.current) {
            isFirstRun.current = false;
            // If widgets are empty on first run, it's likely the initial state.
            return;
        }

        const timer = setTimeout(async () => {
            try {
                await fetch('/api/v1/user/preferences', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({
                        'dashboard-layout': JSON.stringify(widgets)
                    })
                });
                // Sync to local storage as backup/cache
                localStorage.setItem("dashboard-layout", JSON.stringify(widgets));
            } catch (err) {
                console.error("Failed to save layout", err);
            }
        }, 1000); // Increased debounce to 1s for network

        return () => clearTimeout(timer);
    }, [widgets, isMounted, loading]);

    const onLayoutChange = (layout: any[]) => {
        // Sync RGL layout back to widget state
        const updatedWidgets = widgets.map(w => {
            const layoutItem = layout.find((l: any) => l.i === w.instanceId);
            if (layoutItem) {
                return {
                    ...w,
                    layout: {
                        x: layoutItem.x,
                        y: layoutItem.y,
                        w: layoutItem.w,
                        h: layoutItem.h
                    }
                };
            }
            return w;
        });
        saveWidgets(updatedWidgets);
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

        // Find first available spot or just append at bottom
        // Simple logic: place at y = max_y + height, x = 0
        const maxY = widgets.reduce((acc, w) => Math.max(acc, w.layout.y + w.layout.h), 0);

        let width = 4;
        let height = 4;
        switch(def.defaultSize) {
            case 'full': width = 12; break;
            case 'two-thirds': width = 8; break;
            case 'half': width = 6; break;
            default: width = 4;
        }

        // Metrics overview can be shorter
        if (def.type === "metrics") height = 3;

        const newWidget: WidgetInstance = {
            instanceId: crypto.randomUUID(),
            type: def.type,
            title: def.title,
            size: def.defaultSize,
            layout: { x: 0, y: maxY, w: width, h: height },
            hidden: false
        };

        // Add to the top
        saveWidgets([...widgets, newWidget]);
    };

    if (!isMounted) return null;

    if (loading) {
        return (
            <div className="flex h-64 items-center justify-center">
                <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
            </div>
        );
    }

    const renderWidget = (widget: WidgetInstance) => {
        const def = getWidgetDefinition(widget.type);
        if (!def) return <div className="p-4 border border-dashed text-muted-foreground">Unknown Widget Type: {widget.type}</div>;

        const Component = def.component;
        return <Component />;
    };

    const visibleWidgets = widgets.filter(w => !w.hidden);
    const layout = visibleWidgets.map(w => ({
        i: w.instanceId,
        x: w.layout.x,
        y: w.layout.y,
        w: w.layout.w,
        h: w.layout.h
    }));

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

            <ResponsiveGridLayout
                className="layout"
                layouts={{ lg: layout }}
                breakpoints={{ lg: 1200, md: 996, sm: 768, xs: 480, xxs: 0 }}
                cols={{ lg: 12, md: 10, sm: 6, xs: 4, xxs: 2 }}
                rowHeight={60}
                onLayoutChange={(layout) => onLayoutChange(layout)}
                draggableHandle=".drag-handle"
                margin={[16, 16]}
            >
                {visibleWidgets.map((widget) => (
                    <div key={widget.instanceId} className="relative group/widget rounded-lg bg-card border shadow-sm overflow-hidden">
                        <div className="absolute top-2 right-2 flex items-center space-x-1 opacity-0 group-hover/widget:opacity-100 transition-opacity z-20">
                                <div
                                className="drag-handle p-1 hover:bg-muted/80 bg-background/50 backdrop-blur-sm rounded cursor-grab active:cursor-grabbing border border-transparent hover:border-border"
                            >
                                <MoreHorizontal className="h-4 w-4 text-muted-foreground rotate-90" />
                            </div>

                            <DropdownMenu>
                                <DropdownMenuTrigger asChild>
                                    <div className="p-1 hover:bg-muted/80 bg-background/50 backdrop-blur-sm rounded cursor-pointer border border-transparent hover:border-border">
                                        <Settings2 className="h-4 w-4 text-muted-foreground" />
                                    </div>
                                </DropdownMenuTrigger>
                                <DropdownMenuContent align="end">
                                    <DropdownMenuLabel>Widget Options</DropdownMenuLabel>
                                    <DropdownMenuSeparator />
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

                        <div className="h-full w-full overflow-hidden">
                            {renderWidget(widget)}
                        </div>
                    </div>
                ))}
            </ResponsiveGridLayout>

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
