/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useMemo } from "react";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { JsonView } from "@/components/ui/json-view";
import { ScrollArea } from "@/components/ui/scroll-area";
import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';
import { cn } from "@/lib/utils";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { AlertCircle, FileText, Code, Table as TableIcon, FileJson } from "lucide-react";

interface RichResultViewerProps {
    result: unknown;
    isError?: boolean;
    className?: string;
}

/**
 * RichResultViewer component.
 * Renders execution results in various formats (Table, JSON, Markdown, Text).
 *
 * @param props - The component props.
 * @param props.result - The result data to display.
 * @param props.isError - Whether the result represents an error.
 * @param props.className - Additional CSS classes.
 * @returns The rendered component.
 */
export function RichResultViewer({ result, isError, className }: RichResultViewerProps) {
    const [parsedJson, isJson, isTableCompatible] = useMemo(() => {
        if (!result) return [null, false, false];

        let content = result;
        let validJson = false;

        // Try parsing string to JSON
        if (typeof content === 'string') {
            try {
                const parsed = JSON.parse(content);
                // Simple numbers/bools are valid JSON but we might prefer text view
                if (typeof parsed === 'object' && parsed !== null) {
                    content = parsed;
                    validJson = true;
                }
            } catch (_e) {
                // Not JSON
            }
        } else if (typeof content === 'object' && content !== null) {
            validJson = true;
        }

        let tableCompat = false;
        if (validJson && Array.isArray(content) && content.length > 0) {
             const isListOfObjects = content.every((item: unknown) => typeof item === 'object' && item !== null && !Array.isArray(item));
             if (isListOfObjects) tableCompat = true;
        }

        return [content, validJson, tableCompat];
    }, [result]);

    const isMarkdown = useMemo(() => {
        if (isJson || typeof result !== 'string') return false;
        // Heuristic: check for markdown syntax
        const mdSyntax = ["# ", "* ", "- ", "```", "> ", "**", "__", "`"];
        return mdSyntax.some(s => result.includes(s));
    }, [result, isJson]);

    // Determine default tab
    const defaultTab = useMemo(() => {
        if (isTableCompatible) return "table";
        if (isJson) return "json";
        if (isMarkdown) return "markdown";
        return "text";
    }, [isTableCompatible, isJson, isMarkdown]);

    if (!result) return null;

    return (
        <div className={cn("w-full space-y-2", className)}>
            {isError && (
                <Alert variant="destructive">
                    <AlertCircle className="h-4 w-4" />
                    <AlertTitle>Execution Error</AlertTitle>
                    <AlertDescription className="font-mono text-xs break-all">
                        {typeof result === 'string' ? result : JSON.stringify(result)}
                    </AlertDescription>
                </Alert>
            )}

            {!isError && (
                <Tabs defaultValue={defaultTab} key={defaultTab} className="w-full">
                    <div className="flex justify-between items-center mb-2">
                        <TabsList className="h-8">
                            {isTableCompatible && (
                                <TabsTrigger value="table" className="text-xs h-6 px-2">
                                    <TableIcon className="mr-1 h-3 w-3" /> Table
                                </TabsTrigger>
                            )}
                            {isJson && (
                                <TabsTrigger value="json" className="text-xs h-6 px-2">
                                    <FileJson className="mr-1 h-3 w-3" /> JSON
                                </TabsTrigger>
                            )}
                            {isMarkdown && (
                                <TabsTrigger value="markdown" className="text-xs h-6 px-2">
                                    <FileText className="mr-1 h-3 w-3" /> Markdown
                                </TabsTrigger>
                            )}
                            <TabsTrigger value="text" className="text-xs h-6 px-2">
                                <Code className="mr-1 h-3 w-3" /> Raw
                            </TabsTrigger>
                        </TabsList>
                        <div className="text-[10px] text-muted-foreground uppercase font-semibold">
                            {isTableCompatible ? "List Result" : isJson ? "JSON Object" : isMarkdown ? "Markdown Text" : "Plain Text"}
                        </div>
                    </div>

                    {isTableCompatible && (
                        <TabsContent value="table" className="mt-0">
                            <JsonView data={parsedJson} smartTable={true} maxHeight={0} />
                        </TabsContent>
                    )}

                    {isJson && (
                        <TabsContent value="json" className="mt-0">
                            <JsonView data={parsedJson} smartTable={false} maxHeight={500} />
                        </TabsContent>
                    )}

                    {isMarkdown && (
                         <TabsContent value="markdown" className="mt-0">
                            <ScrollArea className="h-[400px] w-full rounded-md border p-4 bg-background">
                                <div className="prose prose-sm dark:prose-invert max-w-none">
                                    <ReactMarkdown remarkPlugins={[remarkGfm]}>
                                        {result as string}
                                    </ReactMarkdown>
                                </div>
                            </ScrollArea>
                        </TabsContent>
                    )}

                    <TabsContent value="text" className="mt-0">
                        <ScrollArea className="h-[300px] w-full rounded-md border p-4 bg-muted/50">
                            <pre className="text-xs font-mono break-all whitespace-pre-wrap">
                                {typeof result === 'string' ? result : JSON.stringify(result, null, 2)}
                            </pre>
                        </ScrollArea>
                    </TabsContent>
                </Tabs>
            )}
        </div>
    );
}
