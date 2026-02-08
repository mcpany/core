/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import React from "react";
import { Trace, SpanStatus } from "@/types/trace";
import {
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Badge } from "@/components/ui/badge";
import { AlertCircle, CheckCircle2, Clock, Terminal, Globe, Database } from "lucide-react";
import { cn } from "@/lib/utils";
import { formatDistanceToNow } from "date-fns";
import { TableVirtuoso } from "react-virtuoso";

interface TraceTableProps {
  traces: Trace[];
  loading?: boolean;
  onSelect?: (trace: Trace) => void;
}

function StatusIcon({ status, className }: { status: SpanStatus, className?: string }) {
  if (status === 'error') return <AlertCircle className={cn("text-destructive", className)} />;
  if (status === 'success') return <CheckCircle2 className={cn("text-green-500", className)} />;
  return <Clock className={cn("text-muted-foreground", className)} />;
}

function TypeIcon({ type, className }: { type: string, className?: string }) {
    switch(type) {
        case 'tool': return <Terminal className={className} />;
        case 'service': return <Globe className={className} />;
        case 'resource': return <Database className={className} />;
        default: return <Clock className={className} />;
    }
}

/**
 * TraceTable component.
 * @param props - The component props.
 * @param props.traces - The traces property.
 * @param props.loading - The loading property.
 * @param props.onSelect - The onSelect property.
 * @returns The rendered component.
 */
export function TraceTable({ traces, loading, onSelect }: TraceTableProps) {
  return (
    <div className="rounded-md border bg-card h-full w-full overflow-hidden">
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
                data={traces}
                components={{
                    Table: ({ style, ...props }) => (
                        <table {...props} style={{...style, width: '100%', borderCollapse: 'collapse'}} className="w-full caption-bottom text-sm" />
                    ),
                    TableHead: TableHeader,
                    TableBody: TableBody,
                    TableRow: ({ item, ...props }) => (
                        <TableRow
                            {...props}
                            className={cn("cursor-pointer hover:bg-muted/50", onSelect && "cursor-pointer")}
                            onClick={() => onSelect?.(item)}
                        />
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
                            <span className="text-xs text-muted-foreground font-mono">{trace.id.substring(0, 8)}...</span>
                        </div>
                    </TableCell>
                    <TableCell>
                        <Badge variant={trace.status === 'success' ? 'outline' : 'destructive'} className="gap-1">
                            <StatusIcon status={trace.status} className="h-3 w-3" />
                            {trace.status}
                        </Badge>
                    </TableCell>
                    <TableCell className="text-right font-mono text-xs">
                        {trace.totalDuration < 1000 ? `${Math.round(trace.totalDuration)}ms` : `${(trace.totalDuration / 1000).toFixed(2)}s`}
                    </TableCell>
                    </>
                )}
            />
        )}
    </div>
  );
}
