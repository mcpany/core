/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { useState, useMemo } from "react";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Button } from "@/components/ui/button";
import { Code, Table as TableIcon, FileText, ListTree } from "lucide-react";
import { JsonView } from "@/components/ui/json-view";
import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';
import { unwrapMcpResult } from "@/lib/mcp-unwrap";

interface RichResultViewerProps {
    result: unknown;
}

interface TextContent {
    type: 'text';
    text: string;
}

interface ImageContent {
    type: 'image';
    data: string;
    mimeType: string;
}

interface ResourceContent {
    type: 'resource';
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    resource?: any;
}

type McpContent = TextContent | ImageContent | ResourceContent;

function McpContentRenderer({ content }: { content: McpContent[] }) {
    return (
        <div className="space-y-6 p-4">
            {content.map((item, index) => {
                if (item.type === "text") {
                    return (
                        <div key={index} className="prose prose-sm dark:prose-invert max-w-none break-words">
                            <ReactMarkdown remarkPlugins={[remarkGfm]}>
                                {item.text}
                            </ReactMarkdown>
                        </div>
                    );
                } else if (item.type === "image") {
                    return (
                        <div key={index} className="rounded-lg overflow-hidden border bg-muted/20 inline-block max-w-full">
                            <img
                                src={`data:${item.mimeType};base64,${item.data}`}
                                alt="Tool Result Image"
                                className="max-w-full h-auto"
                            />
                        </div>
                    );
                } else if (item.type === "resource") {
                    return (
                        <div key={index} className="flex items-center gap-2 p-3 bg-muted/30 rounded-md border border-white/5">
                            <FileText className="h-4 w-4 text-primary" />
                            <span className="text-sm font-medium">Resource: {item.resource?.uri || 'Unknown'}</span>
                        </div>
                    );
                }
                return null;
            })}
        </div>
    );
}

/**
 * RichResultViewer displays tool execution results in a user-friendly format.
 * It automatically detects if the result contains JSON, tabular data, or MCP content
 * and provides appropriate views with smooth transitions.
 */
