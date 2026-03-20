/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */



import { useMemo, useState } from "react";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { ScrollArea } from "@/components/ui/scroll-area";
import { FileJson, Table as TableIcon, Terminal, FileText, ChevronDown, ChevronsUpDown, Search } from "lucide-react";
import { JsonView } from "@/components/ui/json-view";
import { Input } from "@/components/ui/input";
import { cn } from "@/lib/utils";
import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';

interface RichResultViewerProps {
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
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
        if (Array.isArray(content)) {
            // eslint-disable-next-line @typescript-eslint/no-explicit-any
            const isValidArray = content.every((item: any) =>
                (item.type === 'text' && typeof item.text === 'string') ||
                (item.type === 'image' && typeof item.data === 'string' && typeof item.mimeType === 'string')
            );
            if (isValidArray) {
                return content as McpContent[];
            }
        }

        if (content && typeof content === 'object' && Array.isArray(content.content)) {
            // Check if it looks like MCP content
            // eslint-disable-next-line @typescript-eslint/no-explicit-any
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

    const isTableEligible = useMemo(() => {
        return !mcpContent && Array.isArray(content) && content.length > 0 && typeof content[0] === 'object' && content[0] !== null;
    }, [content, mcpContent]);

    const [searchQuery, setSearchQuery] = useState("");
    const [sortCol, setSortCol] = useState<string | null>(null);
    const [sortDirection, setSortDirection] = useState<'asc' | 'desc'>('asc');

    // Get columns for table
    const columns = useMemo(() => {
        if (!isTableEligible) return [];
        // aggregate all keys from all objects to handle sparse data
        const keys = new Set<string>();
        // Limit rows scanned for columns to avoid perf issues on huge datasets
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        content.slice(0, 50).forEach((item: any) => {
            if (typeof item === 'object' && item !== null) {
                Object.keys(item as Record<string, unknown>).forEach(k => keys.add(k));
            }
        });
        return Array.from(keys);
    }, [content, isTableEligible]);

    // Filter and sort the table content
    const tableData = useMemo(() => {
        if (!isTableEligible) return [];
        let data = [...content];

        // Apply search filter
        if (searchQuery) {
            const query = searchQuery.toLowerCase();
            // eslint-disable-next-line @typescript-eslint/no-explicit-any
            data = data.filter((row: any) => {
                return Object.values(row as Record<string, unknown>).some((val) =>
                    String(val).toLowerCase().includes(query)
                );
            });
        }

        // Apply sorting
        if (sortCol) {
            // eslint-disable-next-line @typescript-eslint/no-explicit-any
            data.sort((a: any, b: any) => {
                const aVal = a[sortCol];
                const bVal = b[sortCol];

                if (aVal === bVal) return 0;
                if (aVal === null || aVal === undefined) return sortDirection === 'asc' ? -1 : 1;
                if (bVal === null || bVal === undefined) return sortDirection === 'asc' ? 1 : -1;

                const aStr = String(aVal).toLowerCase();
                const bStr = String(bVal).toLowerCase();

                if (aStr < bStr) return sortDirection === 'asc' ? -1 : 1;
                if (aStr > bStr) return sortDirection === 'asc' ? 1 : -1;
                return 0;
            });
        }

        return data;
    }, [content, isTableEligible, searchQuery, sortCol, sortDirection]);

    const handleSort = (col: string) => {
        if (sortCol === col) {
            setSortDirection(prev => prev === 'asc' ? 'desc' : 'asc');
        } else {
            setSortCol(col);
            setSortDirection('asc');
        }
    };

    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const renderCell = (value: any) => {
        if (value === null || value === undefined) return <span className="text-muted-foreground">-</span>;
        if (typeof value === 'object') return <span className="font-mono text-xs text-muted-foreground truncate max-w-[200px] block" title={JSON.stringify(value)}>{JSON.stringify(value)}</span>;
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
                <TabsContent value="table" className="border rounded-md bg-background/50 backdrop-blur-sm flex flex-col h-[400px]">
                    <div className="p-2 border-b">
                        <div className="relative">
                            <Search className="absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" />
                            <Input
                                placeholder="Search table..."
                                className="pl-8 bg-background/50 h-9"
                                value={searchQuery}
                                onChange={(e) => setSearchQuery(e.target.value)}
                            />
                        </div>
                    </div>
                    <ScrollArea className="flex-1">
                        <Table>
                            <TableHeader className="sticky top-0 bg-background/95 backdrop-blur z-10 border-b">
                                <TableRow>
                                    {columns.map(col => (
                                        <TableHead
                                            key={col}
                                            className="whitespace-nowrap cursor-pointer hover:bg-muted/50 transition-colors"
                                            onClick={() => handleSort(col)}
                                        >
                                            <div className="flex items-center gap-2">
                                                {col}
                                                {sortCol === col ? (
                                                    <ChevronDown className={cn("h-3 w-3 transition-transform", sortDirection === 'desc' ? "" : "rotate-180")} />
                                                ) : (
                                                    <ChevronsUpDown className="h-3 w-3 opacity-50" />
                                                )}
                                            </div>
                                        </TableHead>
                                    ))}
                                </TableRow>
                            </TableHeader>
                            <TableBody>
                                {tableData.length === 0 ? (
                                    <TableRow>
                                        <TableCell colSpan={columns.length} className="h-24 text-center">
                                            <div className="flex flex-col items-center justify-center text-muted-foreground">
                                                <p className="text-base font-medium">No results found</p>
                                                <p className="text-sm opacity-70">Try adjusting your search query.</p>
                                            </div>
                                        </TableCell>
                                    </TableRow>
                                ) : (
                                    // eslint-disable-next-line @typescript-eslint/no-explicit-any
                                    tableData.map((row: any, i: number) => (
                                        <TableRow key={i}>
                                            {columns.map(col => (
                                                <TableCell key={col} className="py-2">
                                                    {renderCell(row[col])}
                                                </TableCell>
                                            ))}
                                        </TableRow>
                                    ))
                                )}
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
