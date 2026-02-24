/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useMemo } from "react";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Prism as SyntaxHighlighter } from 'react-syntax-highlighter';
import { vscDarkPlus } from 'react-syntax-highlighter/dist/esm/styles/prism';
import { FileJson, Table as TableIcon, Terminal, LayoutTemplate, Image as ImageIcon } from "lucide-react";
import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';
import { cn } from "@/lib/utils";

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

interface ResourceContent {
  type: "resource";
  resource: {
      uri: string;
      text?: string;
      blob?: string;
      mimeType?: string;
  };
}

type McpContent = TextContent | ImageContent | ResourceContent;

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

    const isMcpResult = useMemo(() => {
        return content && typeof content === 'object' && Array.isArray(content.content) && content.content.length > 0;
    }, [content]);

    const isTableEligible = useMemo(() => {
        return Array.isArray(content) && content.length > 0 && typeof content[0] === 'object' && content[0] !== null;
    }, [content]);

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

    const renderMcpContent = (item: McpContent, index: number) => {
        if (item.type === 'text') {
            return (
                <div key={index} className="prose prose-sm dark:prose-invert max-w-none p-4 rounded-md border bg-card mb-4 last:mb-0">
                    <ReactMarkdown remarkPlugins={[remarkGfm]} components={{
                         code(props) {
                             const {children, className, node, ...rest} = props
                             const match = /language-(\w+)/.exec(className || '')
                             return match ? (
                                 <SyntaxHighlighter
                                     {...rest}
                                     PreTag="div"
                                     children={String(children).replace(/\n$/, '')}
                                     language={match[1]}
                                     style={vscDarkPlus}
                                 />
                             ) : (
                                 <code {...rest} className={cn("bg-muted px-1.5 py-0.5 rounded font-mono text-sm", className)}>
                                     {children}
                                 </code>
                             )
                         }
                    }}>
                        {item.text}
                    </ReactMarkdown>
                </div>
            );
        }
        if (item.type === 'image') {
            return (
                <div key={index} className="flex flex-col gap-2 p-4 rounded-md border bg-card mb-4 last:mb-0 items-start">
                    <div className="flex items-center gap-2 text-sm text-muted-foreground mb-2">
                        <ImageIcon className="h-4 w-4" />
                        <span>Image ({item.mimeType})</span>
                    </div>
                    <img
                        src={`data:${item.mimeType};base64,${item.data}`}
                        alt="Tool Output"
                        className="max-w-full rounded-md border shadow-sm"
                        style={{ maxHeight: '400px' }}
                    />
                </div>
            );
        }
        if (item.type === 'resource') {
             return (
                 <div key={index} className="p-4 rounded-md border bg-card mb-4 last:mb-0">
                     <h4 className="font-medium mb-2">Resource: {item.resource.uri}</h4>
                     {item.resource.text ? (
                         <pre className="text-xs overflow-auto max-h-[200px] bg-muted p-2 rounded">{item.resource.text}</pre>
                     ) : (
                         <div className="text-xs text-muted-foreground italic">Binary content ({item.resource.mimeType})</div>
                     )}
                 </div>
             )
        }
        return null;
    };

    return (
        <Tabs defaultValue={isMcpResult ? "rendered" : (isTableEligible ? "table" : "json")} className="w-full">
            <div className="flex items-center justify-between mb-2">
                <TabsList>
                    {isMcpResult && (
                        <TabsTrigger value="rendered" className="flex items-center gap-2">
                            <LayoutTemplate className="h-4 w-4" /> Rendered
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

            {isMcpResult && (
                <TabsContent value="rendered" className="border-none">
                    <ScrollArea className="h-[500px] pr-4">
                        <div className="flex flex-col gap-4">
                            {content.content.map((item: McpContent, idx: number) => renderMcpContent(item, idx))}
                        </div>
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
