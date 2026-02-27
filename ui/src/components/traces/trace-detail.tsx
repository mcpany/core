/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { ScrollArea } from "@/components/ui/scroll-area";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Separator } from "@/components/ui/separator";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";
import { Clock, Activity, Terminal, Code, Coins, RefreshCcw, Play, Copy, Download, Lightbulb, AlertTriangle, Cpu, Globe, Database } from "lucide-react";
import { Trace, Span } from "@/types/trace";
import { useState } from "react";
import { useToast } from "@/hooks/use-toast";
import React from "react";
import { useRouter } from "next/navigation";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { JsonView } from "@/components/ui/json-view";
import { RichResultViewer } from "@/components/tools/rich-result-viewer";
import { analyzeTrace } from "@/lib/diagnostics";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { SequenceDiagram } from "@/components/traces/sequence-diagram";
import { estimateTokens, calculateCost, formatCost } from "@/lib/tokens";
import { LogStream } from "@/components/logs/log-stream";
import { ReplayDiffDialog } from "@/components/traces/replay-diff-dialog";
import { InteractiveWaterfall } from "@/components/traces/interactive-waterfall";
import { ResizableHandle, ResizablePanel, ResizablePanelGroup } from "@/components/ui/resizable";

function SpanIcon({ type }: { type: Span['type'] }) {
    switch (type) {
        case 'tool': return <Terminal className="h-4 w-4 text-amber-500" />;
        case 'service': return <Globe className="h-4 w-4 text-indigo-500" />;
        case 'resource': return <Database className="h-4 w-4 text-cyan-500" />;
        case 'core': return <Cpu className="h-4 w-4 text-blue-500" />;
        default: return <Activity className="h-4 w-4 text-muted-foreground" />;
    }
}

/**
 * TraceDetail.
 *
 * @param { trace - The { trace.
 */
