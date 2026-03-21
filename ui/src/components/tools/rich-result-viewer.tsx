/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */



import { useMemo } from "react";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { ScrollArea } from "@/components/ui/scroll-area";
import { FileJson, Table as TableIcon, Terminal, FileText, Maximize2 } from "lucide-react";
import { JsonView } from "@/components/ui/json-view";
import { JsonTree } from "@/components/ui/json-tree";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";
import { Button } from "@/components/ui/button";
import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';

interface RichResultViewerProps {
    result: unknown;
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
        if (Array.isArray(content)) {
            const isValidArray = content.every((item: Record<string, unknown>) =>
                (item.type === 'text' && typeof item.text === 'string') ||
                (item.type === 'image' && typeof item.data === 'string' && typeof item.mimeType === 'string')
            );
            if (isValidArray) {
                return content as unknown as McpContent[];
            }
        }

        if (content && typeof content === 'object' && 'content' in content && Array.isArray((content as Record<string, unknown>).content)) {
            const contentArray = (content as Record<string, unknown>).content as Record<string, unknown>[];
            // Check if it looks like MCP content
            const isValid = contentArray.every((item: Record<string, unknown>) =>
                (item.type === 'text' && typeof item.text === 'string') ||
                (item.type === 'image' && typeof item.data === 'string' && typeof item.mimeType === 'string')
            );
            if (isValid) {
                return contentArray as unknown as McpContent[];
            }
        }
        return null;
    }, [content]);

    const isTableEligible = useMemo(() => {
        return !mcpContent && Array.isArray(content) && content.length > 0 && typeof content[0] === 'object' && content[0] !== null;
    }, [content, mcpContent]);

    // Get columns for table
    const columns = useMemo(() => {
        if (!isTableEligible) return [];
        // aggregate all keys from all objects to handle sparse data
        const keys = new Set<string>();
        // Limit rows scanned for columns to avoid perf issues on huge datasets
        const items = content as Record<string, unknown>[];
        items.slice(0, 50).forEach((item: Record<string, unknown>) => {
            if (typeof item === 'object' && item !== null) {
                Object.keys(item).forEach(k => keys.add(k));
            }
        });
        return Array.from(keys);
    }, [content, isTableEligible]);

    const renderCell = (value: unknown) => {
        if (value === null || value === undefined) return <span className="text-muted-foreground">-</span>;
        if (typeof value === 'object') {
             return (
                 <Popover>
                     <PopoverTrigger asChild>
                         <Button variant="ghost" size="sm" className="h-6 px-2 text-[10px] font-mono text-muted-foreground truncate max-w-[150px] justify-start bg-muted/20">
                             <Maximize2 className="h-3 w-3 mr-1 shrink-0" />
                             {Array.isArray(value) ? `Array(${value.length})` : `Object({${Object.keys(value).length}})`}
                         </Button>
                     </PopoverTrigger>
                     <PopoverContent className="w-80 p-0" align="start">
                         <div className="max-h-[300px] overflow-auto p-2 bg-[#1e1e1e]">
                             <JsonTree data={value} defaultExpandedLevel={1} />
                         </div>
                     </PopoverContent>
                 </Popover>
             );
        }
        if (typeof value === 'boolean') return <span className={value ? "text-green-500 font-medium" : "text-red-500 font-medium"}>{String(value)}</span>;
        return <span className="truncate max-w-[300px] block" title={String(value)}>{String(value)}</span>;
    }

    const defaultTab = mcpContent ? "rendered" : (isTableEligible ? "table" : "json");

    return (
        <Tabs defaultValue={defaultTab} className="w-full">
            <div className="flex items-center justify-between mb-2">
                <TabsList>
                    {mcpContent && (
                        <TabsTrigger value="rendered" className="flex items-center gap-2">
                            <FileText className="h-4 w-4" /> Rendered
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
                                {(content as Record<string, unknown>[]).map((row: Record<string, unknown>, i: number) => (
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
                <JsonView data={content} maxHeight={400} defaultExpandedLevel={2} />
            </TabsContent>

            {isExtracted && (
                <TabsContent value="raw">
                    <JsonView data={result} maxHeight={400} />
                </TabsContent>
            )}
        </Tabs>
    );
}
