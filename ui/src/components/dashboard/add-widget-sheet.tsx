/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
  SheetTrigger,
} from "@/components/ui/sheet";
import { Button } from "@/components/ui/button";
import { Plus } from "lucide-react";
import { WIDGET_DEFINITIONS } from "@/components/dashboard/widget-registry";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Card, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { useState } from "react";

interface AddWidgetSheetProps {
    onAdd: (type: string) => void;
}

/**
 * AddWidgetSheet component.
 * A sheet that displays a gallery of available widgets to add to the dashboard.
 * @param props - The component props.
 * @returns The rendered component.
 */
export function AddWidgetSheet({ onAdd }: AddWidgetSheetProps) {
    const [open, setOpen] = useState(false);

    const handleAdd = (type: string) => {
        onAdd(type);
        setOpen(false);
    };

    return (
        <Sheet open={open} onOpenChange={setOpen}>
            <SheetTrigger asChild>
                <Button size="sm" className="gap-2">
                    <Plus className="h-4 w-4" /> Add Widget
                </Button>
            </SheetTrigger>
            <SheetContent className="w-[400px] sm:w-[540px]">
                <SheetHeader>
                    <SheetTitle>Add Widget</SheetTitle>
                    <SheetDescription>
                        Choose a widget to add to your dashboard. You can resize and reorder them later.
                    </SheetDescription>
                </SheetHeader>
                <ScrollArea className="h-[calc(100vh-8rem)] mt-6 pr-4">
                    <div className="grid grid-cols-1 gap-4">
                        {WIDGET_DEFINITIONS.map((widget) => {
                            const Icon = widget.icon;
                            return (
                                <Card
                                    key={widget.type}
                                    className="cursor-pointer hover:bg-muted/50 transition-colors border-dashed hover:border-solid group"
                                    onClick={() => handleAdd(widget.type)}
                                >
                                    <CardHeader className="flex flex-row items-center gap-4 space-y-0 p-4">
                                        <div className="p-2 bg-primary/10 rounded-md group-hover:bg-primary/20 transition-colors">
                                            <Icon className="h-6 w-6 text-primary" />
                                        </div>
                                        <div className="flex-1">
                                            <CardTitle className="text-base">{widget.title}</CardTitle>
                                            <CardDescription className="text-xs mt-1">
                                                {widget.description}
                                            </CardDescription>
                                        </div>
                                        <Button variant="ghost" size="icon" className="opacity-0 group-hover:opacity-100 transition-opacity">
                                            <Plus className="h-4 w-4" />
                                        </Button>
                                    </CardHeader>
                                </Card>
                            );
                        })}
                    </div>
                </ScrollArea>
            </SheetContent>
        </Sheet>
    );
}
