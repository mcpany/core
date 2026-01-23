/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { ScrollArea } from "@/components/ui/scroll-area";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";
import { AlertCircle, CheckCircle2, Clock, ChevronDown, ChevronRight, Activity, Terminal, Code, Cpu, Database, Globe, Play, Copy, Check } from "lucide-react";
import { Trace, Span } from "@/app/api/traces/route";
import { useState } from "react";
import { useToast } from "@/hooks/use-toast";
import React from "react";
import { useRouter } from "next/navigation";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Prism as SyntaxHighlighter } from 'react-syntax-highlighter';
import { vscDarkPlus } from 'react-syntax-highlighter/dist/esm/styles/prism';

/**
 * JsonView component.
 * @param props - The component props.
 * @param props.data - The data to display.
 * @returns The rendered component.
 */
function JsonView({ data }: { data: any }) {
    const [copied, setCopied] = useState(false);

    if (!data) return <span className="text-muted-foreground italic">null</span>;

    const handleCopy = () => {
        navigator.clipboard.writeText(JSON.stringify(data, null, 2));
        setCopied(true);
        setTimeout(() => setCopied(false), 2000);
    };

    return (
        <div className="relative group">
            <Button
                variant="ghost"
                size="icon"
                className="absolute right-2 top-2 h-6 w-6 opacity-0 group-hover:opacity-100 transition-opacity z-10 bg-muted/50 hover:bg-muted"
                onClick={handleCopy}
                title="Copy JSON"
            >
                {copied ? <Check className="h-3 w-3 text-green-500" /> : <Copy className="h-3 w-3" />}
            </Button>
            <div className="rounded-md overflow-hidden border">
                <SyntaxHighlighter
                    language="json"
                    style={vscDarkPlus}
                    customStyle={{ margin: 0, padding: '1rem', fontSize: '0.75rem', lineHeight: '1.5' }}
                    wrapLines={true}
                    wrapLongLines={true}
                >
                    {JSON.stringify(data, null, 2)}
                </SyntaxHighlighter>
            </div>
        </div>
    );
}

/**
 * HeadersTable component.
 * @param props - The component props.
 * @param props.headers - The headers to display.
 * @returns The rendered component.
 */
function HeadersTable({ headers }: { headers?: Record<string, string[]> }) {
    if (!headers || Object.keys(headers).length === 0) {
        return <div className="text-sm text-muted-foreground p-4 italic">No headers recorded.</div>;
    }

    return (
        <div className="border rounded-md">
            <Table>
                <TableHeader>
                    <TableRow>
                        <TableHead className="w-[200px]">Key</TableHead>
                        <TableHead>Value</TableHead>
                    </TableRow>
                </TableHeader>
                <TableBody>
                    {Object.entries(headers).map(([key, value]) => (
                        <TableRow key={key}>
                            <TableCell className="font-mono text-xs font-medium">{key}</TableCell>
                            <TableCell className="font-mono text-xs text-muted-foreground break-all">
                                {value.join(", ")}
                            </TableCell>
                        </TableRow>
                    ))}
                </TableBody>
            </Table>
        </div>
    );
}

/**
 * SpanIcon component.
 * @param props - The component props.
 * @param props.type - The type definition.
 * @returns The rendered component.
 */
function SpanIcon({ type }: { type: Span['type'] }) {
    switch (type) {
        case 'tool': return <Terminal className="h-3 w-3 text-amber-500" />;
        case 'service': return <Globe className="h-3 w-3 text-indigo-500" />;
        case 'resource': return <Database className="h-3 w-3 text-cyan-500" />;
        case 'core': return <Cpu className="h-3 w-3 text-blue-500" />;
        default: return <Activity className="h-3 w-3 text-muted-foreground" />;
    }
}

/**
 * WaterfallItem component.
 * @param props - The component props.
 * @param props.span - The span property.
 * @param props.depth - The nesting depth.
 * @param props.traceStart - The traceStart property.
 * @param props.traceDuration - The traceDuration property.
 * @returns The rendered component.
 */
