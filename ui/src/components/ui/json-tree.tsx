/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import React, { useState, useEffect } from "react";
import { ChevronRight, ChevronDown, Copy, Check } from "lucide-react";
import { cn } from "@/lib/utils";
import { Button } from "@/components/ui/button";
import { useToast } from "@/hooks/use-toast";

/**
 * Props for the JsonTree component.
 */
export interface JsonTreeProps {
  /** The data object to be rendered as a tree. */
  data: unknown;
  /** Initial depth to expand. Default is 1. */
  defaultExpandDepth?: number;
  /** Whether to show the root object braces/brackets. Default true. */
  showRoot?: boolean;
  /** Optional CSS class names. */
  className?: string;
}

interface JsonTreeNodeProps {
  name?: string;
  value: unknown;
  isLast: boolean;
  depth: number;
  defaultExpandDepth: number;
}

/**
 * Copies text to clipboard with toast feedback.
 */
const useClipboard = () => {
    const [copied, setCopied] = useState(false);
    const { toast } = useToast();

    const copy = (text: string, description = "Copied to clipboard") => {
        if (!text) return;
        navigator.clipboard.writeText(text);
        setCopied(true);
        toast({
            title: "Copied",
            description: description,
            duration: 2000,
        });
        setTimeout(() => setCopied(false), 2000);
    };

    return { copied, copy };
};

const ValueDisplay = ({ value }: { value: unknown }) => {
    if (value === null) return <span className="text-muted-foreground italic">null</span>;
    if (value === undefined) return <span className="text-muted-foreground italic">undefined</span>;
    if (typeof value === "boolean") return <span className="text-purple-600 dark:text-purple-400">{value.toString()}</span>;
    if (typeof value === "number") return <span className="text-blue-600 dark:text-blue-400">{value}</span>;
    if (typeof value === "string") return <span className="text-green-600 dark:text-green-400">"{value}"</span>;
    return <span>{String(value)}</span>;
};

const JsonTreeNode = ({ name, value, isLast, depth, defaultExpandDepth }: JsonTreeNodeProps) => {
    const isObject = value !== null && typeof value === "object";
    const isArray = Array.isArray(value);
    const isEmpty = isObject && Object.keys(value as object).length === 0;

    // Determine initial expansion state based on depth and defaultExpandDepth
    const [expanded, setExpanded] = useState(depth < defaultExpandDepth);
    const { copied, copy } = useClipboard();
    const [hovered, setHovered] = useState(false);

    // If it's empty, no need to expand
    useEffect(() => {
        if (isEmpty) setExpanded(false);
    }, [isEmpty]);

    if (!isObject) {
        return (
            <div
                className="flex items-start group/line hover:bg-muted/30 rounded px-1 -ml-1"
                onMouseEnter={() => setHovered(true)}
                onMouseLeave={() => setHovered(false)}
            >
                <div style={{ paddingLeft: `${depth * 1.5}rem` }} className="flex items-center gap-1 font-mono text-xs w-full">
                    {name && <span className="text-foreground/80 font-medium">{name}: </span>}
                    <ValueDisplay value={value} />
                    {!isLast && <span className="text-muted-foreground">,</span>}

                    {hovered && (
                        <Button
                            variant="ghost"
                            size="icon"
                            className="h-4 w-4 ml-2 opacity-50 hover:opacity-100"
                            onClick={() => copy(String(value), "Value copied")}
                            title="Copy Value"
                        >
                            {copied ? <Check className="h-2 w-2" /> : <Copy className="h-2 w-2" />}
                        </Button>
                    )}
                </div>
            </div>
        );
    }

    const keys = Object.keys(value as object);
    const openBracket = isArray ? "[" : "{";
    const closeBracket = isArray ? "]" : "}";
    const itemCount = keys.length;

    return (
        <div className="font-mono text-xs">
             <div
                className="flex items-center group/line hover:bg-muted/30 rounded px-1 -ml-1 cursor-pointer select-none"
                onClick={(e) => {
                    e.stopPropagation();
                    if (!isEmpty) setExpanded(!expanded);
                }}
                onMouseEnter={() => setHovered(true)}
                onMouseLeave={() => setHovered(false)}
            >
                <div style={{ paddingLeft: `${(depth * 1.5)}rem` }} className="flex items-center gap-1 w-full relative">
                    {/* Expand/Collapse Icon - positioned absolutely to the left of the content if we wanted, but inline is fine */}
                    <div className="w-4 h-4 flex items-center justify-center -ml-5 absolute" style={{ left: `${(depth * 1.5)}rem` }}>
                         {!isEmpty && (
                             expanded ?
                                <ChevronDown className="h-3 w-3 text-muted-foreground" /> :
                                <ChevronRight className="h-3 w-3 text-muted-foreground" />
                         )}
                    </div>

                    {name && <span className="text-foreground/80 font-medium">{name}: </span>}
                    <span className="text-muted-foreground font-semibold">{openBracket}</span>

                    {!expanded && (
                        <div className="flex items-center gap-1">
                             <span className="text-muted-foreground italic text-[10px] ml-1">
                                {isEmpty ? "" : `... ${itemCount} item${itemCount !== 1 ? 's' : ''} ...`}
                             </span>
                             <span className="text-muted-foreground font-semibold">{closeBracket}</span>
                             {!isLast && <span className="text-muted-foreground">,</span>}
                        </div>
                    )}

                    {hovered && !expanded && (
                         <Button
                            variant="ghost"
                            size="icon"
                            className="h-4 w-4 ml-2 opacity-50 hover:opacity-100"
                            onClick={(e) => {
                                e.stopPropagation();
                                copy(JSON.stringify(value, null, 2));
                            }}
                            title="Copy Object"
                        >
                            {copied ? <Check className="h-2 w-2" /> : <Copy className="h-2 w-2" />}
                        </Button>
                    )}
                </div>
            </div>

            {expanded && (
                <div>
                    {keys.map((key, index) => (
                        <JsonTreeNode
                            key={key}
                            name={isArray ? undefined : key} // Don't show index keys for arrays
                            value={(value as any)[key]}
                            isLast={index === keys.length - 1}
                            depth={depth + 1}
                            defaultExpandDepth={defaultExpandDepth}
                        />
                    ))}
                    <div
                        className="hover:bg-muted/30 rounded px-1 -ml-1"
                         onMouseEnter={() => setHovered(true)}
                         onMouseLeave={() => setHovered(false)}
                    >
                        <div style={{ paddingLeft: `${(depth * 1.5)}rem` }} className="flex items-center">
                            <span className="text-muted-foreground font-semibold">{closeBracket}</span>
                            {!isLast && <span className="text-muted-foreground">,</span>}

                            {hovered && (
                                <Button
                                    variant="ghost"
                                    size="icon"
                                    className="h-4 w-4 ml-2 opacity-50 hover:opacity-100"
                                    onClick={() => copy(JSON.stringify(value, null, 2))}
                                    title="Copy Object"
                                >
                                    {copied ? <Check className="h-2 w-2" /> : <Copy className="h-2 w-2" />}
                                </Button>
                            )}
                        </div>
                    </div>
                </div>
            )}
        </div>
    );
};

/**
 * A recursive, interactive JSON tree viewer.
 */
export function JsonTree({ data, defaultExpandDepth = 1, className }: JsonTreeProps) {
    return (
        <div className={cn("overflow-x-auto", className)}>
            <JsonTreeNode
                value={data}
                isLast={true}
                depth={0}
                defaultExpandDepth={defaultExpandDepth}
            />
        </div>
    );
}
