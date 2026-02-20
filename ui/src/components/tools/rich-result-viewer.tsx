/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useMemo } from "react";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Prism as SyntaxHighlighter } from 'react-syntax-highlighter';
import { vscDarkPlus } from 'react-syntax-highlighter/dist/esm/styles/prism';
import { FileJson, Table as TableIcon, Terminal } from "lucide-react";

interface RichResultViewerProps {
    result: any;
}

/**
 * RichResultViewer displays tool execution results in a user-friendly format.
 * It automatically detects if the result contains JSON or tabular data and provides
 * appropriate views (Table, JSON, Raw).
 *
 * @param props - The component props.
 * @param props.result - The raw result object from the tool execution.
 * @returns The rendered component.
 */
export function RichResultViewer({ result }: RichResultViewerProps) {
    // Attempt to extract meaningful content if it's a command result
    const [content, isExtracted] = useMemo(() => {
        if (!result) return [result, false];

        // Handle Command Execution Result (stdout contains JSON)
        if (typeof result === 'object' && 'stdout' in result && typeof result.stdout === 'string') {
            try {
                // Only treat as extracted if parsing succeeds
                const parsed = JSON.parse(result.stdout);
                return [parsed, true];
            } catch {
                return [result, false];
            }
        }

        // Handle raw string that is JSON
        if (typeof result === 'string') {
             try {
                const parsed = JSON.parse(result);
                return [parsed, true];
            } catch {
                return [result, false];
            }
        }
        return [result, false];
    }, [result]);

    const isTableEligible = useMemo(() => {
        return Array.isArray(content) && content.length > 0 && typeof content[0] === 'object' && content[0] !== null;
    }, [content]);

    // Get columns for table
    const columns = useMemo(() => {
        if (!isTableEligible) return [];
        // aggregate all keys from all objects to handle sparse data
        const keys = new Set<string>();
        // Limit rows scanned for columns to avoid perf issues on huge datasets
        content.slice(0, 50).forEach((item: any) => {
            if (typeof item === 'object' && item !== null) {
                Object.keys(item).forEach(k => keys.add(k));
            }
        });
        return Array.from(keys);
    }, [content, isTableEligible]);

    const renderCell = (value: any) => {
        if (value === null || value === undefined) return <span className="text-muted-foreground">-</span>;
        if (typeof value === 'object') return <span className="font-mono text-xs text-muted-foreground truncate max-w-[200px] block" title={JSON.stringify(value)}>{JSON.stringify(value)}</span>;
        if (typeof value === 'boolean') return <span className={value ? "text-green-500 font-medium" : "text-red-500 font-medium"}>{String(value)}</span>;
        return <span className="truncate max-w-[300px] block" title={String(value)}>{String(value)}</span>;
    }

    return (
        <Tabs defaultValue={isTableEligible ? "table" : "json"} className="w-full">
            <div className="flex items-center justify-between mb-2">
                <TabsList>
                    {isTableEligible && (
                        <TabsTrigger value="table" className="flex items-center gap-2">
                            <TableIcon className="h-4 w-4" /> Table
                        </TabsTrigger>
                    )}
                    <TabsTrigger value="json" className="flex items-center gap-2">
                        <FileJson className="h-4 w-4" /> JSON
                    </TabsTrigger>
                    {isExtracted && (
                         <TabsTrigger value="raw" className="flex items-center gap-2">
                            <Terminal className="h-4 w-4" /> Raw Output
                        </TabsTrigger>
                    )}
                </TabsList>
            </div>

            {isTableEligible && (
                <TabsContent value="table" className="border rounded-md">
                    <ScrollArea className="h-[400px]">
                        <Table>
                            <TableHeader>
                                <TableRow>
                                    {columns.map(col => (
                                        <TableHead key={col} className="whitespace-nowrap">{col}</TableHead>
                                    ))}
                                </TableRow>
                            </TableHeader>
                            <TableBody>
                                {content.map((row: any, i: number) => (
                                    <TableRow key={i}>
                                        {columns.map(col => (
                                            <TableCell key={col} className="py-2">
                                                {renderCell(row[col])}
                                            </TableCell>
                                        ))}
                                    </TableRow>
                                ))}
                            </TableBody>
                        </Table>
                    </ScrollArea>
                </TabsContent>
            )}

            <TabsContent value="json">
                <div className="rounded-md overflow-hidden border">
                    <SyntaxHighlighter
                        language="json"
                        style={vscDarkPlus}
                        customStyle={{ margin: 0, fontSize: '12px', maxHeight: '400px', overflowY: 'auto' }}
                        showLineNumbers
                    >
                        {JSON.stringify(content, null, 2)}
                    </SyntaxHighlighter>
                </div>
            </TabsContent>

             {isExtracted && (
                <TabsContent value="raw">
                    <div className="rounded-md overflow-hidden border">
                        <SyntaxHighlighter
                            language="json"
                            style={vscDarkPlus}
                            customStyle={{ margin: 0, fontSize: '12px', maxHeight: '400px', overflowY: 'auto' }}
                        >
                            {JSON.stringify(result, null, 2)}
                        </SyntaxHighlighter>
                    </div>
                </TabsContent>
            )}
        </Tabs>
    );
}
