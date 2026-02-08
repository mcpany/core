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

const isMcpContent = (data: any): boolean => {
    if (!Array.isArray(data)) return false;
    return data.some(item =>
        typeof item === 'object' &&
        item !== null &&
        (item.type === 'text' || item.type === 'image' || item.type === 'resource')
    );
};

const RichContentRenderer = ({ content }: { content: any[] }) => {
    return (
        <div className="flex flex-col gap-4">
            {content.map((item, idx) => {
                if (item.type === 'image') {
                    const src = `data:${item.mimeType};base64,${item.data}`;
                    return (
                        <div key={idx} className="border rounded-md overflow-hidden bg-accent/20 border-dashed p-2 flex justify-center">
                            <img src={src} alt="Tool Result" className="max-w-full h-auto rounded shadow-sm" />
                        </div>
                    );
                }
                if (item.type === 'text') {
                    return (
                        <div key={idx} className="bg-muted/30 p-3 rounded-md border text-sm whitespace-pre-wrap break-words font-mono">
                            {item.text}
                        </div>
                    );
                }
                if (item.type === 'resource') {
                    // Fallback for resource type (maybe text or blob)
                    return (
                         <div key={idx} className="bg-muted/30 p-3 rounded-md border text-sm">
                             <div className="flex items-center gap-2 mb-2 text-muted-foreground">
                                 <FileText className="size-4" />
                                 <span className="font-medium">Resource: {item.resource?.uri}</span>
                             </div>
                             {item.resource?.text && (
                                <div className="whitespace-pre-wrap break-words font-mono text-xs opacity-80">
                                    {item.resource.text}
                                </div>
                             )}
                         </div>
                    );
                }
                return (
                    <div key={idx} className="bg-destructive/10 text-destructive p-2 rounded text-xs">
                        Unknown content type: {item.type}
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

    // Attempt to parse a tabular structure from the result
    const { tableData, richContent } = useMemo(() => {
        let content = result;
        // eslint-disable-next-line @typescript-eslint/no-unused-vars
        let richContent = null;

        // 1. Check if it's already a CallToolResult with content
        if (result && typeof result === 'object' && Array.isArray(result.content)) {
             // Check if we should render as Rich Content (images or mixed)
            const hasImage = result.content.some((c: any) => c.type === 'image');
            const hasMultiple = result.content.length > 1;

            if (hasImage || hasMultiple) {
                return { tableData: null, richContent: result.content };
            }

             // Otherwise, if single text item, unwrap it for potential Table parsing
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
                // Not valid JSON string, leave as is
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

        // 4. Check if final content is MCP Content (e.g. from stdout)
        if (Array.isArray(content) && isMcpContent(content)) {
             return { tableData: null, richContent: content };
        }

        // 5. Final validation: Must be an array of objects
        if (Array.isArray(content) && content.length > 0) {
            // Check if items are objects
            const isListOfObjects = content.every(item => typeof item === 'object' && item !== null && !Array.isArray(item));
            if (isListOfObjects) {
                return { tableData: content, richContent: null };
            }
        }

        return { tableData: null, richContent: null };
    }, [result]);

    const hasSmartView = tableData !== null || richContent !== null;

    const renderRaw = () => (
        <JsonView data={result} maxHeight={400} />
    );

    const renderSmart = () => {
        if (richContent) return <RichContentRenderer content={richContent} />;
        if (!tableData) return renderRaw();

        // Determine columns from all keys in the first 10 rows (to be reasonably safe)
        const allKeys = new Set<string>();
        tableData.slice(0, 10).forEach((row: any) => {
            Object.keys(row).forEach((k: string) => allKeys.add(k));
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
                             {richContent ? <ImageIcon className="size-3" /> : <TableIcon className="size-3" />}
                             {richContent ? "Preview" : "Table"}
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
