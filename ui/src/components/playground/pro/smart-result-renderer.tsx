/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useMemo } from "react";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Button } from "@/components/ui/button";
import { Code, Table as TableIcon, Image as ImageIcon } from "lucide-react";
import { JsonView } from "@/components/ui/json-view";

/**
 * Props for the SmartResultRenderer component.
 */
interface SmartResultRendererProps {
    /** The result object to render. Can be a JSON string, an object, or an array. */
    result: any;
}

interface McpContent {
    type: string;
    text?: string;
    data?: string;
    mimeType?: string;
}

const isMcpContent = (data: any): data is McpContent[] => {
    return Array.isArray(data) && data.length > 0 && data.every(item =>
        typeof item === 'object' && item !== null && 'type' in item &&
        (item.type === 'text' || item.type === 'image' || item.type === 'resource')
    );
};

const RichContentRenderer = ({ content }: { content: McpContent[] }) => {
    return (
        <div className="flex flex-col gap-4 w-full">
            {content.map((item, idx) => {
                if (item.type === 'image' && item.data && item.mimeType) {
                    return (
                        <div key={idx} className="flex flex-col gap-1 items-start">
                             <img
                                src={`data:${item.mimeType};base64,${item.data}`}
                                alt="Tool Result"
                                className="max-w-full rounded-md border shadow-sm"
                            />
                            <div className="text-[10px] text-muted-foreground font-mono opacity-70">
                                {item.mimeType} ({Math.round(item.data.length * 0.75 / 1024)} KB)
                            </div>
                        </div>
                    );
                }
                if (item.type === 'text' && item.text) {
                     // Check if text is JSON
                    let jsonContent = null;
                    try {
                        const parsed = JSON.parse(item.text);
                        if (typeof parsed === 'object' && parsed !== null) {
                            jsonContent = parsed;
                        }
                    } catch (e) {
                        // Not JSON
                    }

                    if (jsonContent) {
                         return <JsonView key={idx} data={jsonContent} maxHeight={400} />;
                    }

                    return (
                        <pre key={idx} className="whitespace-pre-wrap font-mono text-sm bg-muted/30 p-2 rounded-md overflow-x-auto border">
                            {item.text}
                        </pre>
                    );
                }
                return (
                    <div key={idx} className="p-2 border rounded bg-muted/20 text-xs font-mono">
                        Unknown content type: {item.type}
                    </div>
                );
            })}
        </div>
    );
}


/**
 * Renders the result of a tool execution in a smart, tabular format if possible,
 * falling back to a raw JSON view.
 *
 * @param props - The component props.
 * @param props.result - The result object to render. Can be a JSON string, an object, or an array.
 * @returns A React component that displays the result.
 */
export function SmartResultRenderer({ result }: SmartResultRendererProps) {
    const [viewMode, setViewMode] = useState<"smart" | "raw">("smart");

    // Attempt to parse content from the result
    const { smartContent, tableData, isMcp } = useMemo(() => {
        let content = result;

        // Helper to check if text content is actually a table (array of objects)
        const tryParseTable = (text: string) => {
            try {
                const parsed = JSON.parse(text);
                if (Array.isArray(parsed) && parsed.length > 0 && parsed.every((i: any) => typeof i === 'object' && i !== null && !Array.isArray(i))) {
                    return parsed;
                }
            } catch {}
            return null;
        };

        // 1. Unwrap CallToolResult structure
        if (result && typeof result === 'object' && Array.isArray(result.content)) {
            // Check if we should unwrap text to show Table (legacy behavior preservation)
             if (result.content.length === 1 && result.content[0].type === 'text') {
                 const table = tryParseTable(result.content[0].text);
                 if (table) {
                     return { smartContent: null, tableData: table, isMcp: false };
                 }
             }

            content = result.content;
            if (isMcpContent(content)) {
                return { smartContent: content, tableData: null, isMcp: true };
            }
        }

        // 2. Parse JSON string if it's a string
        if (typeof content === 'string') {
            try {
                content = JSON.parse(content);
            } catch (e) {
                // Not valid JSON string
                return { smartContent: null, tableData: null, isMcp: false };
            }
        }

        // 3. Handle Command Output wrapper (e.g. { stdout: "[...]", ... })
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
             } else if (content.content && Array.isArray(content.content)) {
                  // Maybe nested content?
                  content = content.content;
             }
        }

        // Re-check after unwrapping command output
        // It might be an MCP content array inside stdout
        if (isMcpContent(content)) {
             // Same check for table inside text item
             if (content.length === 1 && content[0].type === 'text') {
                 const table = tryParseTable(content[0].text);
                 if (table) {
                     return { smartContent: null, tableData: table, isMcp: false };
                 }
             }
             return { smartContent: content, tableData: null, isMcp: true };
        }

        // Also check if unwrapped content is { content: [...] } (nested CallToolResult)
        if (content && typeof content === 'object' && Array.isArray(content.content)) {
            if (isMcpContent(content.content)) {
                if (content.content.length === 1 && content.content[0].type === 'text') {
                    const table = tryParseTable(content.content[0].text);
                    if (table) {
                        return { smartContent: null, tableData: table, isMcp: false };
                    }
                }
                return { smartContent: content.content, tableData: null, isMcp: true };
            }
        }


        // 4. Final validation: Must be an array of objects for Table View
        if (Array.isArray(content) && content.length > 0) {
            // Check if items are objects
            const isListOfObjects = content.every((item: any) => typeof item === 'object' && item !== null && !Array.isArray(item));
            if (isListOfObjects) {
                return { smartContent: null, tableData: content, isMcp: false };
            }
        }

        return { smartContent: null, tableData: null, isMcp: false };
    }, [result]);

    const hasSmartView = isMcp || tableData !== null;

    const renderRaw = () => (
        <JsonView data={result} maxHeight={400} />
    );

    const renderSmart = () => {
        if (!hasSmartView) return renderRaw();

        if (isMcp && smartContent) {
            return <RichContentRenderer content={smartContent} />;
        }

        if (tableData) {
            // Determine columns from all keys in the first 10 rows (to be reasonably safe)
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
        }

        return renderRaw();
    };

    return (
        <div className="flex flex-col gap-0 w-full">
            {hasSmartView && (
                <div className="flex justify-end mb-1 px-1">
                     <div className="flex items-center bg-muted/50 rounded-lg p-0.5 border">
                         <Button
                            variant={viewMode === "smart" ? "secondary" : "ghost"}
                            size="sm"
                            className="h-6 px-2 text-[10px] gap-1"
                            onClick={() => setViewMode("smart")}
                         >
                             {isMcp ? <ImageIcon className="size-3" /> : <TableIcon className="size-3" />}
                             {isMcp ? "Visual" : "Table"}
                         </Button>
                         <Button
                            variant={viewMode === "raw" ? "secondary" : "ghost"}
                            size="sm"
                            className="h-6 px-2 text-[10px] gap-1"
                            onClick={() => setViewMode("raw")}
                         >
                             <Code className="size-3" /> JSON
                         </Button>
                     </div>
                </div>
            )}

            {viewMode === "smart" && hasSmartView ? renderSmart() : renderRaw()}
        </div>
    );
}
