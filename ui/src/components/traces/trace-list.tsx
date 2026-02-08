/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { cn } from "@/lib/utils";
import { Search, AlertCircle, CheckCircle2, Clock, Terminal, Database, User, Webhook as WebhookIcon, Play, Pause, RefreshCw } from "lucide-react";
import { Trace, SpanStatus } from "@/types/trace";
import { formatDistanceToNow } from "date-fns";
import React, { memo, useMemo } from "react";
import { Virtuoso } from "react-virtuoso";

interface TraceListProps {
  traces: Trace[];
  selectedId: string | null;
  onSelect: (id: string) => void;
  searchQuery: string;
  onSearchChange: (query: string) => void;
  isPaused: boolean;
  onTogglePause: (paused: boolean) => void;
  onRefresh: () => void;
}

// Optimization: Memoize TraceListItem to prevent re-renders of all items when one is selected.
// Only the selected and previously selected items will re-render.
/**
 * TraceListItem component.
 * @param props - The component props.
 * @param props.trace - The trace property.
 * @param props.isSelected - The isSelected property.
 * @param props.onSelect - The onSelect property.
 * @returns The rendered component.
 */
const TraceListItem = memo(({ trace, isSelected, onSelect }: { trace: Trace, isSelected: boolean, onSelect: (id: string) => void }) => {
  return (
    <button
      onClick={() => onSelect(trace.id)}
      className={cn(
        "flex flex-col items-start gap-2 p-4 text-left text-sm transition-all hover:bg-accent/50 border-b last:border-0 w-full",
        isSelected && "bg-accent border-l-2 border-l-primary"
      )}
    >
      <div className="flex w-full flex-col gap-1">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <StatusIcon status={trace.status} className="h-4 w-4" />
            <span className="font-semibold">{trace.rootSpan.name}</span>
          </div>
          <span className="text-xs text-muted-foreground font-mono">
            {formatDuration(trace.totalDuration)}
          </span>
        </div>

        <div className="flex items-center justify-between w-full mt-1">
           <div className="flex items-center gap-2 text-xs text-muted-foreground">
                <TriggerIcon trigger={trace.trigger} className="h-3 w-3" />
                <span title={trace.id}>{trace.id.substring(0, 8)}...</span>
           </div>
           <span className="text-xs text-muted-foreground whitespace-nowrap">
             {formatDistanceToNow(new Date(trace.timestamp), { addSuffix: true })}
           </span>
        </div>
      </div>
    </button>
  );
});
TraceListItem.displayName = "TraceListItem";

/**
 * TraceList.
 *
 * @param props - The component props.
 * @returns The rendered component.
 */
export function TraceList({ traces, selectedId, onSelect, searchQuery, onSearchChange, isPaused, onTogglePause, onRefresh }: TraceListProps) {

  // Optimization: Memoize filtered traces to avoid re-calculating on every render,
  // especially when only selectedId changes.
  const filteredTraces = useMemo(() => {
    const lowerQuery = searchQuery.toLowerCase();
    return traces.filter(t =>
      t.rootSpan.name.toLowerCase().includes(lowerQuery) ||
      t.id.toLowerCase().includes(lowerQuery)
    );
  }, [traces, searchQuery]);

  return (
    <div className="flex flex-col h-full border-r bg-background/50 backdrop-blur-sm">
      <div className="p-4 border-b flex items-center gap-2">
        <div className="relative flex-1">
          <Search className="absolute left-2 top-2.5 h-4 w-4 text-muted-foreground" />
          <Input
            placeholder="Search traces..."
            className="pl-8"
            value={searchQuery}
            onChange={(e) => onSearchChange(e.target.value)}
          />
        </div>
        <div className="flex gap-1">
            <Button
                variant={isPaused ? "outline" : "secondary"}
                size="icon"
                onClick={() => onTogglePause(!isPaused)}
                title={isPaused ? "Resume Live Updates" : "Start Live Updates"}
                className={cn("shrink-0", !isPaused && "bg-green-100 dark:bg-green-900/20 text-green-600 dark:text-green-400 hover:bg-green-200 dark:hover:bg-green-900/40")}
            >
                {isPaused ? <Play className="h-4 w-4" /> : <Pause className="h-4 w-4" />}
            </Button>
            <Button
                variant="outline"
                size="icon"
                onClick={onRefresh}
                title="Refresh History"
            >
                <RefreshCw className="h-4 w-4" />
            </Button>
        </div>
      </div>
      <div className="flex-1 min-h-0">
        {filteredTraces.length === 0 ? (
           <div className="p-8 text-center text-muted-foreground text-sm">
              No traces found.
           </div>
        ) : (
          // ⚡ BOLT: Implemented virtualization for trace list using react-virtuoso.
          // Randomized Selection from Top 5 High-Impact Targets
          <Virtuoso
            style={{ height: '100%' }}
            data={filteredTraces}
            itemContent={(index, trace) => (
              <TraceListItem
                key={trace.id}
                trace={trace}
                isSelected={selectedId === trace.id}
                onSelect={onSelect}
              />
            )}
          />
        )}
      </div>
    </div>
  );
}

/**
 * StatusIcon component.
 * @param props - The component props.
 * @param props.status - The current status.
 * @param props.className - The name of the class.
 * @returns The rendered component.
 */
function StatusIcon({ status, className }: { status: SpanStatus, className?: string }) {
  if (status === 'error') return <AlertCircle className={cn("text-destructive", className)} />;
  if (status === 'success') return <CheckCircle2 className={cn("text-green-500", className)} />;
  return <Clock className={cn("text-muted-foreground", className)} />;
}

/**
 * TriggerIcon component.
 * @param props - The component props.
 * @param props.trigger - The trigger property.
 * @param props.className - The name of the class.
 * @returns The rendered component.
 */
function TriggerIcon({ trigger, className }: { trigger: Trace['trigger'], className?: string }) {
    switch(trigger) {
        case 'user': return <User className={className} />;
        case 'webhook': return <WebhookIcon className={className} />;
        case 'system': return <Database className={className} />; // generic system
        default: return <Terminal className={className} />;
    }
}

function formatDuration(ms: number): string {
  if (ms < 1000) return `${Math.round(ms)}ms`;
  return `${(ms / 1000).toFixed(2)}s`;
}
