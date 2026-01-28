/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useMemo } from "react";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Button } from "@/components/ui/button";
import { Code, Table as TableIcon, Copy, Check } from "lucide-react";
import { Prism as SyntaxHighlighter } from 'react-syntax-highlighter';
import { vscDarkPlus } from 'react-syntax-highlighter/dist/esm/styles/prism';
import { ScrollArea, ScrollBar } from "@/components/ui/scroll-area";

interface SmartResultRendererProps {
    result: unknown;
}

export function SmartResultRenderer({ result }: SmartResultRendererProps) {
    const [viewMode, setViewMode] = useState<"smart" | "raw">("smart");
    const [copied, setCopied] = useState(false);

    const copyToClipboard = () => {
        navigator.clipboard.writeText(JSON.stringify(result, null, 2));
        setCopied(true);
        setTimeout(() => setCopied(false), 2000);
    };

    // Unwrap MCP content if it's a JSON string
    const effectiveResult = useMemo(() => {
        let res = result;
        // 1. Unwrap content[0].text
        if (typeof res === 'object' && res !== null && 'content' in res && Array.isArray((res as any).content)) {
            const content = (res as any).content;
            if (content.length > 0 && content[0].type === 'text' && content[0].text) {
                try {
                    res = JSON.parse(content[0].text);
                } catch {
                    // Not JSON
                }
            }
        }

        // 2. Unwrap stdout if present (Command Service pattern)
        if (typeof res === 'object' && res !== null && 'stdout' in res && typeof (res as any).stdout === 'string') {
             try {
                const parsed = JSON.parse((res as any).stdout);
                // Only unwrap if it looks like complex data (array or object)
                if (typeof parsed === 'object' && parsed !== null) {
                    res = parsed;
                }
            } catch {
                // Not JSON
            }
        }

        return res;
    }, [result]);

    // Heuristic: Is it an array of objects?
    const tableData = useMemo(() => {
        if (Array.isArray(effectiveResult) && effectiveResult.length > 0 && typeof effectiveResult[0] === 'object' && effectiveResult[0] !== null) {
            // Check if all items are objects
            const isAllObjects = effectiveResult.every(item => typeof item === 'object' && item !== null && !Array.isArray(item));
            if (isAllObjects) {
                return effectiveResult as Record<string, unknown>[];
            }
        }
        return null;
    }, [effectiveResult]);

    // Heuristic: Is it a simple object?
    const objectData = useMemo(() => {
        if (typeof effectiveResult === 'object' && effectiveResult !== null && !Array.isArray(effectiveResult)) {
            // Count keys
            const keys = Object.keys(effectiveResult as Record<string, unknown>);
            // If it has too many keys or nested objects, maybe raw is better?
            // Let's stick to < 20 keys for grid view.
            if (keys.length > 0 && keys.length < 20) {
                 return effectiveResult as Record<string, unknown>;
            }
        }
        return null;
    }, [effectiveResult]);

    const canBeSmart = tableData !== null || objectData !== null;

    const renderValue = (val: unknown): string => {
        if (typeof val === 'object' && val !== null) {
            return JSON.stringify(val);
        }
        return String(val);
    };

    const renderSmartView = () => {
        if (tableData) {
            // Extract all unique keys for columns
            const allKeys = Array.from(new Set(tableData.flatMap(Object.keys)));
            // Limit columns if too many?
            const columns = allKeys.slice(0, 10); // Show first 10 columns
            const hasMoreColumns = allKeys.length > 10;

            return (
                <div className="border rounded-md overflow-hidden bg-card">
                     <ScrollArea className="h-full max-h-[400px]">
                        <Table>
                            <TableHeader>
                                <TableRow>
                                    <TableHead className="w-[50px]">#</TableHead>
                                    {columns.map(key => (
                                        <TableHead key={key} className="whitespace-nowrap">{key}</TableHead>
                                    ))}
                                    {hasMoreColumns && <TableHead>...</TableHead>}
                                </TableRow>
                            </TableHeader>
                            <TableBody>
                                {tableData.slice(0, 100).map((row, i) => (
                                    <TableRow key={i}>
                                        <TableCell className="text-muted-foreground text-xs">{i + 1}</TableCell>
                                        {columns.map(key => (
                                            <TableCell key={key} className="max-w-[300px] truncate text-xs font-mono">
                                                {renderValue(row[key])}
                                            </TableCell>
                                        ))}
                                         {hasMoreColumns && <TableCell>...</TableCell>}
                                    </TableRow>
                                ))}
                                {tableData.length > 100 && (
                                    <TableRow>
                                        <TableCell colSpan={columns.length + 2} className="text-center text-muted-foreground">
                                            {tableData.length - 100} more rows...
                                        </TableCell>
                                    </TableRow>
                                )}
                            </TableBody>
                        </Table>
                         <ScrollBar orientation="horizontal" />
                    </ScrollArea>
                </div>
            );
        }

        if (objectData) {
            return (
                <div className="border rounded-md bg-card p-4">
                    <div className="grid grid-cols-[minmax(120px,auto)_1fr] gap-x-4 gap-y-2 text-sm">
                        {Object.entries(objectData).map(([key, value]) => (
                            <div key={key} className="contents border-b border-border/50">
                                <span className="font-semibold text-muted-foreground py-1">{key}</span>
                                <span className="font-mono py-1 break-all">{renderValue(value)}</span>
                            </div>
                        ))}
                    </div>
                </div>
            );
        }

        return null;
    };

    return (
        <div className="flex flex-col gap-2 w-full">
            {canBeSmart && (
                <div className="flex justify-end gap-2 mb-1">
                    <div className="flex bg-muted p-0.5 rounded-lg">
                        <Button
                            variant="ghost"
                            size="sm"
                            className={`h-6 px-2 text-xs rounded-md ${viewMode === "smart" ? "bg-background shadow-sm text-foreground" : "text-muted-foreground hover:text-foreground"}`}
                            onClick={() => setViewMode("smart")}
                        >
                            <TableIcon className="w-3 h-3 mr-1" />
                            Table
                        </Button>
                        <Button
                            variant="ghost"
                            size="sm"
                            className={`h-6 px-2 text-xs rounded-md ${viewMode === "raw" ? "bg-background shadow-sm text-foreground" : "text-muted-foreground hover:text-foreground"}`}
                            onClick={() => setViewMode("raw")}
                        >
                            <Code className="w-3 h-3 mr-1" />
                            JSON
                        </Button>
                    </div>
                </div>
            )}

            {viewMode === "smart" && canBeSmart ? (
                renderSmartView()
            ) : (
                <div className="relative group/code max-h-[400px] overflow-auto border rounded-md">
                     <SyntaxHighlighter
                        language="json"
                        style={vscDarkPlus}
                        customStyle={{ margin: 0, padding: '1rem', fontSize: '12px', minHeight: '100%' }}
                        wrapLines={true}
                        wrapLongLines={true}
                    >
                        {JSON.stringify(result, null, 2)}
                    </SyntaxHighlighter>
                     <Button
                        size="icon"
                        variant="ghost"
                        className="absolute right-2 top-2 h-6 w-6 opacity-0 group-hover/code:opacity-100 transition-opacity bg-muted/20 hover:bg-muted/50 text-white"
                        onClick={copyToClipboard}
                    >
                        {copied ? <Check className="h-3 w-3" /> : <Copy className="h-3 w-3" />}
                    </Button>
                </div>
            )}
        </div>
    );
}
