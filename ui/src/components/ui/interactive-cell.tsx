/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { useState } from "react";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";
import { JsonTree } from "@/components/ui/json-tree";
import { ScrollArea } from "@/components/ui/scroll-area";

export function InteractiveCell({ value }: { value: any }) {
    const [open, setOpen] = useState(false);

    if (value === null || value === undefined) {
        return <span className="text-muted-foreground">-</span>;
    }

    if (typeof value === 'boolean') {
        return <span className={value ? "text-green-500 font-medium" : "text-red-500 font-medium"}>{String(value)}</span>;
    }

    if (typeof value === 'object') {
        const isArray = Array.isArray(value);
        const keys = Object.keys(value);
        const isEmpty = keys.length === 0;

        if (isEmpty) {
            return <span className="text-muted-foreground">{isArray ? "[]" : "{}"}</span>;
        }

        const label = isArray ? `Array(${keys.length})` : `Object(${keys.length})`;

        return (
            <Popover open={open} onOpenChange={setOpen}>
                <PopoverTrigger asChild>
                    <button className="inline-flex items-center justify-center rounded-md border border-input bg-background/50 hover:bg-muted px-2 py-0.5 text-[10px] font-medium transition-colors focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring cursor-pointer text-muted-foreground hover:text-foreground">
                        {label}
                    </button>
                </PopoverTrigger>
                <PopoverContent className="w-80 p-0 shadow-lg border-muted/50 rounded-lg overflow-hidden z-50 animate-in zoom-in-95 duration-200" align="start">
                    <div className="bg-muted/50 px-3 py-2 border-b text-xs font-semibold flex justify-between items-center backdrop-blur-sm">
                        <span>{isArray ? "Array Details" : "Object Details"}</span>
                        <span className="text-muted-foreground text-[10px] font-normal">{keys.length} items</span>
                    </div>
                    <ScrollArea className="max-h-64 bg-[#1e1e1e]">
                        <div className="p-3">
                            <JsonTree data={value} defaultExpandedLevel={2} />
                        </div>
                    </ScrollArea>
                </PopoverContent>
            </Popover>
        );
    }

    return <span className="truncate max-w-[300px] block" title={String(value)}>{String(value)}</span>;
}
