/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useMemo } from "react";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Button } from "@/components/ui/button";
import { Code, Table as TableIcon, Image as ImageIcon, FileText } from "lucide-react";
import { JsonView } from "@/components/ui/json-view";

/**
 * Props for the SmartResultRenderer component.
 */
interface SmartResultRendererProps {
    /** The result object to render. Can be a JSON string, an object, or an array. */
    result: any;
}

// Helper to check if it looks like MCP content array
const isMcpContent = (content: any): content is any[] => {
    if (!Array.isArray(content)) return false;
    // Check if it has any item with explicit type 'image' or 'resource' or 'text'
    // And specifically if it contains non-text content or multiple items, we should treat it as rich content.
    return content.some(item =>
        typeof item === 'object' && item !== null &&
        (item.type === 'image' || item.type === 'resource' || (item.type === 'text' && content.length > 1))
    );
};

const RichContentRenderer = ({ content }: { content: any[] }) => {
    return (
        <div className="flex flex-col gap-4">
            {content.map((item, idx) => {
                if (item.type === 'image' && item.data && item.mimeType) {
                    return (
                        <div key={idx} className="rounded-lg overflow-hidden border bg-muted/20">
                            <div className="text-[10px] text-muted-foreground px-2 py-1 bg-muted/30 border-b flex justify-between items-center">
                                <span className="flex items-center gap-1"><ImageIcon className="w-3 h-3" /> Image {idx + 1}</span>
                                <span className="font-mono opacity-70">{item.mimeType}</span>
                            </div>
                            <div className="p-2 flex justify-center bg-[url('/transparent-bg.svg')] bg-repeat">
                                <img
                                    src={`data:${item.mimeType};base64,${item.data}`}
                                    alt={`Tool Output ${idx}`}
                                    className="max-w-full max-h-[500px] object-contain shadow-sm"
                                />
                            </div>
                        </div>
                    );
                }
                if (item.type === 'text' && item.text) {
                     return (
                        <div key={idx} className="flex flex-col gap-0 rounded-md border overflow-hidden">
                            <div className="text-[10px] text-muted-foreground px-2 py-1 bg-muted/30 border-b flex items-center gap-1">
                                <FileText className="w-3 h-3" /> Text
                            </div>
                            <div className="bg-muted/10 p-3 text-sm whitespace-pre-wrap font-mono">
                                {item.text}
                            </div>
                        </div>
                     )
                }
                 if (item.type === 'resource') {
                     return (
                        <div key={idx} className="bg-muted/30 p-3 rounded-md border text-sm font-mono">
                            <div className="font-semibold text-xs text-muted-foreground mb-1">Resource: {item.resource?.uri}</div>
                            <JsonView data={item.resource} />
                        </div>
                     )
                }
                // Fallback for other types or invalid items
                return (
                     <div key={idx} className="bg-muted/30 p-2 rounded-md border text-xs">
                        <JsonView data={item} />
                     </div>
                );
            })}
        </div>
    );
};


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

    // Check for Rich Content (Images/Mixed) FIRST
    const richContent = useMemo(() => {
        let content: any = null;

         // 1. Unwrap CallToolResult structure
        if (result && typeof result === 'object' && Array.isArray(result.content)) {
            content = result.content;
        }
        // 2. Handle Command Output wrapper (parse stdout)
        else if (result && typeof result === 'object' && result.stdout && typeof result.stdout === 'string') {
             try {
                 const inner = JSON.parse(result.stdout);
                 // If parsed stdout is an array (potentially content array)
                 // or an object with content array (e.g. CallToolResult)
                 if (Array.isArray(inner)) {
                     // Check if elements look like content items
                     const isContentArray = inner.every(item => typeof item === 'object' && item.type);
                     if (isContentArray) {
                        content = inner;
                     }
                 } else if (inner && typeof inner === 'object' && Array.isArray(inner.content)) {
                     content = inner.content;
                 }
             } catch (e) {
                 // stdout is not JSON
             }
        }

        if (isMcpContent(content)) {
            return content;
        }
        return null;
    }, [result]);

    // Attempt to parse a tabular structure from the result
    const tableData = useMemo(() => {
        if (richContent) return null;

        let content = result;

        // 1. Unwrap CallToolResult structure
        if (result && typeof result === 'object' && Array.isArray(result.content)) {
            const textItem = result.content.find((c: any) => c.type === 'text' || (c.text && !c.type));
            if (textItem) {
                content = textItem.text;
            }
        }

        // 2. Parse JSON string if it's a string
        if (typeof content === 'string') {
            try {
                content = JSON.parse(content);
            } catch (e) {
                // Not valid JSON string
                return null;
            }
        }

        // 3. Handle Command Output wrapper (e.g. { stdout: "[...]", ... })
        if (content && typeof content === 'object' && !Array.isArray(content)) {
             if (content.stdout && typeof content.stdout === 'string') {
                 try {
                     const inner = JSON.parse(content.stdout);
                     if (Array.isArray(inner)) {
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

        // 4. Final validation: Must be an array of objects
        if (Array.isArray(content) && content.length > 0) {
            // Check if items are objects
            const isListOfObjects = content.every(item => typeof item === 'object' && item !== null && !Array.isArray(item));
            if (isListOfObjects) {
                return content;
            }
        }

        return null;
    }, [result, richContent]);

    const hasSmartView = tableData !== null || richContent !== null;

    const renderRaw = () => (
        <JsonView data={result} maxHeight={400} />
    );

    const renderSmart = () => {
        if (richContent) {
            return <RichContentRenderer content={richContent} />;
        }

        if (!tableData) return renderRaw();

        // Determine columns from all keys in the first 10 rows (to be reasonably safe)
        const allKeys = new Set<string>();
        tableData.slice(0, 10).forEach(row => {
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
                        {tableData.map((row, idx) => (
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
            {hasSmartView && (
                <div className="flex justify-end mb-1 px-1">
                     <div className="flex items-center bg-muted/50 rounded-lg p-0.5 border">
                         <Button
                            variant={viewMode === "smart" ? "secondary" : "ghost"}
                            size="sm"
                            className="h-6 px-2 text-[10px] gap-1"
                            onClick={() => setViewMode("smart")}
                         >
                             {richContent ? <ImageIcon className="size-3" /> : <TableIcon className="size-3" />}
                             {richContent ? " Preview" : " Table"}
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