export function TraceDetail({ trace }: { trace: Trace | null }) {
    const router = useRouter();
    const { toast } = useToast();
    const [isReplayOpen, setIsReplayOpen] = useState(false);
    const [selectedSpan, setSelectedSpan] = useState<Span | null>(null);

    // Default to root span if nothing selected
    const activeSpan = selectedSpan || trace?.rootSpan || null;

    if (!trace) {
        return (
            <div className="flex-1 flex items-center justify-center h-full text-muted-foreground flex-col gap-4">
                <Activity className="h-16 w-16 opacity-10" />
                <p>Select a trace to view details</p>
            </div>
        );
    }

    const diagnostics = analyzeTrace(trace);

    // Calculate estimation
    const rootInputTokens = estimateTokens(trace.rootSpan.input);
    const rootOutputTokens = estimateTokens(trace.rootSpan.output);
    const totalTokens = rootInputTokens + rootOutputTokens;
    const estimatedCost = calculateCost(totalTokens);

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
            <div className="p-4 border-b flex items-start justify-between bg-muted/10 shrink-0">
                <div className="space-y-1">
                    <div className="flex items-center gap-2">
                        <h2 className="text-xl font-bold tracking-tight font-mono">{trace.rootSpan.name}</h2>
                        <Badge variant={trace.status === 'success' ? 'default' : 'destructive'}>
                            {trace.status.toUpperCase()}
                        </Badge>
                    </div>
                    <div className="flex items-center gap-4 text-xs text-muted-foreground">
                        <div className="flex items-center gap-1"><Clock className="h-3 w-3" /> {trace.totalDuration}ms</div>
                        <Separator orientation="vertical" className="h-3" />
                        <div className="flex items-center gap-1" title={`Input: ${rootInputTokens} | Output: ${rootOutputTokens}`}>
                            <span className="font-semibold text-[10px] uppercase tracking-wider text-muted-foreground">Tokens</span>
                            <span className="font-mono">{totalTokens}</span>
                        </div>
                        <div className="flex items-center gap-1" title="Estimated Cost">
                            <Coins className="h-3 w-3 text-amber-500" />
                            <span className="font-mono">{formatCost(estimatedCost)}</span>
                        </div>
                         <Separator orientation="vertical" className="h-3" />
                        <div className="flex items-center gap-1 font-mono text-[10px] bg-muted px-1 rounded">{trace.id}</div>
                    </div>
                </div>
                <div className="flex gap-2">
                    {trace.rootSpan.type === 'tool' && (
                        <>
                            <Button
                                variant="default"
                                size="sm"
                                onClick={() => setIsReplayOpen(true)}
                                className="gap-2 h-8 text-xs"
                            >
                                <RefreshCcw className="h-3 w-3" /> Replay & Diff
                            </Button>
                        </>
                    )}
                    <Button variant="outline" size="sm" onClick={handleExportJSON} className="h-8 w-8 p-0">
                        <Download className="h-4 w-4" />
                    </Button>
                </div>
            </div>

            <Tabs defaultValue="waterfall" className="flex-1 flex flex-col overflow-hidden">
                <div className="px-4 border-b bg-background z-10">
                   <TabsList className="bg-transparent border-b-0 p-0 h-auto w-full justify-start">
                       <TabsTrigger value="waterfall" className="rounded-none border-b-2 border-transparent data-[state=active]:border-primary data-[state=active]:bg-transparent px-4 py-2 text-xs">Waterfall & Span Details</TabsTrigger>
                       <TabsTrigger value="logs" className="rounded-none border-b-2 border-transparent data-[state=active]:border-primary data-[state=active]:bg-transparent px-4 py-2 text-xs">Full Logs</TabsTrigger>
                       <TabsTrigger value="raw" className="rounded-none border-b-2 border-transparent data-[state=active]:border-primary data-[state=active]:bg-transparent px-4 py-2 text-xs">Raw JSON</TabsTrigger>
                   </TabsList>
                </div>

                <TabsContent value="waterfall" className="flex-1 p-0 overflow-hidden m-0 relative">
                     <ResizablePanelGroup direction="vertical">
                        <ResizablePanel defaultSize={50} minSize={25}>
                            <div className="h-full overflow-hidden flex flex-col">
                                <InteractiveWaterfall
                                    trace={trace}
                                    selectedSpanId={activeSpan?.id || null}
                                    onSpanSelect={setSelectedSpan}
                                />
                            </div>
                        </ResizablePanel>

                        <ResizableHandle withHandle />

                        <ResizablePanel defaultSize={50} minSize={25}>
                             <div className="h-full overflow-y-auto bg-muted/5 border-t">
                                 {activeSpan ? (
                                     <div className="p-4 space-y-6">
                                         <div className="flex items-center justify-between">
                                             <div className="flex items-center gap-2">
                                                 <SpanIcon type={activeSpan.type} />
                                                 <h3 className="text-lg font-semibold">{activeSpan.name}</h3>
                                             </div>
                                             <div className="text-xs font-mono text-muted-foreground">
                                                 {activeSpan.endTime - activeSpan.startTime}ms
                                             </div>
                                         </div>

                                         {activeSpan.errorMessage && (
                                              <Alert variant="destructive">
                                                <AlertTriangle className="h-4 w-4" />
                                                <AlertTitle>Error</AlertTitle>
                                                <AlertDescription className="font-mono text-xs mt-1">
                                                    {activeSpan.errorMessage}
                                                </AlertDescription>
                                            </Alert>
                                         )}

                                        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                                            <Card>
                                                <CardHeader className="py-2 px-4 border-b bg-muted/20">
                                                    <CardTitle className="text-xs font-medium flex items-center gap-2"><Code className="h-3 w-3"/> Input</CardTitle>
                                                </CardHeader>
                                                <CardContent className="p-0">
                                                    <div className="max-h-[300px] overflow-auto p-2">
                                                        <RichResultViewer result={activeSpan.input} />
                                                    </div>
                                                </CardContent>
                                            </Card>
                                            <Card>
                                                <CardHeader className="py-2 px-4 border-b bg-muted/20">
                                                    <CardTitle className="text-xs font-medium flex items-center gap-2"><Terminal className="h-3 w-3"/> Output</CardTitle>
                                                </CardHeader>
                                                <CardContent className="p-0">
                                                    <div className="max-h-[300px] overflow-auto p-2">
                                                         <RichResultViewer result={activeSpan.output} />
                                                    </div>
                                                </CardContent>
                                            </Card>
                                        </div>
                                     </div>
                                 ) : (
                                     <div className="h-full flex items-center justify-center text-muted-foreground text-sm">
                                         Select a span to view details
                                     </div>
                                 )}
                             </div>
                        </ResizablePanel>
                     </ResizablePanelGroup>
                </TabsContent>

                <TabsContent value="logs" className="flex-1 p-0 overflow-hidden m-0">
                    <LogStream
                        traceId={trace.id}
                        traceStartTime={trace.rootSpan.startTime}
                        traceEndTime={trace.rootSpan.endTime}
                    />
                </TabsContent>

                <TabsContent value="raw" className="flex-1 p-0 overflow-hidden m-0">
                     <ScrollArea className="h-full p-6">
                        <JsonView data={trace} />
                     </ScrollArea>
                </TabsContent>
            </Tabs>

            <ReplayDiffDialog
                open={isReplayOpen}
                onOpenChange={setIsReplayOpen}
                trace={trace}
            />
        </div>
    );
}
