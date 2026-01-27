/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { ScrollArea } from "@/components/ui/scroll-area";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";
import { Clock, ChevronDown, ChevronRight, Activity, Terminal, Code, Cpu, Database, Globe, Play, Download, Copy, Lightbulb, AlertTriangle } from "lucide-react";
import { Trace, Span } from "@/types/trace";
import { useState } from "react";
import { useToast } from "@/hooks/use-toast";
import React from "react";
import { useRouter } from "next/navigation";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { JsonView } from "@/components/ui/json-view";
import { analyzeTrace } from "@/lib/diagnostics";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { traceToMermaid } from "@/lib/trace-to-mermaid";
import { Mermaid } from "@/components/ui/mermaid";

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

    const diagnostics = analyzeTrace(trace);

    const handleReplay = (toolName: string, args: Record<string, unknown> | undefined) => {
         const argsStr = JSON.stringify(args || {});
         const encodedArgs = encodeURIComponent(argsStr);
         router.push(`/playground?tool=${toolName}&args=${encodedArgs}`);
    };

    const handleExportJSON = () => {
        if (!trace) return;
        const dataStr = "data:text/json;charset=utf-8," + encodeURIComponent(JSON.stringify(trace, null, 2));
        const downloadAnchorNode = document.createElement('a');
        downloadAnchorNode.setAttribute("href", dataStr);
        downloadAnchorNode.setAttribute("download", `trace-${trace.id}.json`);
        document.body.appendChild(downloadAnchorNode); // required for firefox
        downloadAnchorNode.click();
        downloadAnchorNode.remove();
    };

    const handleCopyJSON = () => {
        if (!trace) return;
        navigator.clipboard.writeText(JSON.stringify(trace, null, 2));
        toast({
            title: "Copied to clipboard",
            description: "Trace JSON has been copied to your clipboard.",
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
                    <Button variant="outline" size="sm" onClick={handleCopyJSON} className="gap-2">
                        <Copy className="h-3 w-3" /> Copy
                    </Button>
                    <Button variant="outline" size="sm" onClick={handleExportJSON} className="gap-2">
                        <Download className="h-3 w-3" /> Export JSON
                    </Button>
                </div>
            </div>

            <Tabs defaultValue="overview" className="flex-1 flex flex-col overflow-hidden">
                <div className="px-6 border-b">
                   <TabsList className="bg-transparent border-b-0 p-0 h-auto">
                       <TabsTrigger value="overview" className="rounded-none border-b-2 border-transparent data-[state=active]:border-primary data-[state=active]:bg-transparent px-4 py-2">Overview</TabsTrigger>
                       <TabsTrigger value="sequence" className="rounded-none border-b-2 border-transparent data-[state=active]:border-primary data-[state=active]:bg-transparent px-4 py-2">Sequence</TabsTrigger>
                       <TabsTrigger value="payload" className="rounded-none border-b-2 border-transparent data-[state=active]:border-primary data-[state=active]:bg-transparent px-4 py-2">Payload</TabsTrigger>
                   </TabsList>
                </div>
                <TabsContent value="overview" className="flex-1 p-0 overflow-hidden m-0">
                    <ScrollArea className="h-full p-6">
                        {diagnostics.length > 0 && (
                            <Card className="mb-6 border-l-4 border-l-destructive">
                                <CardHeader className="pb-3">
                                    <CardTitle className="text-sm font-medium flex items-center gap-2">
                                        <Lightbulb className="h-4 w-4 text-amber-500" />
                                        Diagnostics & Suggestions
                                    </CardTitle>
                                    <CardDescription>
                                        Intelligent analysis of the error.
                                    </CardDescription>
                                </CardHeader>
                                <CardContent className="space-y-4">
                                    {diagnostics.map((diag, i) => (
                                        <Alert key={i} variant={diag.type === 'error' ? 'destructive' : 'default'}>
                                            <AlertTriangle className="h-4 w-4" />
                                            <AlertTitle>{diag.title}</AlertTitle>
                                            <AlertDescription className="mt-2">
                                                <p className="font-medium">{diag.message}</p>
                                                {diag.suggestion && (
                                                    <p className="mt-1 text-muted-foreground opacity-90">
                                                        <span className="font-semibold">Suggestion:</span> {diag.suggestion}
                                                    </p>
                                                )}
                                            </AlertDescription>
                                        </Alert>
                                    ))}
                                </CardContent>
                            </Card>
                        )}
                        <Card className="mb-6">
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

                        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                            <Card>
                                <CardHeader className="pb-3">
                                    <CardTitle className="text-sm font-medium flex items-center gap-2"><Code className="h-4 w-4"/> Root Input</CardTitle>
                                </CardHeader>
                                <CardContent>
                                    <JsonView data={trace.rootSpan.input} />
                                </CardContent>
                            </Card>
                            <Card>
                                <CardHeader className="pb-3">
                                    <CardTitle className="text-sm font-medium flex items-center gap-2"><Terminal className="h-4 w-4"/> Root Output</CardTitle>
                                </CardHeader>
                                <CardContent>
                                     <JsonView data={trace.rootSpan.output} />
                                </CardContent>
                            </Card>
                        </div>
                    </ScrollArea>
                </TabsContent>
                <TabsContent value="sequence" className="flex-1 p-0 overflow-hidden m-0">
                    <ScrollArea className="h-full p-6">
                        <Card className="h-full border-none shadow-none">
                            <CardHeader className="px-0 pt-0 pb-4">
                                <CardTitle className="text-sm font-medium">Sequence Diagram</CardTitle>
                                <CardDescription>Visualizing the flow of the trace</CardDescription>
                            </CardHeader>
                            <CardContent className="px-0">
                                <Mermaid chart={traceToMermaid(trace)} />
                            </CardContent>
                        </Card>
                    </ScrollArea>
                </TabsContent>
                <TabsContent value="payload" className="flex-1 p-0 overflow-hidden m-0">
                     <ScrollArea className="h-full p-6">
                        <div className="grid grid-cols-1 gap-6">
                            <div className="space-y-2">
                                <h3 className="text-sm font-medium flex items-center gap-2 text-primary">
                                    <Code className="h-4 w-4" /> Request Payload
                                </h3>
                                <div className="bg-muted/30 border rounded-lg p-4 font-mono text-xs overflow-auto max-h-[400px]">
                                    <pre>{JSON.stringify(trace.rootSpan.input, null, 2)}</pre>
                                </div>
                            </div>
                            <div className="space-y-2">
                                <h3 className="text-sm font-medium flex items-center gap-2 text-primary">
                                    <Terminal className="h-4 w-4" /> Response Payload
                                </h3>
                                <div className="bg-muted/30 border rounded-lg p-4 font-mono text-xs overflow-auto max-h-[400px]">
                                     <pre>{JSON.stringify(trace.rootSpan.output, null, 2)}</pre>
                                </div>
                            </div>
                        </div>
                     </ScrollArea>
                </TabsContent>
            </Tabs>
        </div>
    );
}