function WaterfallItem({
    span,
    depth = 0,
    traceStart,
    traceDuration
}: {
    span: Span,
    depth?: number,
    traceStart: number,
    traceDuration: number
}) {
    const [expanded, setExpanded] = useState(true);

    const relativeStart = span.startTime - traceStart;
    const duration = span.endTime - span.startTime;

    // Calculate percentage width and margin for the timeline bar
    const leftPct = (relativeStart / traceDuration) * 100;
    const widthPct = Math.max((duration / traceDuration) * 100, 0.5); // Min 0.5% width to be visible

    return (
        <div className="group">
            <div className={cn(
                "flex items-center py-2 px-2 hover:bg-muted/30 rounded text-sm group-hover:bg-muted/50 transition-colors border-b border-border/40",
            )}>
                {/* Tree Column */}
                <div className="flex-1 flex items-center gap-2 min-w-[200px] overflow-hidden" style={{ paddingLeft: `${depth * 16}px` }}>
                     <button
                        onClick={(e) => { e.stopPropagation(); setExpanded(!expanded); }}
                        className={cn("p-0.5 rounded-sm hover:bg-muted", !span.children?.length && "invisible")}
                    >
                        {expanded ? <ChevronDown className="h-3 w-3" /> : <ChevronRight className="h-3 w-3" />}
                    </button>
                    <SpanIcon type={span.type} />
                    <span className="font-medium truncate" title={span.name}>{span.name}</span>
                    {span.status === 'error' && <Badge variant="destructive" className="h-4 px-1 text-[10px]">ERR</Badge>}
                </div>

                {/* Timeline Column */}
                <div className="w-[40%] md:w-[50%] h-8 relative flex items-center px-4 border-l border-border/30 bg-black/5 dark:bg-black/20">
                    <div
                        className={cn(
                            "h-5 rounded-sm min-w-[2px] opacity-80 shadow-sm relative group/bar transition-all hover:h-6 hover:-mt-1 hover:z-10",
                            span.status === 'error' ? "bg-red-500 dark:bg-red-600" :
                            span.type === 'tool' ? "bg-amber-400 dark:bg-amber-600" :
                            "bg-blue-400 dark:bg-blue-600"
                        )}
                        style={{
                            marginLeft: `${leftPct}%`,
                            width: `${widthPct}%`
                        }}
                    >
                         {/* Tooltip on hover */}
                         <div className="absolute -top-8 left-1/2 -translate-x-1/2 bg-popover text-popover-foreground text-[10px] px-2 py-1 rounded shadow-lg border hidden group-hover/bar:block whitespace-nowrap z-50">
                             {duration}ms
                         </div>
                    </div>
                    <span className="ml-2 text-[10px] text-muted-foreground absolute right-2 opacity-0 group-hover:opacity-100 transition-opacity">
                        {duration}ms
                    </span>
                </div>
            </div>

            {/* Details (Input/Output) - Only for leaf nodes or if interesting? Or maybe separate detail pane?
                Let's put it inline if selected? Or just simple key-values for now.
            */}
            {expanded && (
                <div className="text-xs pl-8">
                    {/* Children */}
                    {span.children?.map(child => (
                         <WaterfallItem
                            key={child.id}
                            span={child}
                            depth={depth + 1}
                            traceStart={traceStart}
                            traceDuration={traceDuration}
                        />
                    ))}

                    {/* Error Message */}
                    {span.errorMessage && (
                         <div className="ml-6 mt-1 mb-2 p-2 bg-red-50 dark:bg-red-900/20 text-red-600 dark:text-red-400 border border-red-200 dark:border-red-900 rounded text-xs font-mono">
                            Error: {span.errorMessage}
                        </div>
                    )}
                </div>
            )}
        </div>
    );
}


/**
 * TraceDetail.
 *
 * @param { trace - The { trace.
 */
