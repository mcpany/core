/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useMemo } from "react";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Button } from "@/components/ui/button";
import { Code, Table as TableIcon, Copy, Check, FileText, ImageIcon } from "lucide-react";
import { Prism as SyntaxHighlighter } from 'react-syntax-highlighter';
import { vscDarkPlus } from 'react-syntax-highlighter/dist/esm/styles/prism';
import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';

/**
 * Props for the SmartResultRenderer component.
 */
interface SmartResultRendererProps {
    /** The result object to render. Can be a JSON string, an object, or an array. */
    result: any;
}

interface MCPContent {
    type: string;
    text?: string;
    data?: string;
    mimeType?: string;
    resource?: {
        uri: string;
        mimeType?: string;
        text?: string;
        blob?: string;
    };
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
    const [viewMode, setViewMode] = useState<"smart" | "raw" | "table">("smart");
    const [copied, setCopied] = useState(false);

    // 1. Analyze the result to extract MCP Content or Table Data
    const { mcpContent, tableData, isMarkdown } = useMemo(() => {
        let contentList: MCPContent[] | null = null;
        let tableRows: any[] | null = null;
        let looksLikeMarkdown = false;

        // Helper to check if string is markdown-ish
        const checkMarkdown = (str: string) => {
            if (!str) return false;
            return str.includes('# ') || str.includes('**') || str.includes('```') || str.includes('| -') || str.includes('- ');
        };

        // A. Direct MCP CallToolResult
        if (result && typeof result === 'object' && Array.isArray(result.content)) {
            contentList = result.content;
        }

        // B. Command Line Output (stdout)
        const output = result?.stdout || result?.combined_output;
        if (!contentList && result && typeof result === 'object' && output && typeof output === 'string') {
            try {
                // Try parsing stdout as JSON
                const parsed = JSON.parse(output);

                // Case B1: Stdout contains MCP CallToolResult ( { content: [...] } )
                if (parsed && typeof parsed === 'object' && Array.isArray(parsed.content)) {
                    contentList = parsed.content;
                }
                // Case B2: Stdout contains Array of Objects (Table)
                else if (Array.isArray(parsed)) {
                    tableRows = parsed;
                }
                // Case B3: Just an object (fallback to raw)
            } catch {
                // Stdout is plain text. Treat as single text content.
                contentList = [{ type: 'text', text: output }];
            }
        }

        // C. Fallback for simple string/json result
        if (!contentList && !tableRows) {
            if (Array.isArray(result)) {
                tableRows = result;
            } else if (typeof result === 'string') {
                try {
                    const parsed = JSON.parse(result);
                    if (Array.isArray(parsed)) tableRows = parsed;
                    else if (parsed && typeof parsed === 'object' && Array.isArray(parsed.content)) contentList = parsed.content;
                } catch {
                    contentList = [{ type: 'text', text: result }];
                }
            }
        }

        // Validation: Ensure tableRows are actually objects
        if (tableRows) {
             const isListOfObjects = tableRows.every(item => typeof item === 'object' && item !== null && !Array.isArray(item));
             if (!isListOfObjects) tableRows = null;
        }

        // Check if contentList has markdown
        if (contentList) {
            looksLikeMarkdown = contentList.some(c => c.type === 'text' && checkMarkdown(c.text || ''));
        }

        return { mcpContent: contentList, tableData: tableRows, isMarkdown: looksLikeMarkdown };
    }, [result]);

    const copyToClipboard = () => {
        const text = typeof result === 'string' ? result : JSON.stringify(result, null, 2);
        navigator.clipboard.writeText(text);
        setCopied(true);
        setTimeout(() => setCopied(false), 2000);
    };

    const hasRichContent = mcpContent && mcpContent.some(c => c.type === 'image' || c.type === 'resource' || (c.type === 'text' && isMarkdown));
    const hasTableData = tableData !== null;

    // Default view logic
    // If rich content -> smart
    // If table data -> table (mapped to smart in UI, but technically a table view)
    // If just plain text -> smart (markdown) or raw?

    const renderRaw = () => (
        <div className="relative group/code max-h-[400px] overflow-auto">
            <SyntaxHighlighter
                language="json"
                style={vscDarkPlus}
                customStyle={{ margin: 0, padding: '1rem', fontSize: '12px', minHeight: '100%' }}
                wrapLines={true}
                wrapLongLines={true}
            >
                {JSON.stringify(result, null, 2)}
            </SyntaxHighlighter>
            <Button
                size="icon"
                variant="ghost"
                className="absolute right-2 top-2 h-6 w-6 opacity-0 group-hover/code:opacity-100 transition-opacity bg-muted/20 hover:bg-muted/50 text-white"
                onClick={copyToClipboard}
            >
                {copied ? <Check className="h-3 w-3" /> : <Copy className="h-3 w-3" />}
            </Button>
        </div>
    );

