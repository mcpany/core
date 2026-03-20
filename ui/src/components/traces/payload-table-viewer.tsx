"use client";

import { useState } from "react";
import { ChevronRight, ChevronDown } from "lucide-react";
import { cn } from "@/lib/utils";

interface PayloadTableViewerProps {
    data: unknown;
    label?: string;
    level?: number;
    initiallyExpanded?: boolean;
}

export function PayloadTableViewer({
    data,
    label,
    level = 0,
    initiallyExpanded = true
}: PayloadTableViewerProps) {
    const [isExpanded, setIsExpanded] = useState(initiallyExpanded);

    const isObject = data !== null && typeof data === 'object';
    const isArray = Array.isArray(data);

    if (data === undefined) {
        return null;
    }

    const toggleExpand = () => setIsExpanded(!isExpanded);

    const renderValue = (val: unknown) => {
        if (val === null) return <span className="text-muted-foreground italic">null</span>;
        if (typeof val === 'boolean') return <span className="text-blue-500 font-mono">{val ? 'true' : 'false'}</span>;
        if (typeof val === 'number') return <span className="text-emerald-500 font-mono">{val}</span>;
        if (typeof val === 'string') return <span className="text-amber-600 dark:text-amber-400 break-all">"{val}"</span>;
        return <span className="text-muted-foreground">Unsupported type</span>;
    };

    if (!isObject) {
        return (
            <div className="flex border-b border-border/40 last:border-0 hover:bg-muted/30 transition-colors py-2 text-sm">
                <div
                    className="w-1/3 font-medium text-muted-foreground shrink-0 pl-4 py-1"
                    style={{ paddingLeft: `${level * 16 + 16}px` }}
                >
                    {label}
                </div>
                <div className="w-2/3 px-4 py-1 break-words font-mono text-xs flex items-center">
                    {renderValue(data)}
                </div>
            </div>
        );
    }

    const entries = Object.entries(data as Record<string, unknown>);
    const isEmpty = entries.length === 0;

    const summary = isArray
        ? `Array(${(data as unknown[]).length})`
        : `Object(${entries.length})`;

    return (
        <div className="flex flex-col border-b border-border/40 last:border-0">
            <div
                className={cn(
                    "flex hover:bg-muted/30 transition-colors py-2 cursor-pointer text-sm",
                    level === 0 && label === undefined && "hidden"
                )}
                onClick={toggleExpand}
            >
                <div
                    className="w-1/3 font-medium text-foreground shrink-0 flex items-center gap-1 pl-4 py-1 select-none"
                    style={{ paddingLeft: `${Math.max(0, level * 16 + (isEmpty ? 16 : 0))}px` }}
                >
                    {!isEmpty && (
                        <div className="text-muted-foreground">
                            {isExpanded ? <ChevronDown className="h-3.5 w-3.5" /> : <ChevronRight className="h-3.5 w-3.5" />}
                        </div>
                    )}
                    <span>{label}</span>
                </div>
                <div className="w-2/3 px-4 py-1 flex items-center text-xs text-muted-foreground font-mono">
                    {(!isExpanded || isEmpty) && summary}
                </div>
            </div>

            {isExpanded && !isEmpty && (
                <div className={cn("flex flex-col w-full", (level > 0 || label) && "border-l border-border/20 ml-2")}>
                    {entries.map(([key, value]) => (
                        <PayloadTableViewer
                            key={key}
                            label={isArray ? `[${key}]` : key}
                            data={value}
                            level={label !== undefined ? level + 1 : level}
                            initiallyExpanded={level < 2}
                        />
                    ))}
                </div>
            )}
        </div>
    );
}
