/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import React, { useState } from "react";
import { Trace, SpanStatus } from "@/types/trace";
import {
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import {
  Sheet,
  SheetContent,
} from "@/components/ui/sheet";
import { Badge } from "@/components/ui/badge";
import { TraceDetail } from "@/components/traces/trace-detail";
import { CheckCircle2, AlertCircle, Clock, Terminal, Globe, Database } from "lucide-react";
import { cn } from "@/lib/utils";
import { formatDistanceToNow } from "date-fns";
import { TableVirtuoso } from "react-virtuoso";

/**
 * Props for the InspectorTable component.
 */
interface InspectorTableProps {
  /**
   * List of traces to display in the table.
   */
  traces: Trace[];
  /**
   * Whether the table is currently loading data.
   */
  loading?: boolean;
  /**
   * Optional class name for the container.
   */
  className?: string;
}

/**
 * Renders an icon representing the status of a trace span.
 *
 * @param props - The component props.
 * @param props.status - The status of the span (e.g., 'success', 'error').
 * @param props.className - Optional CSS classes.
 * @returns The status icon component.
 */
function StatusIcon({ status, className }: { status: SpanStatus, className?: string }) {
  if (status === 'error') return <AlertCircle className={cn("text-destructive", className)} />;
  if (status === 'success') return <CheckCircle2 className={cn("text-green-500", className)} />;
  return <Clock className={cn("text-muted-foreground", className)} />;
}

/**
 * Renders an icon representing the type of a trace span.
 *
 * @param props - The component props.
 * @param props.type - The type of the span (e.g., 'tool', 'service', 'resource').
 * @param props.className - Optional CSS classes.
 * @returns The type icon component.
 */
function TypeIcon({ type, className }: { type: string, className?: string }) {
    switch(type) {
        case 'tool': return <Terminal className={className} />;
        case 'service': return <Globe className={className} />;
        case 'resource': return <Database className={className} />;
        default: return <Clock className={className} />;
    }
}

/**
 * A table component for displaying and inspecting traces.
 * Allows clicking on a row to view detailed trace information in a sheet.
 *
 * @param props - The component props.
 * @param props.traces - The list of traces to display.
 * @param props.loading - Whether the data is loading.
 * @returns The rendered table component.
 */
export function InspectorTable({ traces, loading, className }: InspectorTableProps) {
  const [selectedTrace, setSelectedTrace] = useState<Trace | null>(null);

  return (
    <>
      <div className={cn("h-full w-full rounded-md border bg-card relative overflow-hidden", className)}>
        {/* ⚡ BOLT: Implemented virtualization using TableVirtuoso to handle large trace lists efficiently.
            Randomized Selection from Top 5 High-Impact Targets */}
        <TableVirtuoso
            style={{ height: '100%' }}
            data={traces}
            components={{
                // Mimic shadcn Table styling
                Table: ({ style, ...props }) => (
                    <table {...props} style={{ ...style, width: '100%', borderCollapse: 'collapse' }} className="w-full caption-bottom text-sm" />
                ),
                TableHead: TableHeader,
                TableBody: TableBody,
                // @ts-expect-error - TableVirtuoso passes 'item' prop which TableRow doesn't strictly expect in types but accepts in JS/any
                TableRow: ({ item, ...props }) => (
                     <TableRow {...props} className="cursor-pointer hover:bg-muted/50" onClick={() => setSelectedTrace(item)} />
                ),
            }}
            fixedHeaderContent={() => (
                <TableRow>
                  <TableHead className="w-[180px] bg-card z-10">Timestamp</TableHead>
                  <TableHead className="w-[50px] bg-card z-10">Type</TableHead>
                  <TableHead className="bg-card z-10">Method / Name</TableHead>
                  <TableHead className="w-[100px] bg-card z-10">Status</TableHead>
                  <TableHead className="w-[100px] text-right bg-card z-10">Duration</TableHead>
                </TableRow>
            )}
            itemContent={(index, trace) => (
                <>
                  <TableCell className="font-mono text-xs text-muted-foreground">
                    {new Date(trace.timestamp).toLocaleTimeString()}
                    <br />
                    <span className="opacity-50 text-[10px]">
                        {formatDistanceToNow(new Date(trace.timestamp), { addSuffix: true })}
                    </span>
                  </TableCell>
                  <TableCell>
                      <TypeIcon type={trace.rootSpan.type} className="h-4 w-4 text-muted-foreground" />
                  </TableCell>
                  <TableCell>
                      <div className="flex flex-col">
                          <span className="font-medium">{trace.rootSpan.name}</span>
                          <span className="text-xs text-muted-foreground font-mono">{trace.id}</span>
                      </div>
                  </TableCell>
                  <TableCell>
                      <Badge variant={trace.status === 'success' ? 'outline' : 'destructive'} className="gap-1">
                        <StatusIcon status={trace.status} className="h-3 w-3" />
                        {trace.status}
                      </Badge>
                  </TableCell>
                  <TableCell className="text-right font-mono text-xs">
                      {trace.totalDuration < 1000 ? `${trace.totalDuration}ms` : `${(trace.totalDuration / 1000).toFixed(2)}s`}
                  </TableCell>
                </>
            )}
        />

        {traces.length === 0 && !loading && (
             <div className="absolute inset-0 flex items-center justify-center pointer-events-none">
                <div className="text-muted-foreground">No traces found.</div>
             </div>
        )}
        {loading && traces.length === 0 && (
             <div className="absolute inset-0 flex items-center justify-center pointer-events-none bg-background/50">
                <div className="text-muted-foreground">Loading traces...</div>
             </div>
        )}
      </div>

      <Sheet open={!!selectedTrace} onOpenChange={(open) => !open && setSelectedTrace(null)}>
        <SheetContent className="w-full sm:w-[800px] sm:max-w-[800px] p-0 overflow-y-auto border-l">
            {selectedTrace && <TraceDetail trace={selectedTrace} />}
        </SheetContent>
      </Sheet>
    </>
  );
}
