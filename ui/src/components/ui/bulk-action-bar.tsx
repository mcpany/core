/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import React from "react";
import { X } from "lucide-react";
import { Button } from "@/components/ui/button";

export interface BulkActionBarProps {
    selectedCount: number;
    onClearSelection: () => void;
    children: React.ReactNode;
}

/**
 * BulkActionBar component.
 * Displays a fixed action bar at the bottom of the screen when items are selected.
 */
export function BulkActionBar({ selectedCount, onClearSelection, children }: BulkActionBarProps) {
    if (selectedCount === 0) return null;

    return (
        <div className="fixed bottom-6 left-1/2 -translate-x-1/2 z-50 animate-in slide-in-from-bottom-8 fade-in duration-300">
            <div className="flex items-center gap-3 px-4 py-3 bg-background/80 backdrop-blur-md border shadow-lg rounded-full">
                <div className="flex items-center gap-2 pr-3 border-r">
                    <span className="text-sm font-medium whitespace-nowrap">
                        {selectedCount} {selectedCount === 1 ? 'selected' : 'selected'}
                    </span>
                    <Button
                        variant="ghost"
                        size="icon"
                        className="h-6 w-6 rounded-full hover:bg-muted"
                        onClick={onClearSelection}
                    >
                        <X className="h-4 w-4" />
                        <span className="sr-only">Clear selection</span>
                    </Button>
                </div>
                <div className="flex items-center gap-2">
                    {children}
                </div>
            </div>
        </div>
    );
}
