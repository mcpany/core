/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState } from "react";
import { cn } from "@/lib/utils";
import { ChevronDown, ChevronRight, Activity, Terminal, Code, Cpu, Database, Globe } from "lucide-react";
import { Trace, Span } from "@/types/trace";
import { Badge } from "@/components/ui/badge";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { RichResultViewer } from "@/components/tools/rich-result-viewer";

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
export function WaterfallItem({
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
                "flex items-center py-2 px-2 hover:bg-muted/30 rounded text-sm group-hover:bg-muted/50 transition-colors border-b border-border/40 cursor-pointer",
            )} onClick={() => setExpanded(!expanded)}>
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

            {expanded && (
                <div className="text-xs">
                    {/* Nested Payload Viewer */}
                    {((span.input && Object.keys(span.input).length > 0) || (span.output && Object.keys(span.output).length > 0) || span.errorMessage) && (
                        <div className="ml-8 mr-4 my-2 p-3 bg-card border rounded-md shadow-sm">
                            <Tabs defaultValue="input" className="w-full">
                                <div className="flex items-center justify-between mb-2">
                                    <TabsList className="h-8">
                                        <TabsTrigger value="input" className="text-xs px-3 h-6"><Code className="h-3 w-3 mr-1" /> Input</TabsTrigger>
                                        <TabsTrigger value="output" className="text-xs px-3 h-6"><Terminal className="h-3 w-3 mr-1" /> Output</TabsTrigger>
                                    </TabsList>
                                </div>
                                <TabsContent value="input" className="mt-0 border-none p-0">
                                    {span.input && Object.keys(span.input).length > 0 ? (
                                        <RichResultViewer result={span.input} />
                                    ) : (
                                        <div className="text-muted-foreground p-2">No input data.</div>
                                    )}
                                </TabsContent>
                                <TabsContent value="output" className="mt-0 border-none p-0">
                                    {span.errorMessage ? (
                                        <div className="p-3 bg-red-50 dark:bg-red-900/20 text-red-600 dark:text-red-400 border border-red-200 dark:border-red-900 rounded text-xs font-mono whitespace-pre-wrap">
                                            {span.errorMessage}
                                        </div>
                                    ) : span.output && Object.keys(span.output).length > 0 ? (
                                        <RichResultViewer result={span.output} />
                                    ) : (
                                        <div className="text-muted-foreground p-2">No output data.</div>
                                    )}
                                </TabsContent>
                            </Tabs>
                        </div>
                    )}

                    {/* Children */}
                    <div className="pl-8">
                        {span.children?.map(child => (
                             <WaterfallItem
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

/**
 * InteractiveTimeline component.
 * @param props - The component props.
 * @param props.trace - The trace property.
 * @returns The rendered component.
 */
export function InteractiveTimeline({ trace }: { trace: Trace }) {
    return (
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
    );
}
