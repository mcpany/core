/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useMemo } from "react";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Button } from "@/components/ui/button";
import { Code, Table as TableIcon, Image as ImageIcon, FileText, Terminal } from "lucide-react";
import { JsonView } from "@/components/ui/json-view";
import { CommandResultView } from "./command-result-view";

/**
 * Props for the SmartResultRenderer component.
 */
interface SmartResultRendererProps {
    /** The result object to render. Can be a JSON string, an object, or an array. */
    result: any;
}

interface McpContent {
    type: 'text' | 'image' | 'resource';
    text?: string;
    data?: string;
    mimeType?: string;
    resource?: any;
}

/**
 * Renders the result of a tool execution in a smart, tabular format if possible,
 * falling back to a raw JSON view.
 */
export function SmartResultRenderer({ result }: SmartResultRendererProps) {
    const [userViewMode, setUserViewMode] = useState<"smart" | "raw" | "rich" | "console" | null>(null);

    // 1. Shared unwrapping logic
    const unwrappedContent = useMemo(() => {
        let content = result;

        // Unwrap CallToolResult structure
        if (result && typeof result === 'object' && Array.isArray(result.content)) {
            content = result.content;
        }

        // Handle Command Output wrapper
        if (content && typeof content === 'object' && !Array.isArray(content)) {
             if (content.stdout && typeof content.stdout === 'string') {
                 try {
                     const inner = JSON.parse(content.stdout);
                     if (Array.isArray(inner) || (typeof inner === 'object' && inner !== null)) {
                         content = inner;
                     }
                 } catch (e) {
                     // stdout is not JSON
                 }
             }
        }

        // Handle deeply nested "content" (e.g. from stdout containing MCP content object)
        if (content && typeof content === 'object' && !Array.isArray(content) && Array.isArray(content.content)) {
            content = content.content;
        }

        return content;
    }, [result]);

    // 2. Identify MCP Content
    const mcpContent = useMemo<McpContent[] | null>(() => {
        if (Array.isArray(unwrappedContent) && unwrappedContent.length > 0) {
            const isMcp = unwrappedContent.every((item: any) =>
                typeof item === 'object' &&
                (item.type === 'text' || item.type === 'image' || item.type === 'resource')
            );
            if (isMcp) return unwrappedContent as McpContent[];
        }
        return null;
    }, [unwrappedContent]);

    // 3. Identify Table Data
    const tableData = useMemo(() => {
        // If MCP content, try to extract table data from text
        if (mcpContent) {
             const hasNonText = mcpContent.some(c => c.type !== 'text');
             if (hasNonText) return null;

             // Only support single text block for table view to avoid complexity
             if (mcpContent.length === 1 && mcpContent[0].text) {
                 try {
                    const parsed = JSON.parse(mcpContent[0].text);
                    if (Array.isArray(parsed) && parsed.every(item => typeof item === 'object')) {
                        return parsed;
                    }
                } catch (e) {}
             }
             return null;
        }

        // If NOT MCP content, check if unwrapped content itself is tabular data (CLI use case)
        if (Array.isArray(unwrappedContent) && unwrappedContent.length > 0) {
             const isTable = unwrappedContent.every((item: any) => typeof item === 'object' && item !== null);
             if (isTable) return unwrappedContent;
        }

        return null;
    }, [unwrappedContent, mcpContent]);

    // 4. Identify Command Result
    const commandResult = useMemo(() => {
        let content = result;
        // Unwrap CallToolResult structure to find the inner content
        if (result && typeof result === 'object' && Array.isArray(result.content)) {
            content = result.content;
        }

        // Case 1: content is an array (Standard MCP)
        // Check if the first item contains a JSON string that parses to a command result
        if (Array.isArray(content) && content.length > 0) {
            const firstItem = content[0];
            if (firstItem && typeof firstItem === 'object' && firstItem.type === 'text' && typeof firstItem.text === 'string') {
                try {
                    const parsed = JSON.parse(firstItem.text);
                    if (parsed && typeof parsed === 'object' && typeof parsed.stdout === 'string' && (typeof parsed.return_code === 'number' || typeof parsed.status === 'string')) {
                        return parsed;
                    }
                } catch (e) {
                    // Not JSON or parse error, ignore
                }
            }
        }

        // Case 2: content is already the command result object (Non-standard / CLI wrapper direct return)
        if (content && typeof content === 'object' && !Array.isArray(content)) {
             // It must have stdout (string) AND (return_code (number) OR status (string))
             if (typeof content.stdout === 'string' && (typeof content.return_code === 'number' || typeof content.status === 'string')) {
                 return content;
             }
        }
        return null;
    }, [result]);

    const activeView = useMemo(() => {
        // User override
        if (userViewMode === 'smart' && tableData) return 'smart';
        if (userViewMode === 'rich' && mcpContent) return 'rich';
        if (userViewMode === 'console' && commandResult) return 'console';
        if (userViewMode === 'raw') return 'raw';

        // Auto defaults (if user mode invalid or null)
        if (tableData) return 'smart';
        if (mcpContent) return 'rich';
        if (commandResult) return 'console';
        return 'raw';
    }, [userViewMode, tableData, mcpContent, commandResult]);

    const renderRaw = () => (
        <JsonView data={result} maxHeight={400} />
    );

    const renderRich = () => {
        if (!mcpContent) return renderRaw();

        return (
            <div className="flex flex-col gap-4 p-4 border rounded-md bg-muted/10">
                {mcpContent.map((item, idx) => (
                    <div key={idx} className="flex flex-col gap-2">
                        {item.type === 'text' && (
                            <div className="whitespace-pre-wrap font-mono text-sm bg-muted/30 p-3 rounded-md border border-white/5">
                                {item.text}
                            </div>
                        )}
                        {item.type === 'image' && item.data && (
                            <div className="flex flex-col gap-1 items-start">
                                <img
                                    src={`data:${item.mimeType || 'image/png'};base64,${item.data}`}
                                    alt="Tool Result"
                                    className="max-w-full h-auto rounded-lg border border-white/10 shadow-sm"
                                />
                                <span className="text-[10px] text-muted-foreground self-end">
                                    {item.mimeType}
                                </span>
                            </div>
                        )}
                        {item.type === 'resource' && (
                            <div className="flex items-center gap-2 p-3 bg-muted/30 rounded-md border border-white/5">
                                <FileText className="h-4 w-4 text-primary" />
                                <span className="text-sm font-medium">Resource: {item.resource?.uri || 'Unknown'}</span>
                            </div>
                        )}
                    </div>
                ))}
            </div>
        );
    };

    const renderConsole = () => {
        if (!commandResult) return renderRaw();
        return <CommandResultView result={commandResult} />;
    };

    const renderSmartTable = () => {
        if (!tableData) return null;

        // Determine columns from all keys in the first 10 rows
        const allKeys = new Set<string>();
        tableData.slice(0, 10).forEach((row: any) => {
            Object.keys(row).forEach(k => allKeys.add(k));
        });
        const columns = Array.from(allKeys);

        return (
            <div className="rounded-md border max-h-[400px] overflow-auto">
                <Table>
                    <TableHeader className="bg-muted/50 sticky top-0">
                        <TableRow>
                            {columns.map(col => (
                                <TableHead key={col} className="whitespace-nowrap font-medium text-xs px-2 py-1 h-8">
                                    {col}
                                </TableHead>
                            ))}
                        </TableRow>
                    </TableHeader>
                    <TableBody>
                        {tableData.map((row: any, idx: number) => (
                            <TableRow key={idx} className="hover:bg-muted/50">
                                {columns.map(col => {
                                    const val = row[col];
                                    let displayVal = val;
                                    if (typeof val === 'object' && val !== null) {
                                        displayVal = JSON.stringify(val);
                                    } else if (typeof val === 'boolean') {
                                        displayVal = val ? "true" : "false";
                                    }

                                    return (
                                        <TableCell key={col} className="px-2 py-1 text-xs max-w-[200px] truncate" title={String(displayVal)}>
                                            {String(displayVal ?? "")}
                                        </TableCell>
                                    );
                                })}
                            </TableRow>
                        ))}
                    </TableBody>
                </Table>
                <div className="bg-muted/30 px-2 py-1 text-[10px] text-muted-foreground border-t">
                    Showing {tableData.length} rows
                </div>
            </div>
        );
    };

    return (
        <div className="flex flex-col gap-0 w-full">
            <div className="flex justify-end mb-1 px-1">
                 <div className="flex items-center bg-muted/50 rounded-lg p-0.5 border">
                     {tableData && (
                        <Button
                            variant={activeView === "smart" ? "secondary" : "ghost"}
                            size="sm"
                            className="h-6 px-2 text-[10px] gap-1"
                            onClick={() => setUserViewMode("smart")}
                        >
                            <TableIcon className="size-3" /> Table
                        </Button>
                     )}
                     {mcpContent && (
                        <Button
                            variant={activeView === "rich" ? "secondary" : "ghost"}
                            size="sm"
                            className="h-6 px-2 text-[10px] gap-1"
                            onClick={() => setUserViewMode("rich")}
                        >
                            <ImageIcon className="size-3" /> Rich
                        </Button>
                     )}
                     {commandResult && (
                        <Button
                            variant={activeView === "console" ? "secondary" : "ghost"}
                            size="sm"
                            className="h-6 px-2 text-[10px] gap-1"
                            onClick={() => setUserViewMode("console")}
                        >
                            <Terminal className="size-3" /> Console
                        </Button>
                     )}
                     <Button
                        variant={activeView === "raw" ? "secondary" : "ghost"}
                        size="sm"
                        className="h-6 px-2 text-[10px] gap-1"
                        onClick={() => setUserViewMode("raw")}
                     >
                         <Code className="size-3" /> JSON
                     </Button>
                 </div>
            </div>

            {activeView === 'smart' && renderSmartTable()}
            {activeView === 'rich' && renderRich()}
            {activeView === 'console' && renderConsole()}
            {activeView === 'raw' && renderRaw()}
        </div>
    );
}