    const renderTable = (data: any[]) => {
        // Determine columns from all keys in the first 10 rows
        const allKeys = new Set<string>();
        data.slice(0, 10).forEach(row => {
            Object.keys(row).forEach(k => allKeys.add(k));
        });
        const columns = Array.from(allKeys);

        return (
            <div className="rounded-md border max-h-[400px] overflow-auto">
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
                        {data.map((row, idx) => (
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
                    Showing {data.length} rows
                </div>
            </div>
        );
    };

    const renderRichContent = (contentList: MCPContent[]) => {
        return (
            <div className="flex flex-col gap-4 p-4 bg-background/50 rounded-md border min-h-[100px]">
                {contentList.map((content, idx) => {
                    if (content.type === 'text') {
                        return (
                            <div key={idx} className="prose prose-sm dark:prose-invert max-w-none break-words">
                                <ReactMarkdown
                                    remarkPlugins={[remarkGfm]}
                                    components={{
                                        code({className, children, ...props}) {
                                            const match = /language-(\w+)/.exec(className || '')
                                            return match ? (
                                                <SyntaxHighlighter
                                                    style={vscDarkPlus}
                                                    language={match[1]}
                                                    PreTag="div"
                                                    {...props as any}
                                                >{String(children).replace(/\n$/, '')}</SyntaxHighlighter>
                                            ) : (
                                                <code className={className} {...props}>
                                                    {children}
                                                </code>
                                            )
                                        }
                                    }}
                                >
                                    {content.text || ''}
                                </ReactMarkdown>
                            </div>
                        );
                    }
                    if (content.type === 'image') {
                        return (
                            <div key={idx} className="flex flex-col items-start gap-2">
                                <div className="relative group overflow-hidden rounded-lg border bg-muted/20">
                                    <img
                                        src={`data:${content.mimeType};base64,${content.data}`}
                                        alt="Tool Output"
                                        className="max-h-[500px] w-auto object-contain"
                                    />
                                    <div className="absolute bottom-0 right-0 bg-black/60 text-white text-[10px] px-2 py-1 opacity-0 group-hover:opacity-100 transition-opacity rounded-tl-lg">
                                        {content.mimeType}
                                    </div>
                                </div>
                            </div>
                        );
                    }
                    if (content.type === 'resource') {
                        const res = content.resource;
                        return (
                            <div key={idx} className="p-3 border rounded-md bg-muted/20 flex items-center gap-3">
                                <div className="bg-primary/10 p-2 rounded">
                                    <FileText className="size-5 text-primary" />
                                </div>
                                <div className="flex flex-col">
                                    <span className="text-sm font-medium">{res?.uri}</span>
                                    <span className="text-xs text-muted-foreground">{res?.mimeType}</span>
                                </div>
                                {/* Preview if embedded text/blob */}
                                {res?.blob && res.mimeType?.startsWith('image/') && (
                                     <img
                                        src={`data:${res.mimeType};base64,${res.blob}`}
                                        alt={res.uri}
                                        className="h-10 w-10 object-cover rounded border ml-auto"
                                    />
                                )}
                            </div>
                        );
                    }
                    return null;
                })}
            </div>
        );
    };

    const activeView = viewMode;

    return (
        <div className="flex flex-col gap-0 w-full">
            <div className="flex justify-end mb-1 px-1">
                 <div className="flex items-center bg-muted/50 rounded-lg p-0.5 border">
                     {(hasRichContent || mcpContent) && (
                         <Button
                            variant={activeView === "smart" ? "secondary" : "ghost"}
                            size="sm"
                            className="h-6 px-2 text-[10px] gap-1"
                            onClick={() => setViewMode("smart")}
                         >
                             <ImageIcon className="size-3" /> Rich
                         </Button>
                     )}
                     {hasTableData && (
                         <Button
                            variant={activeView === "table" ? "secondary" : "ghost"}
                            size="sm"
                            className="h-6 px-2 text-[10px] gap-1"
                            onClick={() => setViewMode("table")}
                         >
                             <TableIcon className="size-3" /> Table
                         </Button>
                     )}
                     <Button
                        variant={activeView === "raw" ? "secondary" : "ghost"}
                        size="sm"
                        className="h-6 px-2 text-[10px] gap-1"
                        onClick={() => setViewMode("raw")}
                     >
                         <Code className="size-3" /> JSON
                     </Button>
                 </div>
            </div>

            {activeView === "smart" && mcpContent ? renderRichContent(mcpContent) :
             activeView === "table" && tableData ? renderTable(tableData) :
             // Fallback logic
             (activeView === "smart" && tableData) ? renderTable(tableData) :
             renderRaw()}
        </div>
    );
}
