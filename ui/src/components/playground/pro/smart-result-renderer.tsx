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

const isMcpContent = (content: any): boolean => {
    return Array.isArray(content) && content.length > 0 && content.every((item: any) =>
        typeof item === 'object' && item !== null &&
        (item.type === 'text' || item.type === 'image' || item.type === 'resource')
    );
};

const RichContentRenderer = ({ content }: { content: any[] }) => {
    return (
        <div className="flex flex-col gap-4 p-2 bg-muted/10 rounded-lg border border-border/50">
            {content.map((item, idx) => {
                if (item.type === 'image') {
                    return (
                        <div key={idx} className="flex flex-col gap-1">
                            <div className="text-[10px] uppercase text-muted-foreground font-semibold tracking-wider flex items-center gap-1">
                                <ImageIcon className="w-3 h-3" /> Image ({item.mimeType})
                            </div>
                            <div className="border rounded-md overflow-hidden bg-black/20 self-start">
                                {/* eslint-disable-next-line @next/next/no-img-element */}
                                <img
                                    src={`data:${item.mimeType};base64,${item.data}`}
                                    alt="Tool Result"
                                    className="max-w-full h-auto object-contain max-h-[500px]"
                                />
                            </div>
                        </div>
                    );
                }
                if (item.type === 'text') {
                     return (
                         <div key={idx} className="flex flex-col gap-1 w-full overflow-hidden">
                             {content.length > 1 && (
                                <div className="text-[10px] uppercase text-muted-foreground font-semibold tracking-wider">
                                    Text Output
                                </div>
                             )}
                             <div className="whitespace-pre-wrap font-mono text-xs bg-muted/30 p-3 rounded-md border border-white/5 overflow-x-auto">
                                 {item.text}
                             </div>
                         </div>
                     );
                }
                return (
                    <div key={idx} className="text-xs text-muted-foreground italic p-2 border border-dashed rounded">
                        Unsupported content type: {item.type}
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

    // Check for MCP Content (Images, Mixed) first
    const smartContent = useMemo(() => {
        let content = result;

        // 1. Check if it's already CallToolResult structure
         if (result && typeof result === 'object' && Array.isArray(result.content)) {
            // Check if it has images
            if (result.content.some((c:any) => c.type === 'image')) {
                 return result.content;
            }
             // If mixed content, return it
            if (result.content.length > 1) {
                return result.content;
            }
         }

         // 2. Check Command Output (nested JSON content)
         if (content && typeof content === 'object' && !Array.isArray(content)) {
             if (content.stdout && typeof content.stdout === 'string') {
                 try {
                     const inner = JSON.parse(content.stdout);
                     // Check if inner is MCP Content Array containing images
                     if (isMcpContent(inner)) {
                          if (inner.some((c:any) => c.type === 'image')) {
                              return inner;
                          }
                     }
                 } catch {
                    // ignore parse error
                 }
             }
         }

         return null;
    }, [result]);

    // Attempt to parse a tabular structure from the result (Fallback for text/data)
    const tableData = useMemo(() => {
        // If we found smart content (images), skip table parsing
        if (smartContent) return null;

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
            } catch {
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
                 } catch {
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
    }, [result, smartContent]);

    const hasSmartView = tableData !== null || smartContent !== null;

    const renderRaw = () => (
        <JsonView data={result} maxHeight={400} />
    );

    const renderSmart = () => {
        if (smartContent) {
            return <RichContentRenderer content={smartContent} />;
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
                             {smartContent ? <ImageIcon className="size-3" /> : <TableIcon className="size-3" />}
                             {smartContent ? " Visual" : " Table"}
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
