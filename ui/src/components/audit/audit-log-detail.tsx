/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useMemo } from "react";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Badge } from "@/components/ui/badge";
import { Copy, Terminal, Activity, User, Clock, AlertTriangle, FileText, Image as ImageIcon } from "lucide-react";
import { Button } from "@/components/ui/button";
import { useToast } from "@/hooks/use-toast";

import ReactSyntaxHighlighter from 'react-syntax-highlighter/dist/esm/light';
import json from 'react-syntax-highlighter/dist/esm/languages/hljs/json';
import { vs2015 } from 'react-syntax-highlighter/dist/esm/styles/hljs';

ReactSyntaxHighlighter.registerLanguage('json', json);

export interface AuditLogEntry {
    timestamp: string;
    toolName: string;
    userId: string;
    profileId: string;
    arguments: string;
    result: string;
    error: string;
    duration: string;
    durationMs: number;
}

interface AuditLogDetailProps {
    entry: AuditLogEntry;
}

export function AuditLogDetail({ entry }: AuditLogDetailProps) {
    const { toast } = useToast();
    const [activeTab, setActiveTab] = useState("rich");

    const formatJson = (jsonStr: string) => {
        if (!jsonStr) return null;
        try {
            const obj = JSON.parse(jsonStr);
            return JSON.stringify(obj, null, 2);
        } catch (e) {
            return jsonStr;
        }
    };

    const handleCopy = (text: string) => {
        navigator.clipboard.writeText(text);
        toast({ title: "Copied", description: "Content copied to clipboard." });
    };

    // Analyze result for rich content
    const richContent = useMemo(() => {
        if (!entry.result) return null;
        try {
            const parsed = JSON.parse(entry.result);

            // Standard MCP Tool Result: { content: [{ type, text?, data?, mimeType? }] }
            if (parsed && Array.isArray(parsed.content)) {
                return parsed.content.map((item: any, idx: number) => {
                    if (item.type === 'image' && item.data) {
                        return (
                            <div key={idx} className="mb-4">
                                <div className="text-xs text-muted-foreground mb-1 flex items-center gap-1">
                                    <ImageIcon className="h-3 w-3" /> Image ({item.mimeType || 'unknown'})
                                </div>
                                <div className="bg-muted/50 p-4 rounded border flex justify-center">
                                    {/* eslint-disable-next-line @next/next/no-img-element */}
                                    <img
                                        src={`data:${item.mimeType || 'image/png'};base64,${item.data}`}
                                        alt={`Result ${idx}`}
                                        className="max-w-full max-h-[400px] object-contain shadow-sm"
                                    />
                                </div>
                            </div>
                        );
                    }
                    if (item.type === 'text' && item.text) {
                         return (
                            <div key={idx} className="mb-4">
                                <div className="text-xs text-muted-foreground mb-1 flex items-center gap-1">
                                    <FileText className="h-3 w-3" /> Text Output
                                </div>
                                <div className="bg-muted/30 p-3 rounded border text-sm whitespace-pre-wrap font-mono">
                                    {item.text}
                                </div>
                            </div>
                        );
                    }
                    // Resource type
                    if (item.type === 'resource' && item.resource) {
                         return (
                            <div key={idx} className="mb-4">
                                <div className="text-xs text-muted-foreground mb-1 flex items-center gap-1">
                                    <FileText className="h-3 w-3" /> Embedded Resource
                                </div>
                                <div className="bg-muted/30 p-3 rounded border text-sm font-mono">
                                    <div className="font-semibold text-xs text-primary mb-1">{item.resource.uri}</div>
                                    {item.resource.text ? item.resource.text : `(Blob content: ${item.resource.blob?.length || 0} bytes)`}
                                </div>
                            </div>
                        );
                    }
                    return (
                         <div key={idx} className="mb-4 text-xs text-muted-foreground border p-2 rounded">
                            Unknown content block: {JSON.stringify(item).substring(0, 100)}...
                         </div>
                    );
                });
            }

            // Fallback for simple JSON object or primitive
            return null;
        } catch (e) {
            return null;
        }
    }, [entry.result]);

    return (
        <div className="flex flex-col h-full overflow-hidden">
             <div className="flex items-start justify-between p-6 pb-4 border-b shrink-0">
                <div>
                     <div className="flex items-center gap-2 mb-1">
                        <Badge variant={entry.error ? "destructive" : "outline"} className={entry.error ? "" : "text-green-600 border-green-600/30 bg-green-50 dark:bg-green-900/10"}>
                            {entry.error ? "Failed" : "Success"}
                        </Badge>
                        <h2 className="text-lg font-bold tracking-tight font-mono">{entry.toolName}</h2>
                     </div>
                    <div className="flex items-center gap-4 text-xs text-muted-foreground mt-2">
                        <div className="flex items-center gap-1"><User className="h-3 w-3" /> {entry.userId || "Anonymous"}</div>
                        <div className="flex items-center gap-1"><Clock className="h-3 w-3" /> {entry.duration}</div>
                        <div className="flex items-center gap-1"><Activity className="h-3 w-3" /> {new Date(entry.timestamp).toLocaleString()}</div>
                    </div>
                </div>
                <div className="flex gap-2">
                     <Button variant="outline" size="sm" onClick={() => handleCopy(JSON.stringify(entry, null, 2))}>
                        <Copy className="h-3 w-3 mr-2" /> Copy Log
                    </Button>
                </div>
            </div>

            <Tabs value={activeTab} onValueChange={setActiveTab} className="flex-1 flex flex-col overflow-hidden">
                <div className="px-6 border-b bg-muted/5 shrink-0">
                   <TabsList className="bg-transparent border-b-0 p-0 h-auto w-full justify-start rounded-none">
                       <TabsTrigger value="rich" className="rounded-none border-b-2 border-transparent data-[state=active]:border-primary data-[state=active]:bg-transparent px-4 py-2">Activity View</TabsTrigger>
                       <TabsTrigger value="raw" className="rounded-none border-b-2 border-transparent data-[state=active]:border-primary data-[state=active]:bg-transparent px-4 py-2">Raw Data</TabsTrigger>
                   </TabsList>
                </div>

                <TabsContent value="rich" className="flex-1 p-0 overflow-hidden m-0 data-[state=inactive]:hidden">
                    <ScrollArea className="h-full p-6">
                        {entry.error && (
                            <div className="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-900 rounded-md p-4 mb-6">
                                <div className="flex items-center gap-2 text-red-700 dark:text-red-400 font-semibold mb-2">
                                    <AlertTriangle className="h-4 w-4" /> Execution Error
                                </div>
                                <pre className="text-xs whitespace-pre-wrap font-mono text-red-600 dark:text-red-300">
                                    {entry.error}
                                </pre>
                            </div>
                        )}

                        <div className="space-y-6">
                            {/* Inputs */}
                            <div>
                                <h3 className="text-sm font-medium mb-3 flex items-center gap-2 text-primary">
                                    <Terminal className="h-4 w-4" /> Input Arguments
                                </h3>
                                <div className="rounded-md border bg-muted/30 overflow-hidden">
                                    <ReactSyntaxHighlighter
                                        language="json"
                                        style={vs2015}
                                        customStyle={{ margin: 0, padding: '1rem', fontSize: '12px' }}
                                    >
                                        {formatJson(entry.arguments) || "{}"}
                                    </ReactSyntaxHighlighter>
                                </div>
                            </div>

                            {/* Outputs */}
                            <div>
                                <h3 className="text-sm font-medium mb-3 flex items-center gap-2 text-primary">
                                    <Activity className="h-4 w-4" /> Result
                                </h3>
                                {richContent ? (
                                    <div className="space-y-4">
                                        {richContent}
                                    </div>
                                ) : (
                                    <div className="rounded-md border bg-muted/30 overflow-hidden">
                                        <ReactSyntaxHighlighter
                                            language="json"
                                            style={vs2015}
                                            customStyle={{ margin: 0, padding: '1rem', fontSize: '12px' }}
                                        >
                                            {formatJson(entry.result) || (entry.error ? "null" : "{}")}
                                        </ReactSyntaxHighlighter>
                                    </div>
                                )}
                            </div>
                        </div>
                    </ScrollArea>
                </TabsContent>

                <TabsContent value="raw" className="flex-1 p-0 overflow-hidden m-0 data-[state=inactive]:hidden">
                    <ScrollArea className="h-full">
                        <ReactSyntaxHighlighter
                            language="json"
                            style={vs2015}
                            customStyle={{ margin: 0, padding: '1.5rem', fontSize: '12px', minHeight: '100%' }}
                            showLineNumbers={true}
                        >
                            {JSON.stringify(entry, null, 2)}
                        </ReactSyntaxHighlighter>
                    </ScrollArea>
                </TabsContent>
            </Tabs>
        </div>
    );
}
