/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */



import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Checkbox } from "@/components/ui/checkbox";
import { cn } from "@/lib/utils";
import { Search, AlertCircle, CheckCircle2, Clock, Terminal, Database, User, Webhook as WebhookIcon, Play, Pause, Download } from "lucide-react";
import type { Trace, SpanStatus } from "@/types/trace";
import { formatDistanceToNow } from "date-fns";
import React, { memo, useMemo, useState } from "react";
import { Virtuoso } from "react-virtuoso";

interface TraceListProps {
  traces: Trace[];
  selectedId: string | null;
  onSelect: (id: string) => void;
  searchQuery: string;
  onSearchChange: (query: string) => void;
  isLive: boolean;
  onToggleLive: (live: boolean) => void;
  onBulkExport?: (selectedIds: string[]) => void;
}

// Optimization: Memoize TraceListItem to prevent re-renders of all items when one is selected.
// Only the selected and previously selected items will re-render.
/**
 * TraceListItem component.
 * @param props - The component props.
 * @param props.trace - The trace property.
 * @param props.isSelected - The isSelected property.
 * @param props.isChecked - Whether the item is checked for bulk actions.
 * @param props.onSelect - The onSelect property.
 * @param props.onCheckToggle - The onCheckToggle property.
 * @returns The rendered component.
 */
const TraceListItem = memo(({ trace, isSelected, isChecked, onSelect, onCheckToggle }: { trace: Trace, isSelected: boolean, isChecked: boolean, onSelect: (id: string) => void, onCheckToggle: (id: string) => void }) => {
  return (
    <div
      className={cn(
        "flex items-start gap-3 p-4 text-left text-sm transition-all hover:bg-accent/50 border-b last:border-0 w-full cursor-pointer",
        isSelected && "bg-accent border-l-2 border-l-primary"
      )}
      onClick={() => onSelect(trace.id)}
    >
      <div className="pt-1" onClick={(e) => e.stopPropagation()}>
        <Checkbox
          checked={isChecked}
          onCheckedChange={() => onCheckToggle(trace.id)}
          aria-label={`Select trace ${trace.id}`}
        />
      </div>
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
    </div>
  );
});
TraceListItem.displayName = "TraceListItem";

/**
 * TraceList.
 *
 * @param onToggleLive - The onToggleLive.
 */
export function TraceList({ traces, selectedId, onSelect, searchQuery, onSearchChange, isLive, onToggleLive, onBulkExport }: TraceListProps) {
  const [selectedBulkIds, setSelectedBulkIds] = useState<Set<string>>(new Set());

  // Optimization: Memoize filtered traces to avoid re-calculating on every render,
  // especially when only selectedId changes.
  const filteredTraces = useMemo(() => {
    const lowerQuery = searchQuery.toLowerCase();
    return traces.filter(t =>
      t.rootSpan.name.toLowerCase().includes(lowerQuery) ||
      t.id.toLowerCase().includes(lowerQuery)
    );
  }, [traces, searchQuery]);

  const toggleBulkSelection = (id: string) => {
    setSelectedBulkIds(prev => {
      const newSet = new Set(prev);
      if (newSet.has(id)) {
        newSet.delete(id);
      } else {
        newSet.add(id);
      }
      return newSet;
    });
  };

  const handleSelectAll = (checked: boolean | "indeterminate") => {
    if (checked) {
      setSelectedBulkIds(new Set(filteredTraces.map(t => t.id)));
    } else {
      setSelectedBulkIds(new Set());
    }
  };

  const allSelected = filteredTraces.length > 0 && selectedBulkIds.size === filteredTraces.length;

  return (
    <div className="flex flex-col h-full border-r bg-background/50 backdrop-blur-sm">
      <div className="p-4 border-b flex flex-col gap-2">
        <div className="flex items-center gap-2">
          <div className="relative flex-1">
            <Search className="absolute left-2 top-2.5 h-4 w-4 text-muted-foreground" />
            <Input
              placeholder="Search traces..."
              className="pl-8"
              value={searchQuery}
              onChange={(e) => onSearchChange(e.target.value)}
            />
          </div>
          <Button
              variant={isLive ? "default" : "outline"}
              size="icon"
              onClick={() => onToggleLive(!isLive)}
              title={isLive ? "Pause Live Updates" : "Start Live Updates"}
              className={cn("shrink-0", isLive && "bg-green-600 hover:bg-green-700")}
          >
               {isLive ? <Pause className="h-4 w-4" /> : <Play className="h-4 w-4" />}
          </Button>
        </div>

        {/* Bulk Actions Bar */}
        <div className="flex items-center justify-between py-1">
          <div className="flex items-center gap-2 text-sm text-muted-foreground">
            <Checkbox
              checked={allSelected}
              onCheckedChange={handleSelectAll}
              aria-label="Select all traces"
            />
            <span>{selectedBulkIds.size} selected</span>
          </div>
          {selectedBulkIds.size > 0 && onBulkExport && (
            <Button
              variant="secondary"
              size="sm"
              className="h-7 text-xs gap-1"
              onClick={() => onBulkExport(Array.from(selectedBulkIds))}
            >
              <Download className="h-3 w-3" />
              Export
            </Button>
          )}
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
                isChecked={selectedBulkIds.has(trace.id)}
                onSelect={onSelect}
                onCheckToggle={toggleBulkSelection}
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
  if (ms < 1000) return `${ms}ms`;
  return `${(ms / 1000).toFixed(2)}s`;
}
