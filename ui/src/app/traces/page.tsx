/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useEffect } from "react";
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
  const {
    traces,
    loading,
    isConnected,
    isPaused,
    setIsPaused,
    refresh
  } = useTraces({ fetchHistory: true });

  const [selectedId, setSelectedId] = useState<string | null>(null);
  const [searchQuery, setSearchQuery] = useState("");

  // Select the first trace on initial load if none selected
  useEffect(() => {
    if (!selectedId && traces.length > 0 && !loading) {
       // Only select if it's the very first load to avoid jumping?
       // Let's just select the first one if nothing is selected.
       setSelectedId(traces[0].id);
    }
  }, [loading]); // Run once when loading finishes? Or when traces change?

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
                isPaused={isPaused}
                onTogglePaused={setIsPaused}
                onRefresh={refresh}
                isConnected={isConnected}
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
