/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */



import React, { useState, useMemo } from "react";
import { Trace, SpanStatus, Span } from "@/types/trace";
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
import { CheckCircle2, AlertCircle, Clock, Terminal, Globe, Database, ChevronRight, ChevronDown, Cpu, MessageSquare, Trash2 } from "lucide-react";
import { cn } from "@/lib/utils";
import { Checkbox } from "@/components/ui/checkbox";
import { Button } from "@/components/ui/button";
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
   * Callback to delete selected traces.
   */
  onDeleteTraces?: (ids: string[]) => Promise<void> | void;
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
        case 'core': return <Cpu className={className} />;
        case 'prompt': return <MessageSquare className={className} />;
        default: return <Clock className={className} />;
    }
}

interface VisibleRow {
  trace: Trace;
  span: Span;
  depth: number;
  hasChildren: boolean;
  isExpanded: boolean;
}

/**
 * Intent: Document InspectorTable
 *
 * Params:
 *   - Documented below.
 *
 * Returns:
 *   - Documented below.
 *
 * Errors:
 *   - None
 *
 * Side Effects:
 *   - None
 *
 * A table component for displaying and inspecting traces.
 * Allows clicking on a row to view detailed trace information in a sheet.
 *
 * @param props - The component props.
 * @param props.traces - The list of traces to display.
 * @param props.loading - Whether the data is loading.
 * @returns The rendered table component.
 */
