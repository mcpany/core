/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useMemo } from "react";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Button } from "@/components/ui/button";
import { Code, Table as TableIcon, LayoutList } from "lucide-react";
import { JsonView } from "@/components/ui/json-view";

/**
 * Props for the SmartResultRenderer component.
 */
interface SmartResultRendererProps {
    /** The result object to render. Can be a JSON string, an object, or an array. */
    result: any;
}

/**
 * Content item structure (MCP).
 */
interface ContentItem {
    type?: string;
    text?: string;
    data?: string;
    mimeType?: string;
}

/**
 * Renders mixed content (Text and Images).
 */
function RichContentRenderer({ content }: { content: ContentItem[] }) {
    return (
        <div className="flex flex-col gap-4 p-4 border rounded-md bg-muted/10">
            {content.map((item, index) => {
                if (item.type === 'image' && item.data && item.mimeType) {
                    return (
                        <div key={index} className="flex flex-col gap-2">
                            <span className="text-xs text-muted-foreground font-mono uppercase tracking-wider">Image ({item.mimeType})</span>
                            <div className="border rounded-lg overflow-hidden bg-[url('/checkerboard.svg')] bg-repeat">
                                <img
                                    src={`data:${item.mimeType};base64,${item.data}`}
                                    alt="Tool Result"
                                    className="max-w-full h-auto object-contain max-h-[500px]"
                                />
                            </div>
                        </div>
                    );
                }

                if (item.type === 'text' || (item.text && !item.type)) {
                    return (
                        <div key={index} className="flex flex-col gap-2">
                            <span className="text-xs text-muted-foreground font-mono uppercase tracking-wider">Text Output</span>
                            <pre className="whitespace-pre-wrap text-sm font-mono bg-muted/50 p-3 rounded-md overflow-auto max-h-[300px]">
                                {item.text}
                            </pre>
                        </div>
                    );
                }

                return null;
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

    // Attempt to parse a tabular structure or rich content from the result
    const { tableData, richContent } = useMemo(() => {
        let current = result;
        let contentArray: ContentItem[] | null = null;

        // 1. Handle Command Output wrapper (e.g. { stdout: "[...]", ... })
        // Check if current is object with stdout string
        if (current && typeof current === 'object' && !Array.isArray(current)) {
             if (current.stdout && typeof current.stdout === 'string') {
                 try {
                     const parsed = JSON.parse(current.stdout);
                     // If parsed is valid, replace current with it (unwrap)
                     if (parsed) current = parsed;
                 } catch (e) {
                     // stdout is not JSON, but maybe it's just text content we can wrap?
                     // No, if it's plain text, we treat it as raw string later.
                 }
             }
        }

        // 2. Check for MCP Content (CallToolResult)
        // Structure: { content: [{ type: ... }, ...] }
        if (current && typeof current === 'object' && Array.isArray(current.content)) {
            contentArray = current.content;
        } else if (Array.isArray(current) && current.length > 0 &&
                  (current[0].type === 'text' || current[0].type === 'image')) {
             // Maybe current IS the content array (e.g. from stdout parsing)
             contentArray = current;
        }

        if (contentArray) {
             const hasImage = contentArray.some((c: any) => c.type === 'image');
             const hasMultiple = contentArray.length > 1;

             // If images or multiple items, prefer Rich View
             if (hasImage || hasMultiple) {
                 return { tableData: null, richContent: contentArray };
             }

             // Single text item -> Unwrap for Table logic attempt
             const textItem = contentArray.find((c: any) => c.type === 'text' || (c.text && !c.type));
             if (textItem && textItem.text) {
                 current = textItem.text;
             }
        }

        // 3. Parse JSON string (if current is string)
        if (typeof current === 'string') {
             try {
                 current = JSON.parse(current);
             } catch (e) {
                 // Not valid JSON string, return null
                 return { tableData: null, richContent: null };
             }
        }

        // 4. Final validation: Must be an array of objects
        if (Array.isArray(current) && current.length > 0) {
            // Check if items are objects
            const isListOfObjects = current.every((item: any) => typeof item === 'object' && item !== null && !Array.isArray(item));
            if (isListOfObjects) {
                return { tableData: current, richContent: null };
            }
        }

        return { tableData: null, richContent: null };
    }, [result]);

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
            {hasSmartView && (
                <div className="flex justify-end mb-1 px-1">
                     <div className="flex items-center bg-muted/50 rounded-lg p-0.5 border">
                         <Button
                            variant={viewMode === "smart" ? "secondary" : "ghost"}
                            size="sm"
                            className="h-6 px-2 text-[10px] gap-1"
                            onClick={() => setViewMode("smart")}
                         >
                             {richContent ? <LayoutList className="size-3" /> : <TableIcon className="size-3" />}
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