export function TraceDetail({ trace }: { trace: Trace | null }) {
    const router = useRouter();
    const { toast } = useToast();

    if (!trace) {
        return (
            <div className="flex-1 flex items-center justify-center h-full text-muted-foreground flex-col gap-4">
                <Activity className="h-16 w-16 opacity-10" />
                <p>Select a trace to view details</p>
            </div>
        );
    }

    const handleReplay = (toolName: string, args: Record<string, unknown> | undefined) => {
         const argsStr = JSON.stringify(args || {});
         const encodedArgs = encodeURIComponent(argsStr);
         router.push(`/playground?tool=${toolName}&args=${encodedArgs}`);
    };

    const handleCopyCurl = () => {
        // Construct basic cURL command
        // Note: This is an approximation as we might not have all headers/method details perfectly
        // aligned with the original request context if strictly looking at Span data.
        // But with `requestHeaders` now available, we can do better.

        const method = trace.rootSpan.name.split(' ')[0] || 'POST';
        const url = trace.rootSpan.name.split(' ')[1] || '/'; // Just path usually

        let curl = `curl -X ${method} "${url}"`;

        if (trace.rootSpan.requestHeaders) {
            Object.entries(trace.rootSpan.requestHeaders).forEach(([k, v]) => {
                // Skip some headers that might be auto-generated or sensitive if needed
                if (v && v.length > 0) {
                    curl += ` \\\n  -H "${k}: ${v[0]}"`;
                }
            });
        }

        if (trace.rootSpan.input) {
            curl += ` \\\n  -d '${JSON.stringify(trace.rootSpan.input)}'`;
        }

        navigator.clipboard.writeText(curl);
        toast({
            title: "Copied to clipboard",
            description: "cURL command copied to clipboard",
        });
    };

    return (
        <div className="h-full flex flex-col bg-background">
            <div className="p-6 border-b flex items-start justify-between bg-muted/10">
                <div className="space-y-1">
                    <div className="flex items-center gap-2">
                        <h2 className="text-2xl font-bold tracking-tight font-mono">{trace.rootSpan.name}</h2>
                        <Badge variant={trace.status === 'success' ? 'default' : 'destructive'}>
                            {trace.status.toUpperCase()}
                        </Badge>
                    </div>
                    <div className="flex items-center gap-4 text-sm text-muted-foreground">
                        <div className="flex items-center gap-1"><Clock className="h-3 w-3" /> {trace.totalDuration}ms</div>
                        <div className="flex items-center gap-1"><Activity className="h-3 w-3" /> {new Date(trace.timestamp).toLocaleString()}</div>
                        <div className="flex items-center gap-1 font-mono text-xs bg-muted px-1 rounded">{trace.id}</div>
                    </div>
                </div>
                <div className="flex gap-2">
                     <Button
                        variant="outline"
                        size="sm"
                        onClick={handleCopyCurl}
                        className="gap-2"
                    >
                        <Copy className="h-3 w-3" /> Copy cURL
                    </Button>
                    {trace.rootSpan.type === 'tool' && (
                        <Button
                            variant="default"
                            size="sm"
                            onClick={() => handleReplay(trace.rootSpan.name, trace.rootSpan.input)}
                            className="gap-2"
                        >
                            <Play className="h-3 w-3" /> Replay in Playground
                        </Button>
                    )}
                </div>
            </div>

            <Tabs defaultValue="overview" className="flex-1 flex flex-col overflow-hidden">
                <div className="px-6 border-b">
                    <TabsList className="w-full justify-start h-12 bg-transparent p-0 gap-6">
                        <TabsTrigger value="overview" className="data-[state=active]:bg-transparent data-[state=active]:shadow-none data-[state=active]:border-b-2 data-[state=active]:border-primary rounded-none h-full px-0">Overview</TabsTrigger>
                        <TabsTrigger value="headers" className="data-[state=active]:bg-transparent data-[state=active]:shadow-none data-[state=active]:border-b-2 data-[state=active]:border-primary rounded-none h-full px-0">Headers</TabsTrigger>
                        <TabsTrigger value="request" className="data-[state=active]:bg-transparent data-[state=active]:shadow-none data-[state=active]:border-b-2 data-[state=active]:border-primary rounded-none h-full px-0">Request Payload</TabsTrigger>
                        <TabsTrigger value="response" className="data-[state=active]:bg-transparent data-[state=active]:shadow-none data-[state=active]:border-b-2 data-[state=active]:border-primary rounded-none h-full px-0">Response Body</TabsTrigger>
                    </TabsList>
                </div>

                <div className="flex-1 overflow-hidden bg-muted/5">
                    <ScrollArea className="h-full">
                        <div className="p-6 max-w-5xl mx-auto">
                            <TabsContent value="overview" className="mt-0 space-y-6">
                                <Card>
                                     <CardHeader className="pb-3">
                                        <CardTitle className="text-sm font-medium">Execution Waterfall</CardTitle>
                                    </CardHeader>
                                    <CardContent className="pl-2 pr-6">
                                         <div className="w-full border rounded-md bg-background/50 overflow-hidden">
                                             {/* Header Row */}
                                             <div className="flex text-xs font-medium text-muted-foreground border-b p-2 bg-muted/20">
                                                 <div className="flex-1 pl-2">Span Name</div>
                                                 <div className="w-[40%] md:w-[50%] pl-4 border-l">Timeline</div>
                                             </div>
                                             <WaterfallItem
                                                span={trace.rootSpan}
                                                traceStart={trace.rootSpan.startTime}
                                                traceDuration={trace.totalDuration}
                                            />
                                         </div>
                                    </CardContent>
                                </Card>
                            </TabsContent>

                            <TabsContent value="headers" className="mt-0 space-y-6">
                                <div className="grid gap-6">
                                    <div className="space-y-2">
                                        <h3 className="text-sm font-medium text-muted-foreground flex items-center gap-2">
                                            <Code className="h-4 w-4"/> Request Headers
                                        </h3>
                                        <HeadersTable headers={trace.rootSpan.requestHeaders} />
                                    </div>
                                    <div className="space-y-2">
                                        <h3 className="text-sm font-medium text-muted-foreground flex items-center gap-2">
                                            <Terminal className="h-4 w-4"/> Response Headers
                                        </h3>
                                        <HeadersTable headers={trace.rootSpan.responseHeaders} />
                                    </div>
                                </div>
                            </TabsContent>

                            <TabsContent value="request" className="mt-0">
                                <div className="space-y-2">
                                    <h3 className="text-sm font-medium text-muted-foreground mb-4">Request Body</h3>
                                    <JsonView data={trace.rootSpan.input} />
                                </div>
                            </TabsContent>

                            <TabsContent value="response" className="mt-0">
                                <div className="space-y-2">
                                    <h3 className="text-sm font-medium text-muted-foreground mb-4">Response Body</h3>
                                    <JsonView data={trace.rootSpan.output} />
                                </div>
                            </TabsContent>
                        </div>
                    </ScrollArea>
                </div>
            </Tabs>
        </div>
    );
}
