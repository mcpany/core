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

interface McpContent {
    type: string;
    text?: string;
    data?: string;
    mimeType?: string;
    resource?: any;
}

/**
 * Props for the SmartResultRenderer component.
 */
interface SmartResultRendererProps {
    /** The result object to render. Can be a JSON string, an object, or an array. */
    result: any;
}

function RichContentRenderer({ content }: { content: McpContent[] }) {
    return (
        <div className="flex flex-col gap-4">
            {content.map((item, idx) => {
                if (item.type === 'image' && item.data) {
                    const mime = item.mimeType || 'image/png';
                    const src = `data:${mime};base64,${item.data}`;
                    return (
                        <div key={idx} className="border rounded-lg overflow-hidden bg-muted/20">
                            <div className="flex items-center gap-2 px-3 py-2 bg-muted/50 border-b text-xs text-muted-foreground">
                                <ImageIcon className="size-3" />
                                <span>Image ({mime})</span>
                            </div>
                            <div className="p-4 flex justify-center">
                                <img src={src} alt="Tool Result" className="max-w-full h-auto rounded shadow-sm" />
                            </div>
                        </div>
                    );
                }
                if (item.type === 'text' && item.text) {
                    return (
                        <div key={idx} className="border rounded-lg overflow-hidden bg-muted/20">
                            <div className="flex items-center gap-2 px-3 py-2 bg-muted/50 border-b text-xs text-muted-foreground">
                                <FileText className="size-3" />
                                <span>Text Output</span>
                            </div>
                            <div className="p-4 font-mono text-sm whitespace-pre-wrap break-words">
                                {item.text}
                            </div>
                        </div>
                    );
                }
                return (
                    <div key={idx} className="p-2 border rounded text-xs text-muted-foreground">
                        Unsupported content type: {item.type}
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

    const smartContent = useMemo(() => {
         let content = result;
         // 1. Unwrap CallToolResult structure
         if (result && typeof result === 'object' && Array.isArray(result.content)) {
             return result.content as McpContent[];
         }

         // 2. Handle Command Output wrapper (e.g. { stdout: "[...]", ... })
         if (content && typeof content === 'object' && !Array.isArray(content)) {
             if (content.stdout && typeof content.stdout === 'string') {
                 try {
                     const inner = JSON.parse(content.stdout);
                     // Check if inner is CallToolResult-like content array
                     if (Array.isArray(inner) && inner.length > 0 && inner[0].type) {
                        return inner as McpContent[];
                     }
                      // Or if inner is object with content array
                     if (inner && typeof inner === 'object' && Array.isArray(inner.content)) {
                         return inner.content as McpContent[];
                     }
                 } catch (e) {
                     // stdout is not JSON
                 }
             }
         }
         return null;
    }, [result]);

    const isRichContent = useMemo(() => {
        if (!smartContent) return false;
        // If it has images, it's rich content
        if (smartContent.some(c => c.type === 'image')) return true;
        // If it has mixed content (more than 1 item), treat as rich content
        if (smartContent.length > 1) return true;
        // If single text item, we prefer Table/JSON unless it's explicitly explicitly requested?
        // Actually, existing logic prefers Table if the text parses to JSON array.
        return false;
    }, [smartContent]);


    // Attempt to parse a tabular structure from the result
    const tableData = useMemo(() => {
        if (isRichContent) return null; // Don't try table if rich content

        let content = result;

        // 1. Unwrap CallToolResult structure (find text)
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

        // 3. Handle Command Output wrapper
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
    }, [result, isRichContent]);

    const hasSmartView = tableData !== null || isRichContent;

    const renderRaw = () => (
        <JsonView data={result} maxHeight={400} />
    );

    const renderSmart = () => {
        if (isRichContent && smartContent) {
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
                             {isRichContent ? <ImageIcon className="size-3" /> : <TableIcon className="size-3" />}
                             {isRichContent ? " Preview" : " Table"}
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
