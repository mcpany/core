/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState } from "react";
import { TraceList } from "@/components/traces/trace-list";
import { TraceDetail } from "@/components/traces/trace-detail";
import { ResizableHandle, ResizablePanel, ResizablePanelGroup } from "@/components/ui/resizable";
import { Loader2 } from "lucide-react";
import { useTraces } from "@/hooks/use-traces";

/**
 * TracesPage component.
 * @returns The rendered component.
 */
export default function TracesPage() {
  const { traces, loading, isPaused, setIsPaused, refresh } = useTraces({ fetchHistory: true });
  const [selectedId, setSelectedId] = useState<string | null>(null);
  const [searchQuery, setSearchQuery] = useState("");

  const selectedTrace = traces.find(t => t.id === selectedId) || null;

  if (loading && traces.length === 0) {
      return (
          <div className="h-full flex items-center justify-center text-muted-foreground gap-2">
              <Loader2 className="h-6 w-6 animate-spin" /> Loading traces...
          </div>
      )
  }

  return (
    <div className="h-[calc(100vh-4rem)] overflow-hidden bg-background">
       <ResizablePanelGroup direction="horizontal">
        <ResizablePanel defaultSize={30} minSize={20} maxSize={40}>
           <TraceList
                traces={traces}
                selectedId={selectedId}
                onSelect={setSelectedId}
                searchQuery={searchQuery}
                onSearchChange={setSearchQuery}
                isLive={!isPaused}
                onToggleLive={(live) => setIsPaused(!live)}
                onRefresh={refresh}
            />
        </ResizablePanel>
        <ResizableHandle />
        <ResizablePanel defaultSize={70}>
            <TraceDetail trace={selectedTrace} />
        </ResizablePanel>
      </ResizablePanelGroup>
    </div>
  );
}
