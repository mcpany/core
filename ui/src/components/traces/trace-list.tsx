/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { Badge } from "@/components/ui/badge";
import { Card, CardContent } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { ScrollArea } from "@/components/ui/scroll-area";
import { cn } from "@/lib/utils";
import { Search, AlertCircle, CheckCircle2, Clock, Terminal, Globe, Database, User, Webhook as WebhookIcon } from "lucide-react";
import { Trace, SpanStatus } from "@/app/api/traces/route"; // Import type from route (or move types to shared)
import { formatDistanceToNow } from "date-fns";
import React, { memo, useMemo } from "react";

interface TraceListProps {
  traces: Trace[];
  selectedId: string | null;
  onSelect: (id: string) => void;
  searchQuery: string;
  onSearchChange: (query: string) => void;
}

// Optimization: Memoize TraceListItem to prevent re-renders of all items when one is selected.
// Only the selected and previously selected items will re-render.
const TraceListItem = memo(({ trace, isSelected, onSelect }: { trace: Trace, isSelected: boolean, onSelect: (id: string) => void }) => {
  return (
    <button
      onClick={() => onSelect(trace.id)}
      className={cn(
        "flex flex-col items-start gap-2 p-4 text-left text-sm transition-all hover:bg-accent/50 border-b last:border-0",
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
                <span>{trace.id}</span>
           </div>
           <span className="text-xs text-muted-foreground">
             {formatDistanceToNow(new Date(trace.timestamp), { addSuffix: true })}
           </span>
        </div>
      </div>
    </button>
  );
});
TraceListItem.displayName = "TraceListItem";

export function TraceList({ traces, selectedId, onSelect, searchQuery, onSearchChange }: TraceListProps) {

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
      <div className="p-4 border-b">
        <div className="relative">
          <Search className="absolute left-2 top-2.5 h-4 w-4 text-muted-foreground" />
          <Input
            placeholder="Search traces..."
            className="pl-8"
            value={searchQuery}
            onChange={(e) => onSearchChange(e.target.value)}
          />
        </div>
      </div>
      <ScrollArea className="flex-1">
        <div className="flex flex-col">
          {filteredTraces.length === 0 && (
             <div className="p-8 text-center text-muted-foreground text-sm">
                No traces found.
             </div>
          )}
          {filteredTraces.map((trace) => (
            <TraceListItem
              key={trace.id}
              trace={trace}
              isSelected={selectedId === trace.id}
              onSelect={onSelect}
            />
          ))}
        </div>
      </ScrollArea>
    </div>
  );
}

function StatusIcon({ status, className }: { status: SpanStatus, className?: string }) {
  if (status === 'error') return <AlertCircle className={cn("text-destructive", className)} />;
  if (status === 'success') return <CheckCircle2 className={cn("text-green-500", className)} />;
  return <Clock className={cn("text-muted-foreground", className)} />;
}

function TriggerIcon({ trigger, className }: { trigger: Trace['trigger'], className?: string }) {
    switch(trigger) {
        case 'user': return <User className={className} />;
        case 'webhook': return <WebhookIcon className={className} />;
        case 'system': return <Database className={className} />; // generic system
        default: return <Terminal className={className} />;
    }
}

function formatDuration(ms: number): string {
  if (ms < 1000) return `${ms}ms`;
  return `${(ms / 1000).toFixed(2)}s`;
}