export function InspectorTable({ traces, loading, onDeleteTraces }: InspectorTableProps) {
  const [selectedTrace, setSelectedTrace] = useState<Trace | null>(null);
  const [expandedSpans, setExpandedSpans] = useState<Set<string>>(new Set());
  const [selectedTraces, setSelectedTraces] = useState<Set<string>>(new Set());

  const handleSelectAll = (checked: boolean) => {
      if (checked) {
          setSelectedTraces(new Set(traces.map(t => t.id)));
      } else {
          setSelectedTraces(new Set());
      }
  };

  const handleSelectOne = (traceId: string, checked: boolean) => {
      setSelectedTraces(prev => {
          const newSelected = new Set(prev);
          if (checked) {
              newSelected.add(traceId);
          } else {
              newSelected.delete(traceId);
          }
          return newSelected;
      });
  };

  const [isDeleting, setIsDeleting] = useState(false);

  const handleBulkDelete = async () => {
      if (onDeleteTraces && selectedTraces.size > 0) {
          setIsDeleting(true);
          try {
              await onDeleteTraces(Array.from(selectedTraces));
              setSelectedTraces(new Set());
          } finally {
              setIsDeleting(false);
          }
      }
  };

  const isAllSelected = traces.length > 0 && selectedTraces.size === traces.length;

  const toggleExpand = (spanId: string) => {
      setExpandedSpans(prev => {
          const next = new Set(prev);
          if (next.has(spanId)) {
              next.delete(spanId);
          } else {
              next.add(spanId);
          }
          return next;
      });
  };

  const visibleRows = useMemo(() => {
      const rows: VisibleRow[] = [];

      const addSpan = (trace: Trace, span: Span, depth: number) => {
          const hasChildren = !!span.children && span.children.length > 0;
          const isExpanded = expandedSpans.has(span.id);
          rows.push({ trace, span, depth, hasChildren, isExpanded });

          if (isExpanded && hasChildren) {
              span.children!.forEach(child => addSpan(trace, child, depth + 1));
          }
      };

      traces.forEach(trace => addSpan(trace, trace.rootSpan, 0));
      return rows;
  }, [traces, expandedSpans]);

  return (
    <>
      {selectedTraces.size > 0 && (
          <div className="flex items-center gap-2 p-2 bg-muted/40 rounded-md animate-in fade-in slide-in-from-top-1 duration-200 mb-4 sticky top-0 z-10 backdrop-blur-md border">
              <span className="text-sm text-muted-foreground mr-2 font-medium px-2">{selectedTraces.size} selected</span>
              <div className="h-4 w-px bg-border mx-1" />
              <Button size="sm" variant="ghost" onClick={handleBulkDelete} disabled={isDeleting} className="h-8 text-red-600 hover:text-red-700 hover:bg-red-100 dark:hover:bg-red-900/20">
                  <Trash2 className={cn("mr-2 h-4 w-4", isDeleting ? "animate-pulse" : "")} />
                  {isDeleting ? "Deleting..." : "Delete"}
              </Button>
          </div>
      )}
      <div className="rounded-md border bg-card h-full w-full overflow-hidden">
        {/*
            ⚡ BOLT: Implemented virtualization for trace table using react-virtuoso.
            Randomized Selection from Top 5 High-Impact Targets
        */}
        {traces.length === 0 && !loading ? (
             <div className="flex items-center justify-center h-24 text-muted-foreground text-sm">
                No traces found.
             </div>
        ) : loading && traces.length === 0 ? (
             <div className="flex items-center justify-center h-24 text-muted-foreground text-sm">
                Loading traces...
             </div>
        ) : (
            <TableVirtuoso
                style={{ height: '100%', width: '100%' }}
                data={visibleRows}
                context={{ onClick: setSelectedTrace }}
                components={{
                    // Use shadcn/ui Table components where possible.
                    // Table: The root table element. shadcn Table is a wrapper. We need the table element.
                    Table: ({ style, ...props }) => (
                        <table {...props} style={{...style, width: '100%', borderCollapse: 'collapse'}} className="w-full caption-bottom text-sm" />
                    ),
                    TableHead: TableHeader,
                    TableBody: TableBody,
                    TableRow: ({ item, context, ...props }: any) => (
                        <TableRow {...props} className={cn("cursor-pointer hover:bg-muted/50", selectedTraces.has(item?.trace?.id) ? "bg-muted/50" : "")} onClick={() => context.onClick(item.trace)} />
                    ),
                }}
                fixedHeaderContent={() => (
                    <TableRow>
                    <TableHead className="w-[30px] pr-0 bg-card z-10">
                        <Checkbox
                            checked={isAllSelected}
                            onCheckedChange={(checked) => handleSelectAll(!!checked)}
                            aria-label="Select all"
                            className="translate-y-[2px]"
                        />
                    </TableHead>
                    <TableHead className="w-[180px] bg-card z-10">Timestamp</TableHead>
                    <TableHead className="w-[50px] bg-card z-10">Type</TableHead>
                    <TableHead className="bg-card z-10">Method / Name</TableHead>
                    <TableHead className="w-[100px] bg-card z-10">Status</TableHead>
                    <TableHead className="w-[100px] text-right bg-card z-10">Duration</TableHead>
                    </TableRow>
                )}
                itemContent={(index, row: VisibleRow) => (
                    <>
                    <TableCell className="pr-0" onClick={(e) => e.stopPropagation()}>
                        {row.depth === 0 ? (
                            <Checkbox
                                checked={selectedTraces.has(row.trace.id)}
                                onCheckedChange={(checked) => handleSelectOne(row.trace.id, !!checked)}
                                aria-label={`Select ${row.trace.id}`}
                                className="translate-y-[2px]"
                            />
                        ) : null}
                    </TableCell>
                    <TableCell className="font-mono text-xs text-muted-foreground">
                        {row.depth === 0 ? (
                            <>
                            {new Date(row.trace.timestamp).toLocaleTimeString()}
                            <br />
                            <span className="opacity-50 text-[10px]">
                                {formatDistanceToNow(new Date(row.trace.timestamp), { addSuffix: true })}
                            </span>
                            <br />
                            <span className="opacity-50 text-[10px] font-mono">{row.trace.id}</span>
                            </>
                        ) : null}
                    </TableCell>
                    <TableCell>
                        <TypeIcon type={row.span.type} className="h-4 w-4 text-muted-foreground" />
                    </TableCell>
                    <TableCell>
                        <div className="flex items-center gap-2" style={{ paddingLeft: `${row.depth * 1.5}rem` }}>
                            {row.hasChildren ? (
                                <div className="cursor-pointer p-1 hover:bg-muted rounded-md -ml-1" onClick={(e) => { e.stopPropagation(); toggleExpand(row.span.id); }}>
                                    {row.isExpanded ? <ChevronDown className="h-4 w-4" /> : <ChevronRight className="h-4 w-4" />}
                                </div>
                            ) : <div className="w-5" />}
                            <div className="flex flex-col">
                                <span className="font-medium">{row.span.name}</span>
                                <span className="text-xs text-muted-foreground font-mono">{row.span.id}</span>
                            </div>
                        </div>
                    </TableCell>
                    <TableCell>
                        <Badge variant={row.span.status === 'success' ? 'outline' : 'destructive'} className="gap-1">
                            <StatusIcon status={row.span.status} className="h-3 w-3" />
                            {row.span.status}
                        </Badge>
                    </TableCell>
                    <TableCell className="text-right font-mono text-xs">
                        {row.span.endTime - row.span.startTime < 1000 ? `${row.span.endTime - row.span.startTime}ms` : `${((row.span.endTime - row.span.startTime) / 1000).toFixed(2)}s`}
                    </TableCell>
                    </>
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
