/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import React, { useState, useMemo } from "react";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Input } from "@/components/ui/input";
import { Search, Copy, Check, ChevronRight, ChevronDown } from "lucide-react";
import { Button } from "@/components/ui/button";
import { JsonView } from "@/components/ui/json-view";
import { ScrollArea } from "@/components/ui/scroll-area";
import { cn } from "@/lib/utils";

interface StructuredPayloadViewerProps {
    data: any;
    maxHeight?: number;
}

interface FlatProperty {
    key: string;
    value: any;
    type: string;
    level: number;
    parentKey?: string;
    isExpandable: boolean;
    rawKey: string;
}

function flattenObject(obj: any, parentKey = '', level = 0): FlatProperty[] {
    let result: FlatProperty[] = [];

    if (obj === null || obj === undefined) {
         return [];
    }

    if (typeof obj !== 'object') {
         return [{
             key: parentKey,
             rawKey: parentKey,
             value: obj,
             type: typeof obj,
             level: 0,
             isExpandable: false
         }];
    }

    Object.keys(obj).forEach(key => {
        const fullKey = parentKey ? `${parentKey}.${key}` : key;
        const value = obj[key];
        const type = value === null ? 'null' : Array.isArray(value) ? 'array' : typeof value;
        const isExpandable = value !== null && typeof value === 'object';

        result.push({
            key: fullKey,
            rawKey: key,
            value: value,
            type: type,
            level: level,
            parentKey: parentKey,
            isExpandable: isExpandable
        });

        if (isExpandable) {
            result = result.concat(flattenObject(value, fullKey, level + 1));
        }
    });

    return result;
}

