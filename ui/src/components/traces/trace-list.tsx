/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { ScrollArea } from "@/components/ui/scroll-area";
import { cn } from "@/lib/utils";
import { Search, AlertCircle, CheckCircle2, Clock, Terminal, Globe, Database, User, Webhook as WebhookIcon, Play, Pause, Filter } from "lucide-react";
import { Trace, SpanStatus } from "@/app/api/traces/route";
import { formatDistanceToNow } from "date-fns";
import React, { memo, useMemo, useState } from "react";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";
import { Label } from "@/components/ui/label";
import { Separator } from "@/components/ui/separator";

interface TraceListProps {
  traces: Trace[];
  selectedId: string | null;
  onSelect: (id: string) => void;
  searchQuery: string;
  onSearchChange: (query: string) => void;
  isLive: boolean;
  onToggleLive: (live: boolean) => void;
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
            <span className="font-semibold truncate max-w-[180px]" title={trace.rootSpan.name}>{trace.rootSpan.name}</span>
          </div>
          <span className="text-xs text-muted-foreground font-mono shrink-0">
            {formatDuration(trace.totalDuration)}
          </span>
        </div>

        <div className="flex items-center justify-between w-full mt-1">
           <div className="flex items-center gap-2 text-xs text-muted-foreground">
                <TriggerIcon trigger={trace.trigger} className="h-3 w-3" />
                <span className="truncate max-w-[120px]" title={trace.id}>{trace.id}</span>
           </div>
           <span className="text-xs text-muted-foreground shrink-0">
             {formatDistanceToNow(new Date(trace.timestamp), { addSuffix: true })}
           </span>
        </div>
      </div>
    </button>
  );
});
TraceListItem.displayName = "TraceListItem";

export function TraceList({ traces, selectedId, onSelect, searchQuery, onSearchChange, isLive, onToggleLive }: TraceListProps) {
  const [statusFilter, setStatusFilter] = useState<string>("all");
  const [methodFilter, setMethodFilter] = useState<string>("all");

  // Optimization: Memoize filtered traces to avoid re-calculating on every render
  const filteredTraces = useMemo(() => {
    const lowerQuery = searchQuery.toLowerCase();
    return traces.filter(t => {
      // Search Text
      const matchesSearch = t.rootSpan.name.toLowerCase().includes(lowerQuery) ||
                            t.id.toLowerCase().includes(lowerQuery);
      if (!matchesSearch) return false;

      // Status Filter
      if (statusFilter !== "all" && t.status !== statusFilter) return false;

      // Method Filter (Heuristic based on name "METHOD /path")
      if (methodFilter !== "all") {
         const method = t.rootSpan.name.split(' ')[0];
         if (method !== methodFilter) return false;
      }

      return true;
    });
  }, [traces, searchQuery, statusFilter, methodFilter]);

  // Extract unique methods for filter
  const methods = useMemo(() => {
      const s = new Set<string>();
      traces.forEach(t => {
          const method = t.rootSpan.name.split(' ')[0];
          if (method) s.add(method);
      });
      return Array.from(s).sort();
  }, [traces]);

  const activeFiltersCount = (statusFilter !== "all" ? 1 : 0) + (methodFilter !== "all" ? 1 : 0);

  return (
    <div className="flex flex-col h-full border-r bg-background/50 backdrop-blur-sm">
      <div className="flex flex-col border-b">
          <div className="p-3 flex items-center gap-2">
            <div className="relative flex-1">
              <Search className="absolute left-2 top-2.5 h-4 w-4 text-muted-foreground" />
              <Input
                placeholder="Search traces..."
                className="pl-8 h-9 text-sm"
                value={searchQuery}
                onChange={(e) => onSearchChange(e.target.value)}
              />
            </div>

            <Popover>
                <PopoverTrigger asChild>
                    <Button variant="outline" size="icon" className={cn("h-9 w-9 shrink-0", activeFiltersCount > 0 && "border-primary text-primary bg-primary/10")}>
                        <Filter className="h-4 w-4" />
                        {activeFiltersCount > 0 && (
                            <span className="absolute -top-1 -right-1 flex h-3 w-3 items-center justify-center rounded-full bg-primary text-[8px] text-primary-foreground">
                                {activeFiltersCount}
                            </span>
                        )}
                    </Button>
                </PopoverTrigger>
                <PopoverContent className="w-56 p-3" align="end">
                    <div className="space-y-3">
                        <div className="space-y-1">
                            <Label className="text-xs text-muted-foreground">Status</Label>
                            <Select value={statusFilter} onValueChange={setStatusFilter}>
                                <SelectTrigger className="h-8 text-xs">
                                    <SelectValue placeholder="All" />
                                </SelectTrigger>
                                <SelectContent>
                                    <SelectItem value="all">All Statuses</SelectItem>
                                    <SelectItem value="success">Success</SelectItem>
                                    <SelectItem value="error">Error</SelectItem>
                                    <SelectItem value="pending">Pending</SelectItem>
                                </SelectContent>
                            </Select>
                        </div>
                        <Separator />
                        <div className="space-y-1">
                            <Label className="text-xs text-muted-foreground">Method</Label>
                            <Select value={methodFilter} onValueChange={setMethodFilter}>
                                <SelectTrigger className="h-8 text-xs">
                                    <SelectValue placeholder="All" />
                                </SelectTrigger>
                                <SelectContent>
                                    <SelectItem value="all">All Methods</SelectItem>
                                    {methods.map(m => (
                                        <SelectItem key={m} value={m}>{m}</SelectItem>
                                    ))}
                                </SelectContent>
                            </Select>
                        </div>
                        {(statusFilter !== "all" || methodFilter !== "all") && (
                            <Button
                                variant="ghost"
                                size="sm"
                                className="w-full h-7 text-xs mt-2"
                                onClick={() => { setStatusFilter("all"); setMethodFilter("all"); }}
                            >
                                Clear Filters
                            </Button>
                        )}
                    </div>
                </PopoverContent>
            </Popover>

            <Button
                variant={isLive ? "default" : "outline"}
                size="icon"
                onClick={() => onToggleLive(!isLive)}
                title={isLive ? "Pause Live Updates" : "Start Live Updates"}
                className={cn("h-9 w-9 shrink-0", isLive && "bg-green-600 hover:bg-green-700")}
            >
                 {isLive ? <Pause className="h-4 w-4" /> : <Play className="h-4 w-4" />}
            </Button>
          </div>

          {/* Active Filter Chips could go here if we wanted to be fancy, but Popover is cleaner for now */}
      </div>

      <ScrollArea className="flex-1">
        <div className="flex flex-col">
          {filteredTraces.length === 0 && (
             <div className="flex flex-col items-center justify-center py-12 px-4 text-center text-muted-foreground gap-2">
                <Search className="h-8 w-8 opacity-20" />
                <p className="text-sm">No traces found.</p>
                {(searchQuery || statusFilter !== "all" || methodFilter !== "all") && (
                    <p className="text-xs opacity-70">Try adjusting your filters.</p>
                )}
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
        case 'system': return <Database className={className} />;
        default: return <Terminal className={className} />;
    }
}

function formatDuration(ms: number): string {
  if (ms < 1000) return `${ms}ms`;
  return `${(ms / 1000).toFixed(2)}s`;
}
