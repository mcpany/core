/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { Sheet, SheetContent, SheetHeader, SheetTitle, SheetDescription } from "@/components/ui/sheet";
import { AVAILABLE_WIDGETS } from "@/components/dashboard/widget-registry";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { Plus } from "lucide-react";

interface AddWidgetSheetProps {
    open: boolean;
    onOpenChange: (open: boolean) => void;
    onAddWidget: (type: string) => void;
}

export function AddWidgetSheet({ open, onOpenChange, onAddWidget }: AddWidgetSheetProps) {
    return (
        <Sheet open={open} onOpenChange={onOpenChange}>
            <SheetContent side="right" className="sm:max-w-md overflow-y-auto">
                <SheetHeader className="mb-6">
                    <SheetTitle>Add Widget</SheetTitle>
                    <SheetDescription>
                        Choose a widget to add to your dashboard.
                    </SheetDescription>
                </SheetHeader>

                <div className="grid gap-4">
                    {AVAILABLE_WIDGETS.map((widget) => (
                        <Card
                            key={widget.id}
                            className="p-4 flex items-start gap-4 hover:bg-muted/50 transition-colors cursor-pointer group"
                            onClick={() => {
                                onAddWidget(widget.id);
                                onOpenChange(false);
                            }}
                        >
                            <div className="p-2 bg-background border rounded-md group-hover:border-primary/50 transition-colors">
                                <widget.icon className="h-5 w-5 text-muted-foreground group-hover:text-primary" />
                            </div>
                            <div className="flex-1">
                                <h4 className="font-medium text-sm mb-1 group-hover:text-primary transition-colors">{widget.title}</h4>
                                <p className="text-xs text-muted-foreground line-clamp-2">{widget.description}</p>
                            </div>
                            <Button size="icon" variant="ghost" className="h-8 w-8 -mr-2 text-muted-foreground group-hover:text-primary">
                                <Plus className="h-4 w-4" />
                            </Button>
                        </Card>
                    ))}
                </div>
            </SheetContent>
        </Sheet>
    );
}
