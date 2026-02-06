/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import React, { useState } from "react";
import dynamic from "next/dynamic";
import { Trace, SpanStatus } from "@/types/trace";
import {
  Table,
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

// ⚡ Bolt Optimization: Lazy load TableVirtuoso.
// Randomized Selection from Top 5 High-Impact Targets
const TableVirtuoso = dynamic(() => import("react-virtuoso").then((m) => m.TableVirtuoso), { ssr: false });

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
}

/**
 * Renders an icon representing the status of a trace span.
 */
function StatusIcon({ status, className }: { status: SpanStatus, className?: string }) {
  if (status === 'error') return <AlertCircle className={cn("text-destructive", className)} />;
  if (status === 'success') return <CheckCircle2 className={cn("text-green-500", className)} />;
  return <Clock className={cn("text-muted-foreground", className)} />;
}

/**
 * Renders an icon representing the type of a trace span.
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
 * ⚡ BOLT: Memoized cell content to prevent unnecessary re-renders.
 * Extracted from TraceRow to support virtualization where the Row wrapper is handled by Virtuoso.
 */
const TraceCells = React.memo(({ trace }: { trace: Trace }) => {
  return (
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
  );
});
TraceCells.displayName = 'TraceCells';

/**
 * A table component for displaying and inspecting traces.
 * Allows clicking on a row to view detailed trace information in a sheet.
 */
export function InspectorTable({ traces, loading }: InspectorTableProps) {
  const [selectedTrace, setSelectedTrace] = useState<Trace | null>(null);

  return (
    <>
      <div className="rounded-md border bg-card h-[500px]">
        {traces.length === 0 && !loading && (
             <div className="flex h-full items-center justify-center text-muted-foreground text-sm">
                 No traces found.
             </div>
        )}
        {loading && traces.length === 0 && (
             <div className="flex h-full items-center justify-center text-muted-foreground text-sm">
                 Loading traces...
             </div>
        )}

        {(traces.length > 0 || (loading && traces.length > 0)) && (
            <TableVirtuoso
                style={{ height: '100%' }}
                data={traces}
                components={{
                    Table: (props) => <Table {...props} style={{ ...props.style, width: '100%', borderCollapse: 'collapse' }} />,
                    TableBody: React.forwardRef((props, ref) => <TableBody {...props} ref={ref} />),
                    // @ts-expect-error - item is passed by Virtuoso but not in the standard TableRow props
                    TableRow: ({ item, ...props }) => (
                        <TableRow
                            {...props}
                            className={cn("cursor-pointer hover:bg-muted/50", props.className)}
                            onClick={(e) => {
                                if (props.onClick) props.onClick(e);
                                if (item) setSelectedTrace(item);
                            }}
                        />
                    ),
                }}
                fixedHeaderContent={() => (
                <TableRow className="hover:bg-transparent pointer-events-none">
                    <TableHead className="w-[180px] bg-card z-10">Timestamp</TableHead>
                    <TableHead className="w-[50px] bg-card z-10">Type</TableHead>
                    <TableHead className="bg-card z-10">Method / Name</TableHead>
                    <TableHead className="w-[100px] bg-card z-10">Status</TableHead>
                    <TableHead className="w-[100px] text-right bg-card z-10">Duration</TableHead>
                </TableRow>
                )}
                itemContent={(index, trace) => (
                    <TraceCells trace={trace} />
                )}
            />
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