export function RichResultViewer({ result }: RichResultViewerProps) {
    const [userViewMode, setUserViewMode] = useState<"smart" | "raw" | "rich" | "tree" | null>(null);

    // 1. Shared unwrapping logic
    const unwrappedContent = useMemo(() => {
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        let content = result as any;

        if (result && typeof result === 'object' && Array.isArray((result as any).content)) {
            content = (result as any).content;
        }

        if (content && typeof content === 'object' && !Array.isArray(content)) {
             if (content.stdout && typeof content.stdout === 'string') {
                 try {
                     const inner = JSON.parse(content.stdout);
                     if (Array.isArray(inner) || (typeof inner === 'object' && inner !== null)) {
                         content = inner;
                     }
                 // eslint-disable-next-line @typescript-eslint/no-unused-vars
                 } catch (e) {
                     // stdout is not JSON
                 }
             }
        }

        if (content && typeof content === 'object' && !Array.isArray(content) && Array.isArray(content.content)) {
            content = content.content;
        }

        return content;
    }, [result]);

    const fullyUnwrapped = useMemo(() => unwrapMcpResult(result), [result]);

    // 2. Identify MCP Content
    const mcpContent = useMemo<McpContent[] | null>(() => {
        if (Array.isArray(unwrappedContent) && unwrappedContent.length > 0) {
            // eslint-disable-next-line @typescript-eslint/no-explicit-any
            const isMcp = unwrappedContent.every((item: any) =>
                typeof item === 'object' && item !== null &&
                (item.type === 'text' || item.type === 'image' || item.type === 'resource')
            );
            if (isMcp) return unwrappedContent as McpContent[];
        }
        return null;
    }, [unwrappedContent]);

    // 3. Identify Table Data
    const tableData = useMemo(() => {
        if (Array.isArray(fullyUnwrapped) && fullyUnwrapped.length > 0) {
             // eslint-disable-next-line @typescript-eslint/no-explicit-any
             const isTable = fullyUnwrapped.every((item: any) => typeof item === 'object' && item !== null);
             const isRichMcp = mcpContent && mcpContent.some(c => c.type !== 'text');
             if (isTable && !isRichMcp) return fullyUnwrapped;
        }

        if (mcpContent) {
             const hasNonText = mcpContent.some(c => c.type !== 'text');
             if (hasNonText) return null;

             if (mcpContent.length === 1 && 'text' in mcpContent[0] && typeof mcpContent[0].text === 'string') {
                 try {
                    const parsed = JSON.parse(mcpContent[0].text);
                    if (Array.isArray(parsed) && parsed.every(item => typeof item === 'object')) {
                        return parsed;
                    }
                // eslint-disable-next-line @typescript-eslint/no-unused-vars
                } catch (e) {}
             }
             return null;
        }

        if (Array.isArray(unwrappedContent) && unwrappedContent.length > 0) {
             // eslint-disable-next-line @typescript-eslint/no-explicit-any
             const isTable = unwrappedContent.every((item: any) => typeof item === 'object' && item !== null);
             if (isTable) return unwrappedContent;
        }

        return null;
    }, [fullyUnwrapped, unwrappedContent, mcpContent]);

    const activeView = useMemo(() => {
        if (userViewMode) return userViewMode;

        if (mcpContent) return 'rich';
        if (tableData) return 'smart';

        // If not array, but object, default to tree
        if (typeof fullyUnwrapped === 'object' && fullyUnwrapped !== null && !Array.isArray(fullyUnwrapped)) {
            return 'tree';
        }

        return 'raw';
    }, [userViewMode, tableData, mcpContent, fullyUnwrapped]);

    const renderRaw = () => (
        <JsonView data={result} maxHeight={400} />
    );

    const renderTree = () => (
        <JsonView data={fullyUnwrapped} maxHeight={400} defaultExpandedLevel={2} />
    );

    const renderRich = () => {
        if (!mcpContent) return renderRaw();

        return (
            <div className="border rounded-md bg-card">
                <ScrollArea className="h-[400px]">
                    <McpContentRenderer content={mcpContent} />
                </ScrollArea>
            </div>
        );
    };

    const renderSmartTable = () => {
        if (!tableData) return null;

        const allKeys = new Set<string>();
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        tableData.slice(0, 10).forEach((row: any) => {
            if (row && typeof row === 'object') {
                Object.keys(row).forEach(k => allKeys.add(k));
            }
        });
        const columns = Array.from(allKeys);

        return (
            <div className="rounded-md border bg-card">
                <ScrollArea className="h-[400px]">
                    <Table>
                        <TableHeader className="bg-muted/50 sticky top-0 z-10">
                            <TableRow>
                                {columns.map(col => (
                                    <TableHead key={col} className="whitespace-nowrap font-medium text-xs px-2 py-1 h-8">
                                        {col}
                                    </TableHead>
                                ))}
                            </TableRow>
                        </TableHeader>
                        <TableBody>
                            {/* eslint-disable-next-line @typescript-eslint/no-explicit-any */}
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
                                            <TableCell key={col} className="px-2 py-2 text-xs max-w-[200px] truncate" title={String(displayVal)}>
                                                {String(displayVal ?? "-")}
                                            </TableCell>
                                        );
                                    })}
                                </TableRow>
                            ))}
                        </TableBody>
                    </Table>
                </ScrollArea>
                <div className="bg-muted/30 px-3 py-2 text-xs text-muted-foreground border-t flex justify-between items-center">
                    <span>Showing {tableData.length} rows</span>
                </div>
            </div>
        );
    };

    return (
        <div className="flex flex-col gap-2 w-full">
            <div className="flex justify-between items-center mb-1">
                 <div className="flex items-center bg-muted/50 rounded-lg p-0.5 border">
                     {mcpContent && (
                        <Button
                            variant={activeView === "rich" ? "secondary" : "ghost"}
                            size="sm"
                            className="h-7 px-3 text-xs gap-1.5"
                            onClick={() => setUserViewMode("rich")}
                        >
                            <FileText className="size-3.5" /> Rendered
                        </Button>
                     )}
                     {tableData && (
                        <Button
                            variant={activeView === "smart" ? "secondary" : "ghost"}
                            size="sm"
                            className="h-7 px-3 text-xs gap-1.5"
                            onClick={() => setUserViewMode("smart")}
                        >
                            <TableIcon className="size-3.5" /> Table
                        </Button>
                     )}
                     <Button
                        variant={activeView === "tree" ? "secondary" : "ghost"}
                        size="sm"
                        className="h-7 px-3 text-xs gap-1.5"
                        onClick={() => setUserViewMode("tree")}
                     >
                         <ListTree className="size-3.5" /> JSON
                     </Button>
                     <Button
                        variant={activeView === "raw" ? "secondary" : "ghost"}
                        size="sm"
                        className="h-7 px-3 text-xs gap-1.5"
                        onClick={() => setUserViewMode("raw")}
                     >
                         <Code className="size-3.5" /> Raw
                     </Button>
                 </div>
            </div>

            <div className="mt-0">
                {activeView === 'smart' && renderSmartTable()}
                {activeView === 'rich' && renderRich()}
                {activeView === 'tree' && renderTree()}
                {activeView === 'raw' && renderRaw()}
            </div>
        </div>
    );
}
