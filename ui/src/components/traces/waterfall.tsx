/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import React, { useState } from "react";
import { Span } from "@/types/trace";
import { cn } from "@/lib/utils";
import { ChevronDown, ChevronRight, Terminal, Globe, Database, Cpu, Activity, Code } from "lucide-react";
import { Badge } from "@/components/ui/badge";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { RichResultViewer } from "@/components/tools/rich-result-viewer";

function SpanIcon({ type }: { type: Span['type'] }) {
    switch (type) {
        case 'tool': return <Terminal className="h-3 w-3 text-amber-500" />;
        case 'service': return <Globe className="h-3 w-3 text-indigo-500" />;
        case 'resource': return <Database className="h-3 w-3 text-cyan-500" />;
        case 'core': return <Cpu className="h-3 w-3 text-blue-500" />;
        default: return <Activity className="h-3 w-3 text-muted-foreground" />;
    }
}

export function ExecutionWaterfall({
    span,
    depth = 0,
    traceStart,
    traceDuration
}: {
    span: Span;
    depth?: number;
    traceStart: number;
    traceDuration: number;
}) {
    const [expanded, setExpanded] = useState(true);

    const relativeStart = span.startTime - traceStart;
    const duration = span.endTime - span.startTime;

    const leftPct = (relativeStart / traceDuration) * 100;
    const widthPct = Math.max((duration / traceDuration) * 100, 0.5);

    return (
        <div className="group font-sans">
            <div className={cn(
                "flex items-center py-2 px-2 hover:bg-muted/30 rounded-md text-sm transition-all border-b border-border/40",
            )}>
                {/* Tree Column */}
                <div className="flex-1 flex items-center gap-2 min-w-[200px] overflow-hidden" style={{ paddingLeft: `${depth * 16}px` }}>
                     <button
                        onClick={(e) => { e.stopPropagation(); setExpanded(!expanded); }}
                        className={cn("p-0.5 rounded-sm hover:bg-muted transition-colors", !span.children?.length && "invisible")}
                    >
                        {expanded ? <ChevronDown className="h-3 w-3" /> : <ChevronRight className="h-3 w-3" />}
                    </button>
                    <SpanIcon type={span.type} />
                    <span className="font-medium truncate text-foreground/90 tracking-tight" title={span.name}>{span.name}</span>
                    {span.status === 'error' && <Badge variant="destructive" className="h-4 px-1 text-[10px] font-mono tracking-wider">ERR</Badge>}
                </div>

                {/* Timeline Column */}
                <div className="w-[40%] md:w-[50%] h-8 relative flex items-center px-4 border-l border-border/30 bg-black/5 dark:bg-black/20 rounded-r-md">
                    <div
                        className={cn(
                            "h-5 rounded-md min-w-[2px] shadow-sm relative group/bar transition-all hover:h-6 hover:-mt-0.5 hover:z-10 hover:shadow-md animate-in fade-in slide-in-from-left-2 duration-500 ease-out flex items-center justify-center overflow-hidden",
                            span.status === 'error' ? "bg-red-500/80 hover:bg-red-500 border border-red-500/50" :
                            span.type === 'tool' ? "bg-amber-400/80 hover:bg-amber-400 border border-amber-500/50" :
                            "bg-blue-400/80 hover:bg-blue-400 border border-blue-500/50"
                        )}
                        style={{
                            marginLeft: `${leftPct}%`,
                            width: `${widthPct}%`
                        }}
                    >
                        {widthPct > 5 && (
                            <span className="text-[9px] font-mono text-white/90 truncate px-1 opacity-0 group-hover/bar:opacity-100 transition-opacity drop-shadow-md">
                                {duration}ms
                            </span>
                        )}
                         {/* Tooltip on hover */}
                         <div className="absolute -top-8 left-1/2 -translate-x-1/2 bg-popover/90 backdrop-blur-md text-popover-foreground font-mono text-[10px] px-2 py-1 rounded-md shadow-lg border border-border/50 hidden group-hover/bar:block whitespace-nowrap z-50 transition-all">
                             {duration}ms
                         </div>
                    </div>
                    <span className="ml-2 font-mono text-[10px] text-muted-foreground absolute right-2 opacity-0 group-hover:opacity-100 transition-opacity">
                        {duration}ms
                    </span>
                </div>
            </div>

            {expanded && (
                <div className="text-xs">
                    {/* Nested Payload Viewer */}
                    {((span.input && Object.keys(span.input).length > 0) || (span.output && Object.keys(span.output).length > 0) || span.errorMessage) && (
                        <div className="ml-8 mr-4 my-3 p-4 bg-card/50 backdrop-blur-sm border border-border/50 rounded-xl shadow-sm transition-all hover:shadow-md">
                            <Tabs defaultValue="input" className="w-full">
                                <div className="flex items-center justify-between mb-3 border-b border-border/40 pb-2">
                                    <TabsList className="h-8 bg-muted/50 p-1 rounded-lg">
                                        <TabsTrigger value="input" className="text-xs px-3 h-6 rounded-md transition-all data-[state=active]:bg-background data-[state=active]:shadow-sm"><Code className="h-3 w-3 mr-1.5" /> Input</TabsTrigger>
                                        <TabsTrigger value="output" className="text-xs px-3 h-6 rounded-md transition-all data-[state=active]:bg-background data-[state=active]:shadow-sm"><Terminal className="h-3 w-3 mr-1.5" /> Output</TabsTrigger>
                                    </TabsList>
                                </div>
                                <TabsContent value="input" className="mt-0 border-none p-0 focus-visible:ring-0">
                                    {span.input && Object.keys(span.input).length > 0 ? (
                                        <div className="rounded-lg overflow-hidden border border-border/40 bg-background/50">
                                            <RichResultViewer result={span.input} />
                                        </div>
                                    ) : (
                                        <div className="text-muted-foreground p-3 flex items-center justify-center italic bg-muted/20 rounded-lg border border-dashed">No input data.</div>
                                    )}
                                </TabsContent>
                                <TabsContent value="output" className="mt-0 border-none p-0 focus-visible:ring-0">
                                    {span.errorMessage ? (
                                        <div className="p-4 bg-destructive/10 text-destructive border border-destructive/20 rounded-lg text-xs font-mono whitespace-pre-wrap shadow-inner">
                                            {span.errorMessage}
                                        </div>
                                    ) : span.output && Object.keys(span.output).length > 0 ? (
                                        <div className="rounded-lg overflow-hidden border border-border/40 bg-background/50">
                                            <RichResultViewer result={span.output} />
                                        </div>
                                    ) : (
                                        <div className="text-muted-foreground p-3 flex items-center justify-center italic bg-muted/20 rounded-lg border border-dashed">No output data.</div>
                                    )}
                                </TabsContent>
                            </Tabs>
                        </div>
                    )}

                    {/* Children */}
                    <div className="pl-6 border-l border-border/20 ml-3">
                        {span.children?.map(child => (
                             <ExecutionWaterfall
                                key={child.id}
                                span={child}
                                depth={depth + 1}
                                traceStart={traceStart}
                                traceDuration={traceDuration}
                            />
                        ))}
                    </div>
                </div>
            )}
        </div>
    );
}
