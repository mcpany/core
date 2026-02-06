/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import React, { useState, useEffect } from "react";
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
   * Initial trace ID to select.
   */
  initialSelectedId?: string | null;
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
 * ⚡ BOLT: Memoized row component to prevent unnecessary re-renders when parent updates.
 * Randomized Selection from Top 5 High-Impact Targets
 */
const TraceRow = React.memo(({ trace, onClick }: { trace: Trace; onClick: (t: Trace) => void }) => {
  return (
    <TableRow
      className="cursor-pointer hover:bg-muted/50"
      onClick={() => onClick(trace)}
    >
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
    </TableRow>
  );
});
TraceRow.displayName = 'TraceRow';

/**
 * A table component for displaying and inspecting traces.
 * Allows clicking on a row to view detailed trace information in a sheet.
 *
 * @param props - The component props.
 * @param props.traces - The list of traces to display.
 * @param props.loading - Whether the data is loading.
 * @returns The rendered table component.
 */
export function InspectorTable({ traces, loading, initialSelectedId }: InspectorTableProps) {
  const [selectedTrace, setSelectedTrace] = useState<Trace | null>(null);

  // Auto-select trace if ID is provided and trace is found
  useEffect(() => {
    if (initialSelectedId && traces.length > 0) {
        // Only select if not already manually selected (unless we want to force it)
        // For deep linking, we want to force it once.
        const found = traces.find(t => t.id === initialSelectedId);
        if (found && (!selectedTrace || selectedTrace.id !== initialSelectedId)) {
            setSelectedTrace(found);
        }
    }
  }, [initialSelectedId, traces]); // Re-run when traces load

  return (
    <>
      <div className="rounded-md border bg-card">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead className="w-[180px]">Timestamp</TableHead>
              <TableHead className="w-[50px]">Type</TableHead>
              <TableHead>Method / Name</TableHead>
              <TableHead className="w-[100px]">Status</TableHead>
              <TableHead className="w-[100px] text-right">Duration</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {traces.length === 0 && !loading && (
              <TableRow>
                <TableCell colSpan={5} className="h-24 text-center text-muted-foreground">
                  No traces found.
                </TableCell>
              </TableRow>
            )}
            {loading && traces.length === 0 && (
                 <TableRow>
                    <TableCell colSpan={5} className="h-24 text-center text-muted-foreground">
                      Loading traces...
                    </TableCell>
                  </TableRow>
            )}
            {traces.map((trace) => (
              <TraceRow
                key={trace.id}
                trace={trace}
                onClick={setSelectedTrace}
              />
            ))}
          </TableBody>
        </Table>
      </div>

      <Sheet open={!!selectedTrace} onOpenChange={(open) => !open && setSelectedTrace(null)}>
        <SheetContent className="w-full sm:w-[800px] sm:max-w-[800px] p-0 overflow-y-auto border-l">
            {selectedTrace && <TraceDetail trace={selectedTrace} />}
        </SheetContent>
      </Sheet>
    </>
  );
}
