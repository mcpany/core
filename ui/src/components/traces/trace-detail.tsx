/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */



import { ScrollArea } from "@/components/ui/scroll-area";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Separator } from "@/components/ui/separator";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";
import { Clock, ChevronDown, ChevronRight, Activity, Terminal, Code, Cpu, Database, Globe, Play, Download, Copy, Lightbulb, AlertTriangle, Coins, RefreshCcw } from "lucide-react";
import { Trace, Span } from "@/types/trace";
import { useState } from "react";
import { useToast } from "@/hooks/use-toast";
import React from "react";
import { useNavigate as useRouter } from 'react-router-dom';
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { JsonView } from "@/components/ui/json-view";
import { RichResultViewer } from "@/components/tools/rich-result-viewer";
import { analyzeTrace } from "@/lib/diagnostics";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { SequenceDiagram } from "@/components/traces/sequence-diagram";
import { estimateTokens, calculateCost, formatCost } from "@/lib/tokens";
import { LogStream } from "@/components/logs/log-stream";
import { ReplayDiffDialog } from "@/components/traces/replay-diff-dialog";
import { ExecutionWaterfall } from "@/components/traces/waterfall";

/**
 * TraceDetail.
 *
 * @param { trace - The { trace.
 */
export function TraceDetail({ trace }: { trace: Trace | null }) {
    const router = useRouter();
    const { toast } = useToast();
    const [isReplayOpen, setIsReplayOpen] = useState(false);

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
         router(`/playground?tool=${toolName}&args=${encodedArgs}`);
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

                        <Separator orientation="vertical" className="h-4" />

                        <div className="flex items-center gap-1" title={`Input: ${rootInputTokens} | Output: ${rootOutputTokens}`}>
                            <span className="font-semibold text-[10px] uppercase tracking-wider text-muted-foreground">Tokens</span>
                            <span className="font-mono">{totalTokens}</span>
                        </div>
                        <div className="flex items-center gap-1" title="Estimated Cost">
                            <Coins className="h-3 w-3 text-amber-500" />
                            <span className="font-mono">{formatCost(estimatedCost)}</span>
                        </div>

                        <Separator orientation="vertical" className="h-4" />

                        <div className="flex items-center gap-1 font-mono text-xs bg-muted px-1 rounded">{trace.id}</div>
                    </div>
                </div>
                <div className="flex gap-2">
                    {trace.rootSpan.type === 'tool' && (
                        <>
                            <Button
                                variant="default"
                                size="sm"
                                onClick={() => setIsReplayOpen(true)}
                                className="gap-2"
                            >
                                <RefreshCcw className="h-3 w-3" /> Replay & Diff
                            </Button>
                            <Button
                                variant="outline"
                                size="sm"
                                onClick={() => handleReplay(trace.rootSpan.name, trace.rootSpan.input)}
                                className="gap-2"
                            >
                                <Play className="h-3 w-3" /> Replay in Playground
                            </Button>
                        </>
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
                       <TabsTrigger value="logs" className="rounded-none border-b-2 border-transparent data-[state=active]:border-primary data-[state=active]:bg-transparent px-4 py-2">Logs</TabsTrigger>
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
                                <CardTitle className="text-sm font-medium">Sequence Diagram</CardTitle>
                            </CardHeader>
                            <CardContent>
                                <SequenceDiagram trace={trace} />
                            </CardContent>
                        </Card>

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
                                     <ExecutionWaterfall
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
                                    <RichResultViewer result={trace.rootSpan.input} />
                                </CardContent>
                            </Card>
                            <Card>
                                <CardHeader className="pb-3">
                                    <CardTitle className="text-sm font-medium flex items-center gap-2"><Terminal className="h-4 w-4"/> Root Output</CardTitle>
                                </CardHeader>
                                <CardContent>
                                     <RichResultViewer result={trace.rootSpan.output} />
                                </CardContent>
                            </Card>
                        </div>
                    </ScrollArea>
                </TabsContent>
                <TabsContent value="logs" className="flex-1 p-0 overflow-hidden m-0">
                    <LogStream
                        traceId={trace.id}
                        traceStartTime={trace.rootSpan.startTime}
                        traceEndTime={trace.rootSpan.endTime}
                    />
                </TabsContent>
                <TabsContent value="payload" className="flex-1 p-0 overflow-hidden m-0">
                     <ScrollArea className="h-full p-6">
                        <div className="grid grid-cols-1 gap-6">
                            <div className="space-y-2">
                                <h3 className="text-sm font-medium flex items-center gap-2 text-primary">
                                    <Code className="h-4 w-4" /> Request Payload
                                </h3>
                                <JsonView data={trace.rootSpan.input} maxHeight={400} />
                            </div>
                            <div className="space-y-2">
                                <h3 className="text-sm font-medium flex items-center gap-2 text-primary">
                                    <Terminal className="h-4 w-4" /> Response Payload
                                </h3>
                                <JsonView data={trace.rootSpan.output} maxHeight={400} />
                            </div>
                        </div>
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
