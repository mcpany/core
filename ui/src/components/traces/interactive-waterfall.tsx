/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import React, { useState } from "react";
import { Trace, Span } from "@/types/trace";
import { cn } from "@/lib/utils";
import { Terminal, Globe, Database, Cpu, Activity, AlertCircle } from "lucide-react";

interface InteractiveWaterfallProps {
  trace: Trace;
  selectedSpanId: string | null;
  onSpanSelect: (span: Span) => void;
}

/**
 * Helper to flatten span tree into a list for rendering
 */
function flattenSpans(span: Span, depth = 0, result: Array<{ span: Span; depth: number }> = []) {
  result.push({ span, depth });
  if (span.children) {
    span.children.forEach((child) => flattenSpans(child, depth + 1, result));
  }
  return result;
}

function SpanIcon({ type, className }: { type: Span['type'], className?: string }) {
    switch (type) {
        case 'tool': return <Terminal className={className} />;
        case 'service': return <Globe className={className} />;
        case 'resource': return <Database className={className} />;
        case 'core': return <Cpu className={className} />;
        default: return <Activity className={className} />;
    }
}

/**
 * InteractiveWaterfall component renders a clickable, time-scaled visualization of a trace.
 */
export function InteractiveWaterfall({ trace, selectedSpanId, onSpanSelect }: InteractiveWaterfallProps) {
  const spans = flattenSpans(trace.rootSpan);
  const traceStart = trace.rootSpan.startTime;
  const traceDuration = trace.totalDuration;

  return (
    <div className="w-full h-full flex flex-col bg-background/50">
      {/* Header / Timeline Axis */}
      <div className="flex h-8 border-b items-center text-xs text-muted-foreground px-4 bg-muted/20">
         <div className="w-[200px] shrink-0 font-medium">Span</div>
         <div className="flex-1 relative h-full">
            {/* Simple axis markers */}
            <div className="absolute left-0 top-2 bottom-0 border-l border-border/50 pl-1">0ms</div>
            <div className="absolute left-1/4 top-2 bottom-0 border-l border-dashed border-border/30 pl-1">{Math.round(traceDuration * 0.25)}ms</div>
            <div className="absolute left-1/2 top-2 bottom-0 border-l border-dashed border-border/30 pl-1">{Math.round(traceDuration * 0.5)}ms</div>
            <div className="absolute left-3/4 top-2 bottom-0 border-l border-dashed border-border/30 pl-1">{Math.round(traceDuration * 0.75)}ms</div>
            <div className="absolute right-0 top-2 bottom-0 border-r border-border/50 pr-1 text-right">{traceDuration}ms</div>
         </div>
      </div>

      {/* Spans List */}
      <div className="flex-1 overflow-y-auto p-2 space-y-1">
        {spans.map(({ span, depth }) => {
          const relativeStart = span.startTime - traceStart;
          const duration = span.endTime - span.startTime;

          // Calculate positioning
          // Left offset %
          const leftPct = (relativeStart / traceDuration) * 100;
          // Width % (ensure at least a sliver is visible)
          const widthPct = Math.max((duration / traceDuration) * 100, 0.2);

          const isSelected = selectedSpanId === span.id;
          const isError = span.status === 'error';

          return (
            <div
                key={span.id}
                onClick={() => onSpanSelect(span)}
                className={cn(
                    "flex items-center h-8 rounded-sm cursor-pointer hover:bg-muted/50 transition-colors group",
                    isSelected && "bg-muted ring-1 ring-primary/20"
                )}
            >
               {/* Name Column */}
               <div className="w-[200px] shrink-0 flex items-center gap-2 px-2 overflow-hidden" style={{ paddingLeft: `${(depth * 12) + 8}px` }}>
                   <SpanIcon type={span.type} className={cn("h-3.5 w-3.5 shrink-0", isError ? "text-destructive" : "text-muted-foreground")} />
                   <span className={cn("text-xs truncate", isSelected ? "font-semibold text-primary" : "text-foreground")}>
                       {span.name}
                   </span>
               </div>

               {/* Timeline Bar Container */}
               <div className="flex-1 relative h-full mx-2">
                   <div
                        className={cn(
                            "absolute top-1.5 h-5 rounded-sm min-w-[2px] transition-all shadow-sm border border-transparent",
                            isError ? "bg-red-500/20 border-red-500/50" :
                            span.type === 'tool' ? "bg-amber-500/20 border-amber-500/50" :
                            span.type === 'service' ? "bg-blue-500/20 border-blue-500/50" :
                            "bg-primary/20 border-primary/30",
                            isSelected && "ring-2 ring-primary ring-offset-1 z-10"
                        )}
                        style={{
                            left: `${leftPct}%`,
                            width: `${widthPct}%`
                        }}
                   >
                        {/* Label inside bar if wide enough */}
                        {widthPct > 15 && (
                             <span className="absolute inset-0 flex items-center px-2 text-[10px] text-foreground/70 truncate">
                                 {duration}ms
                             </span>
                        )}
                   </div>

                   {/* Tooltip-ish duration on right if bar is small */}
                   {widthPct <= 15 && (
                       <span
                            className="absolute top-2 text-[10px] text-muted-foreground opacity-0 group-hover:opacity-100 transition-opacity ml-1"
                            style={{ left: `${leftPct + widthPct}%` }}
                        >
                            {duration}ms
                       </span>
                   )}
               </div>
            </div>
          );
        })}
      </div>
    </div>
  );
}
