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
    resource?: any;
    [key: string]: any;
}

function isMcpContent(item: any): item is McpContent {
    return typeof item === 'object' && item !== null && (item.type === 'text' || item.type === 'image' || item.type === 'resource');
}

function RichContentRenderer({ content }: { content: McpContent[] }) {
    return (
        <div className="flex flex-col gap-4">
            {content.map((item, index) => {
                if (item.type === 'image' && item.data) {
                     const mimeType = item.mimeType || 'image/png';
                     return (
                         <div key={index} className="rounded-lg overflow-hidden border bg-black/5 inline-block">
                             <img
                                 src={`data:${mimeType};base64,${item.data}`}
                                 alt="Tool Result"
                                 className="max-w-full h-auto object-contain"
                             />
                         </div>
                     );
                }
                if (item.type === 'text' && item.text) {
                     let isJson = false;
                     let jsonObj = null;
                     const trimmed = item.text.trim();
                     if ((trimmed.startsWith('{') && trimmed.endsWith('}')) || (trimmed.startsWith('[') && trimmed.endsWith(']'))) {
                         try {
                             jsonObj = JSON.parse(trimmed);
                             isJson = true;
                         } catch {}
                     }

                     if (isJson) {
                         return <JsonView key={index} data={jsonObj} />;
                     }
                     return (
                         <div key={index} className="whitespace-pre-wrap font-mono text-sm bg-muted/20 p-2 rounded">
                             {item.text}
                         </div>
                     );
                }
                if (item.type === 'resource' && item.resource) {
                    return (
                        <div key={index} className="bg-muted/20 p-2 rounded text-sm">
                            <div className="font-semibold text-xs uppercase text-muted-foreground mb-1">Resource</div>
                             <JsonView data={item.resource} />
                        </div>
                    );
                }
                return (
                    <div key={index} className="text-muted-foreground text-xs italic">
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

    const richContent = useMemo(() => {
        let content: any = result;
        // Unwrap logic similar to tableData but looking for array of McpContent
        if (result && typeof result === 'object') {
             if (Array.isArray(result.content)) {
                 content = result.content;
             } else if (result.stdout && typeof result.stdout === 'string') {
                  try {
                      const inner = JSON.parse(result.stdout);
                      if (Array.isArray(inner) && inner.some(isMcpContent)) {
                          content = inner;
                      } else if (typeof inner === 'object' && Array.isArray(inner.content)) {
                          // Handle { content: [...] } inside stdout
                          content = inner.content;
                      }
                  } catch {}
             }
        }

        if (Array.isArray(content) && content.length > 0 && content.some(isMcpContent)) {
            // If it looks like MCP content, verify if it has images or resources, OR multiple items.
            // If it's a single TEXT item, we might prefer the Table view if that text is JSON.
            const hasRichItems = content.some(c => c.type === 'image' || c.type === 'resource');
            if (hasRichItems || content.length > 1) {
                return content as McpContent[];
            }
        }
        return null;
    }, [result]);

    // Attempt to parse a tabular structure from the result
    const tableData = useMemo(() => {
        if (richContent) return null; // Prefer rich content if available

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
                             {richContent ? "Visual" : "Table"}
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
