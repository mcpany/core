/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */



import { useMemo, useState } from "react";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { ScrollArea } from "@/components/ui/scroll-area";
import { FileJson, Table as TableIcon, Terminal, FileText, Download, ArrowDown, ArrowUp, ArrowUpDown } from "lucide-react";
import { JsonView } from "@/components/ui/json-view";
import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';
import { Button } from "@/components/ui/button";
import { useToast } from "@/hooks/use-toast";

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
        if (Array.isArray(content)) {
            const isValidArray = content.every((item: unknown) => {
                if (typeof item !== 'object' || item === null) return false;
                const i = item as Record<string, unknown>;
                return (i.type === 'text' && typeof i.text === 'string') ||
                       (i.type === 'image' && typeof i.data === 'string' && typeof i.mimeType === 'string');
            });
            if (isValidArray) {
                return content as McpContent[];
            }
        }

        if (content && typeof content === 'object' && 'content' in content && Array.isArray((content as Record<string, unknown>).content)) {
            const cArr = (content as Record<string, unknown>).content as unknown[];
            const isValid = cArr.every((item: unknown) => {
                if (typeof item !== 'object' || item === null) return false;
                const i = item as Record<string, unknown>;
                return (i.type === 'text' && typeof i.text === 'string') ||
                       (i.type === 'image' && typeof i.data === 'string' && typeof i.mimeType === 'string');
            });
            if (isValid) {
                return cArr as McpContent[];
            }
        }
        return null;
    }, [content]);

    const isTableEligible = useMemo(() => {
        // If content is pure array of objects, it's eligible
        if (Array.isArray(content) && content.length > 0 && typeof content[0] === 'object' && content[0] !== null) {
            return true;
        }

        // If it's a standard MCP response payload: { content: [{ type: "text", text: "[{...}]" }] }
        if (mcpContent && mcpContent.length === 1 && mcpContent[0].type === 'text') {
            try {
                const parsed = JSON.parse(mcpContent[0].text);
                if (Array.isArray(parsed) && parsed.length > 0 && typeof parsed[0] === 'object' && parsed[0] !== null) {
                    return true;
                }
            } catch {
                // Not JSON or not array
            }
        }

        return false;
    }, [content, mcpContent]);

    // Derived table data (extracts from MCP string if necessary)
    const tableData = useMemo(() => {
        if (!isTableEligible) return [];
        if (Array.isArray(content)) return content as Record<string, unknown>[];
        if (mcpContent && mcpContent.length === 1 && mcpContent[0].type === 'text') {
            try {
                const parsed = JSON.parse(mcpContent[0].text);
                if (Array.isArray(parsed)) return parsed as Record<string, unknown>[];
            } catch {
                return [];
            }
        }
        return [];
    }, [content, isTableEligible, mcpContent]);

    const { toast } = useToast();
    const [sortColumn, setSortColumn] = useState<string | null>(null);
    const [sortDirection, setSortDirection] = useState<'asc' | 'desc'>('asc');

    // Get columns for table
    const columns = useMemo(() => {
        if (!isTableEligible) return [];
        // aggregate all keys from all objects to handle sparse data
        const keys = new Set<string>();
        // Limit rows scanned for columns to avoid perf issues on huge datasets
        tableData.slice(0, 50).forEach((item) => {
            if (typeof item === 'object' && item !== null) {
                Object.keys(item).forEach(k => keys.add(k));
            }
        });
        return Array.from(keys);
    }, [tableData, isTableEligible]);

    const handleSort = (column: string) => {
        if (sortColumn === column) {
            setSortDirection(sortDirection === 'asc' ? 'desc' : 'asc');
        } else {
            setSortColumn(column);
            setSortDirection('asc');
        }
    };

    const sortedContent = useMemo(() => {
        if (!isTableEligible || !sortColumn) return tableData;

        const sorted = [...tableData].sort((a, b) => {
            const aVal = a[sortColumn];
            const bVal = b[sortColumn];

            if (aVal === bVal) return 0;
            if (aVal === null || aVal === undefined) return 1;
            if (bVal === null || bVal === undefined) return -1;

            if (typeof aVal === 'string' && typeof bVal === 'string') {
                return sortDirection === 'asc' ? aVal.localeCompare(bVal) : bVal.localeCompare(aVal);
            }
            if (typeof aVal === 'number' && typeof bVal === 'number') {
                return sortDirection === 'asc' ? aVal - bVal : bVal - aVal;
            }

            // Fallback for objects, booleans, etc. Stringify then compare
            const aStr = String(aVal);
            const bStr = String(bVal);
            return sortDirection === 'asc' ? aStr.localeCompare(bStr) : bStr.localeCompare(aStr);
        });
        return sorted;
    }, [tableData, sortColumn, sortDirection, isTableEligible]);

    const handleExportCSV = () => {
        if (!isTableEligible) return;

        // Escape helper for CSV cells
        const escapeCSV = (val: unknown) => {
            if (val === null || val === undefined) return '""';
            const str = typeof val === 'object' ? JSON.stringify(val) : String(val);
            // Replace double quotes with double-double quotes, and wrap in double quotes
            return `"${str.replace(/"/g, '""')}"`;
        };

        const headers = columns.map(escapeCSV).join(',');

        const rows = sortedContent.map((row) =>
            columns.map(col => escapeCSV(row[col])).join(',')
        ).join('\n');

        const csvContent = headers + '\n' + rows;
        const blob = new Blob([csvContent], { type: 'text/csv;charset=utf-8;' });
        const url = URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.href = url;
        a.download = `mcpany-result-${Date.now()}.csv`;
        document.body.appendChild(a);
        a.click();
        document.body.removeChild(a);
        URL.revokeObjectURL(url);

        toast({ title: "Exported to CSV", description: "Your result has been successfully exported." });
    };

    const renderCell = (value: unknown) => {
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
                <TabsContent value="table" className="border rounded-md bg-card flex flex-col">
                    <div className="flex justify-end p-2 border-b bg-muted/20">
                        <Button variant="outline" size="sm" onClick={handleExportCSV} className="h-8 gap-1.5 text-xs">
                            <Download className="h-3.5 w-3.5" /> Export CSV
                        </Button>
                    </div>
                    <ScrollArea className="h-[400px]">
                        <Table>
                            <TableHeader className="sticky top-0 bg-background/95 backdrop-blur z-10 shadow-sm">
                                <TableRow>
                                    {columns.map(col => (
                                        <TableHead
                                            key={col}
                                            className="whitespace-nowrap cursor-pointer hover:bg-muted/50 transition-colors select-none"
                                            onClick={() => handleSort(col)}
                                        >
                                            <div className="flex items-center gap-1">
                                                {col}
                                                {sortColumn === col ? (
                                                    sortDirection === 'asc' ? <ArrowUp className="h-3 w-3" /> : <ArrowDown className="h-3 w-3" />
                                                ) : (
                                                    <ArrowUpDown className="h-3 w-3 opacity-20" />
                                                )}
                                            </div>
                                        </TableHead>
                                    ))}
                                </TableRow>
                            </TableHeader>
                            <TableBody>
                                {sortedContent.map((row, i) => (
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
