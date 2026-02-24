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
import { FileJson, Table as TableIcon, Terminal, FileText, Image as ImageIcon } from "lucide-react";
import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';

interface RichResultViewerProps {
    result: any;
}

interface TextContent {
    type: "text";
    text: string;
}

interface ImageContent {
    type: "image";
    data: string;
    mimeType: string;
}

type McpContent = TextContent | ImageContent;

interface McpContentRendererProps {
    content: McpContent[];
}

function McpContentRenderer({ content }: McpContentRendererProps) {
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
                }
                return null;
            })}
        </div>
    );
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

    const mcpContent = useMemo<McpContent[] | null>(() => {
        if (content && typeof content === 'object' && Array.isArray(content.content)) {
            // Check if it looks like MCP content
            const isValid = content.content.every((item: any) =>
                (item.type === 'text' && typeof item.text === 'string') ||
                (item.type === 'image' && typeof item.data === 'string' && typeof item.mimeType === 'string')
            );
            if (isValid) {
                return content.content;
            }
        }
        return null;
    }, [content]);

    const isImageEligible = useMemo(() => {
        return Array.isArray(content) && content.length === 1 && typeof content[0] === 'object' && content[0] !== null && content[0].type === 'image' && !!content[0].data;
    }, [content]);

    const isTableEligible = useMemo(() => {
        return !mcpContent && !isImageEligible && Array.isArray(content) && content.length > 0 && typeof content[0] === 'object' && content[0] !== null;
    }, [content, mcpContent, isImageEligible]);

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

    const defaultTab = mcpContent ? "rendered" : (isImageEligible ? "image" : (isTableEligible ? "table" : "json"));

    return (
        <Tabs defaultValue={defaultTab} className="w-full">
            <div className="flex items-center justify-between mb-2">
                <TabsList>
                    {mcpContent && (
                        <TabsTrigger value="rendered" className="flex items-center gap-2">
                            <FileText className="h-4 w-4" /> Rendered
                        </TabsTrigger>
                    )}
                    {isImageEligible && (
                        <TabsTrigger value="image" className="flex items-center gap-2">
                            <ImageIcon className="h-4 w-4" /> Image
                        </TabsTrigger>
                    )}
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

            {mcpContent && (
                <TabsContent value="rendered" className="border rounded-md bg-card">
                    <ScrollArea className="h-[400px]">
                        <McpContentRenderer content={mcpContent} />
                    </ScrollArea>
                </TabsContent>
            )}

            {isImageEligible && (
                <TabsContent value="image" className="border rounded-md p-4 bg-muted/20 flex justify-center">
                    <img
                        src={`data:${content[0].mimeType || 'image/png'};base64,${content[0].data}`}
                        alt="Tool Result"
                        className="max-w-full max-h-[500px] object-contain rounded-md shadow-sm border"
                    />
                </TabsContent>
            )}

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
