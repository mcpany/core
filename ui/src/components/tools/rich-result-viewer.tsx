/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */



import { useMemo } from "react";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { ScrollArea } from "@/components/ui/scroll-area";
import { FileJson, Table as TableIcon, Terminal, FileText, AlertTriangle } from "lucide-react";
import { JsonView } from "@/components/ui/json-view";
import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';
import { SmartTable } from "./smart-table";
import { Card, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";


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

/**
 * McpContentRenderer component.
 * @param props - The component props.
 * @param props.content - The content property.
 * @returns The rendered component.
 */
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
 * Intent: Document RichResultViewer
 *
 * Params:
 *   - Documented below.
 *
 * Returns:
 *   - Documented below.
 *
 * Errors:
 *   - None
 *
 * Side Effects:
 *   - None
 *
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

    const isError = useMemo(() => {
        if (result && typeof result === 'object') {
            return result.isError === true || result.error !== undefined;
        }
        return false;
    }, [result]);

    const errorMessage = useMemo(() => {
        if (isError && result) {
             if (typeof result.error === 'string') return result.error;
             if (typeof result.error === 'object') return JSON.stringify(result.error, null, 2);
             if (typeof result.message === 'string') return result.message;
        }
        return "An unknown error occurred during execution.";
    }, [isError, result]);

    const { isTableEligible, tableData } = useMemo(() => {
        if (mcpContent || isError) return { isTableEligible: false, tableData: [] };

        // 1. Array of objects
        if (Array.isArray(content) && content.length > 0 && typeof content[0] === 'object' && content[0] !== null) {
            return { isTableEligible: true, tableData: content };
        }

        // 2. Object containing an array of objects. We aggressively scan the top 2 levels.
        if (content && typeof content === 'object' && !Array.isArray(content) && content !== null) {
            let largestArray: any[] = [];

            // Level 1 scan
            Object.values(content).forEach(val => {
                 if (Array.isArray(val) && val.length > 0 && val.every(item => typeof item === 'object' && item !== null && !Array.isArray(item))) {
                     if (val.length > largestArray.length) {
                         largestArray = val;
                     }
                 }
                 // Level 2 scan
                 else if (val && typeof val === 'object' && !Array.isArray(val) && val !== null) {
                      Object.values(val).forEach(nestedVal => {
                           if (Array.isArray(nestedVal) && nestedVal.length > 0 && nestedVal.every(item => typeof item === 'object' && item !== null && !Array.isArray(item))) {
                               if (nestedVal.length > largestArray.length) {
                                   largestArray = nestedVal;
                               }
                           }
                      });
                 }
            });

            if (largestArray.length > 0) {
                 return { isTableEligible: true, tableData: largestArray };
            }
        }

        return { isTableEligible: false, tableData: [] };
    }, [content, mcpContent]);

    // Get columns for table
    const columns = useMemo(() => {
        if (!isTableEligible) return [];
        // aggregate all keys from all objects to handle sparse data
        const keys = new Set<string>();
        // Limit rows scanned for columns to avoid perf issues on huge datasets
        tableData.slice(0, 50).forEach((item: any) => {
            if (typeof item === 'object' && item !== null) {
                Object.keys(item).forEach(k => keys.add(k));
            }
        });
        return Array.from(keys);
    }, [tableData, isTableEligible]);

    const renderCell = (value: any) => {
        if (value === null || value === undefined) return <span className="text-muted-foreground">-</span>;
        if (typeof value === 'object') return <span className="font-mono text-xs text-muted-foreground truncate max-w-[200px] block" title={JSON.stringify(value)}>{JSON.stringify(value)}</span>;
        if (typeof value === 'boolean') return <span className={value ? "text-green-500 font-medium" : "text-red-500 font-medium"}>{String(value)}</span>;
        return <span className="truncate max-w-[300px] block" title={String(value)}>{String(value)}</span>;
    }

    const defaultTab = isError ? "error" : (mcpContent ? "rendered" : (isTableEligible ? "table" : "json"));

    return (
        <Tabs defaultValue={defaultTab} className="w-full">
            <div className="flex items-center justify-between mb-2">
                <TabsList>
                    {isError && (
                        <TabsTrigger value="error" className="flex items-center gap-2 text-destructive data-[state=active]:text-destructive data-[state=active]:bg-destructive/10">
                            <AlertTriangle className="h-4 w-4" /> Error
                        </TabsTrigger>
                    )}
                    {!isError && mcpContent && (
                        <TabsTrigger value="rendered" className="flex items-center gap-2">
                            <FileText className="h-4 w-4" /> Rendered
                        </TabsTrigger>
                    )}
                    {!isError && isTableEligible && (
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

            {isError && (
                <TabsContent value="error" className="mt-0">
                    <Card className="border-destructive/30 bg-destructive/5 shadow-inner">
                        <CardContent className="pt-6 pb-6 flex flex-col items-center justify-center text-center space-y-4">
                            <div className="h-12 w-12 rounded-full bg-destructive/20 flex items-center justify-center">
                                <AlertTriangle className="h-6 w-6 text-destructive" />
                            </div>
                            <div className="space-y-1 max-w-md">
                                <h3 className="font-semibold text-destructive tracking-tight">Execution Failed</h3>
                                <p className="text-sm text-muted-foreground whitespace-pre-wrap font-mono bg-background/50 p-3 rounded-md border border-destructive/20 mt-2 text-left overflow-x-auto">
                                    {errorMessage}
                                </p>
                            </div>
                        </CardContent>
                    </Card>
                </TabsContent>
            )}

            {mcpContent && (
                <TabsContent value="rendered" className="border rounded-md bg-card">
                    <ScrollArea className="h-[400px]">
                        <McpContentRenderer content={mcpContent} />
                    </ScrollArea>
                </TabsContent>
            )}

            {isTableEligible && (
                <TabsContent value="table" className="border rounded-md">
                    <div className="h-[400px]">
                        <SmartTable data={tableData} />
                    </div>
                </TabsContent>
            )}

            <TabsContent value="json">
                <JsonView data={content} maxHeight={400} defaultExpandedLevel={2} smartTable={true} />
            </TabsContent>

            {isExtracted && (
                <TabsContent value="raw">
                    <JsonView data={result} maxHeight={400} smartTable={false} />
                </TabsContent>
            )}
        </Tabs>
    );
}
