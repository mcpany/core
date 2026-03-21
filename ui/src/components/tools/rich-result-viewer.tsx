/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */



import { useMemo } from "react";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { ScrollArea } from "@/components/ui/scroll-area";
import { FileJson, Table as TableIcon, Terminal, FileText } from "lucide-react";
import { JsonView } from "@/components/ui/json-view";
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

        let extractedContent = result;
        let didExtract = false;

        // Handle Command Execution Result (stdout contains JSON)
        if (typeof result === 'object' && 'stdout' in result && typeof result.stdout === 'string') {
            try {
                // Only treat as extracted if parsing succeeds
                extractedContent = JSON.parse(result.stdout);
                didExtract = true;
            } catch {
                extractedContent = result;
            }
        }

        // Handle raw string that is JSON
        if (!didExtract && typeof result === 'string') {
            try {
                extractedContent = JSON.parse(result);
                didExtract = true;
            } catch {
                extractedContent = result;
            }
        }

        // Deep search for arrays to surface as table data
        // e.g., { data: { items: [...] } } or { response: { data: [...] } }
        if (extractedContent && typeof extractedContent === 'object' && !Array.isArray(extractedContent)) {
            // Find the largest array in the object to use as the table
            // eslint-disable-next-line @typescript-eslint/no-explicit-any
            let largestArray: any[] | null = null;

            // eslint-disable-next-line @typescript-eslint/no-explicit-any
            const findArray = (obj: any, depth = 0) => {
                if (depth > 3 || !obj || typeof obj !== 'object') return;

                if (Array.isArray(obj)) {
                    if (!largestArray || obj.length > largestArray.length) {
                        largestArray = obj;
                    }
                    return;
                }

                for (const key in obj) {
                    findArray(obj[key], depth + 1);
                }
            };

            findArray(extractedContent);

            // eslint-disable-next-line @typescript-eslint/no-explicit-any
            const validLargestArray = largestArray as any[] | null;
            if (validLargestArray && validLargestArray.length > 0 && typeof validLargestArray[0] === 'object') {
                 // We found a meaningful array to display, but we also want to keep the original content
                 // for the JSON view. So we'll just return the array as the content if it's large enough,
                 // or maybe attach it as a separate view?
                 // For simplicity, if we find an array inside an object, we just return the array.
                 return [validLargestArray, true];
            }
        }

        return [extractedContent, didExtract];
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

    const isCardsEligible = useMemo(() => {
        return !mcpContent && !isTableEligible && content && typeof content === 'object' && !Array.isArray(content) && Object.keys(content).length > 0;
    }, [content, mcpContent, isTableEligible]);

    // Get columns for table
    const columns = useMemo(() => {
        if (!isTableEligible) return [];
        // aggregate all keys from all objects to handle sparse data
        const keys = new Set<string>();
        // Limit rows scanned for columns to avoid perf issues on huge datasets
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        content.slice(0, 50).forEach((item: any) => {
            if (typeof item === 'object' && item !== null) {
                Object.keys(item).forEach(k => keys.add(k));
            }
        });
        return Array.from(keys);
    }, [content, isTableEligible]);

    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const renderCell = (value: any) => {
        if (value === null || value === undefined) return <span className="text-muted-foreground">-</span>;
        if (typeof value === 'object') return <span className="font-mono text-xs text-muted-foreground truncate max-w-[200px] block" title={JSON.stringify(value)}>{JSON.stringify(value)}</span>;
        if (typeof value === 'boolean') return <span className={value ? "text-green-500 font-medium" : "text-red-500 font-medium"}>{String(value)}</span>;
        return <span className="truncate max-w-[300px] block" title={String(value)}>{String(value)}</span>;
    }

    const defaultTab = mcpContent ? "rendered" : (isTableEligible ? "table" : (isCardsEligible ? "cards" : "json"));

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
                    {isCardsEligible && (
                        <TabsTrigger value="cards" className="flex items-center gap-2">
                            <FileText className="h-4 w-4" /> Data
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
                                        <TableHead key={col} className="whitespace-nowrap font-semibold bg-muted/50">{col}</TableHead>
                                    ))}
                                </TableRow>
                            </TableHeader>
                            <TableBody>
                                {/* eslint-disable-next-line @typescript-eslint/no-explicit-any */}
                                {content.map((row: any, i: number) => (
                                    <TableRow key={i} className="hover:bg-muted/30 transition-colors">
                                        {columns.map(col => (
                                            <TableCell key={col} className="py-3 align-top">
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

            {isCardsEligible && (
                <TabsContent value="cards" className="border rounded-md bg-card p-4">
                    <ScrollArea className="h-[400px]">
                        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                            {Object.entries(content).map(([key, val]) => (
                                <div key={key} className="flex flex-col p-3 border rounded bg-background/50 hover:bg-muted/20 transition-colors shadow-sm">
                                    <span className="text-xs font-semibold text-muted-foreground uppercase tracking-wider mb-1">{key}</span>
                                    <div className="text-sm break-words overflow-auto max-h-[150px]">
                                        {renderCell(val)}
                                    </div>
                                </div>
                            ))}
                        </div>
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