export function StructuredPayloadViewer({ data, maxHeight = 400 }: StructuredPayloadViewerProps) {
    const [searchQuery, setSearchQuery] = useState("");
    const [copiedKey, setCopiedKey] = useState<string | null>(null);
    const [expandedKeys, setExpandedKeys] = useState<Set<string>>(new Set());

    const flatProperties = useMemo(() => flattenObject(data), [data]);

    // Only show top-level properties by default, unless searching
    const visibleProperties = useMemo(() => {
        if (!searchQuery) {
            return flatProperties.filter(prop =>
                prop.level === 0 || expandedKeys.has(prop.parentKey!)
            );
        }

        const lowerQuery = searchQuery.toLowerCase();
        return flatProperties.filter(prop =>
            prop.key.toLowerCase().includes(lowerQuery) ||
            String(prop.value).toLowerCase().includes(lowerQuery)
        );
    }, [flatProperties, searchQuery, expandedKeys]);

    const toggleExpand = (key: string) => {
        setExpandedKeys(prev => {
            const next = new Set(prev);
            if (next.has(key)) {
                next.delete(key);
                // Also collapse children
                flatProperties.forEach(p => {
                    if (p.key.startsWith(`${key}.`)) {
                        next.delete(p.key);
                    }
                });
            } else {
                next.add(key);
            }
            return next;
        });
    };

    const handleCopy = (value: any, key: string) => {
        let textToCopy = typeof value === 'string' ? value : JSON.stringify(value, null, 2);
        navigator.clipboard.writeText(textToCopy);
        setCopiedKey(key);
        setTimeout(() => setCopiedKey(null), 2000);
    };

    const renderValueCell = (prop: FlatProperty) => {
        if (prop.isExpandable) {
            if (prop.type === 'array') {
                return <span className="text-muted-foreground font-mono text-xs">Array({(prop.value as any[]).length})</span>;
            }
            return <span className="text-muted-foreground font-mono text-xs">Object({Object.keys(prop.value).length})</span>;
        }

        if (prop.type === 'string') {
            const isUrl = typeof prop.value === 'string' && (prop.value.startsWith('http://') || prop.value.startsWith('https://'));
            if (isUrl) {
                 return (
                     <a href={prop.value} target="_blank" rel="noopener noreferrer" className="text-primary hover:underline truncate block max-w-md">
                         {prop.value}
                     </a>
                 );
            }
            return <span className="text-green-600 dark:text-green-400 break-all">"{prop.value}"</span>;
        }

        if (prop.type === 'number') {
            return <span className="text-blue-600 dark:text-blue-400">{prop.value}</span>;
        }

        if (prop.type === 'boolean') {
            return <span className="text-amber-600 dark:text-amber-400 font-semibold">{prop.value ? 'true' : 'false'}</span>;
        }

        if (prop.type === 'null') {
            return <span className="text-muted-foreground italic">null</span>;
        }

        return <span className="truncate block">{String(prop.value)}</span>;
    };

    if (data === null || data === undefined || (typeof data === 'object' && Object.keys(data).length === 0)) {
         return (
             <div className="flex items-center justify-center h-24 text-muted-foreground italic border rounded-md bg-muted/10">
                 Empty payload
             </div>
         );
    }

    return (
        <div className="flex flex-col space-y-3">
             <div className="relative max-w-sm">
                <Search className="absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" />
                <Input
                    placeholder="Search properties..."
                    value={searchQuery}
                    onChange={(e) => setSearchQuery(e.target.value)}
                    className="pl-9 bg-background shadow-sm"
                />
            </div>
            <div className="rounded-md border bg-card shadow-sm overflow-hidden">
                <ScrollArea style={{ maxHeight: `${maxHeight}px` }}>
                    <Table>
                        <TableHeader className="bg-muted/30 sticky top-0 backdrop-blur-md z-10">
                            <TableRow>
                                <TableHead className="w-[40%] font-semibold">Property</TableHead>
                                <TableHead className="font-semibold">Value</TableHead>
                                <TableHead className="w-[60px]"></TableHead>
                            </TableRow>
                        </TableHeader>
                        <TableBody className="font-mono text-sm">
                            {visibleProperties.map((prop) => (
                                <TableRow key={prop.key} className="hover:bg-muted/30 group">
                                    <TableCell className="py-2 flex items-center gap-1">
                                         <div style={{ width: `${prop.level * 16}px` }} className="shrink-0" />
                                         {prop.isExpandable ? (
                                             <Button
                                                variant="ghost"
                                                size="icon"
                                                className="h-5 w-5 shrink-0"
                                                onClick={() => toggleExpand(prop.key)}
                                            >
                                                {expandedKeys.has(prop.key) ? (
                                                    <ChevronDown className="h-3.5 w-3.5 text-muted-foreground" />
                                                ) : (
                                                    <ChevronRight className="h-3.5 w-3.5 text-muted-foreground" />
                                                )}
                                             </Button>
                                         ) : (
                                             <div className="w-5 shrink-0" />
                                         )}
                                         <span className={cn("truncate font-medium", prop.isExpandable ? "text-foreground" : "text-muted-foreground")} title={prop.key}>
                                            {prop.rawKey}
                                         </span>
                                    </TableCell>
                                    <TableCell className="py-2">
                                        <div className="max-h-24 overflow-y-auto">
                                            {renderValueCell(prop)}
                                        </div>
                                    </TableCell>
                                    <TableCell className="py-2 text-right">
                                         {!prop.isExpandable && (
                                            <Button
                                                variant="ghost"
                                                size="icon"
                                                className="h-6 w-6 opacity-0 group-hover:opacity-100 transition-opacity"
                                                onClick={() => handleCopy(prop.value, prop.key)}
                                                title="Copy Value"
                                            >
                                                {copiedKey === prop.key ? <Check className="h-3 w-3 text-green-500" /> : <Copy className="h-3 w-3" />}
                                            </Button>
                                         )}
                                    </TableCell>
                                </TableRow>
                            ))}
                            {visibleProperties.length === 0 && (
                                <TableRow>
                                    <TableCell colSpan={3} className="h-24 text-center text-muted-foreground">
                                        No matching properties found.
                                    </TableCell>
                                </TableRow>
                            )}
                        </TableBody>
                    </Table>
                </ScrollArea>
            </div>
        </div>
    );
}
