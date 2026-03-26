/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */



<<<<<<< HEAD
import { useEffect, useState, useCallback } from "react";
=======
import { useEffect, useState } from "react";
>>>>>>> 4f039895e (⚡ Bolt: Render Optimization for System Status Banner (#6544))
import { TraceList } from "@/components/traces/trace-list";
import { TraceDetail } from "@/components/traces/trace-detail";
import type { Trace } from "@/types/trace";
import { ResizableHandle, ResizablePanel, ResizablePanelGroup } from "@/components/ui/resizable";
import { Loader2 } from "lucide-react";
<<<<<<< HEAD
import { usePolling } from "@/hooks/use-polling";
=======
>>>>>>> 4f039895e (⚡ Bolt: Render Optimization for System Status Banner (#6544))

/**
 * TracesPage component.
 * @returns The rendered component.
 */
export default function TracesPage() {
  const [traces, setTraces] = useState<Trace[]>([]);
  const [selectedId, setSelectedId] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  const [searchQuery, setSearchQuery] = useState("");
  const [isLive, setIsLive] = useState(false);

  // Separate load function for reuse
<<<<<<< HEAD
  const loadTraces = useCallback(async (isFirstLoad = false) => {
=======
  const loadTraces = async (isFirstLoad = false) => {
>>>>>>> 4f039895e (⚡ Bolt: Render Optimization for System Status Banner (#6544))
      try {
        const res = await fetch('/api/v1/traces');
        const data = await res.json();

        // If live, prepend new traces or merge?
        // For simplicity, we just replace since the API returns a fresh list.
        // In a real app we might want to merge.
        setTraces(data);

<<<<<<< HEAD
        if (isFirstLoad && data.length > 0) {
            // We only set selectedId if it's the *very first* load and nothing is selected
            // Use functional state update to avoid depending on selectedId in the useCallback
            setSelectedId(prev => prev === null ? data[0].id : prev);
=======
        if (isFirstLoad && data.length > 0 && !selectedId) {
            setSelectedId(data[0].id);
>>>>>>> 4f039895e (⚡ Bolt: Render Optimization for System Status Banner (#6544))
        }
      } catch (err) {
        console.error("Failed to load traces", err);
      } finally {
        if (isFirstLoad) setLoading(false);
      }
<<<<<<< HEAD
  }, []); // No dependencies needed anymore

  useEffect(() => {
    loadTraces(true);
  }, [loadTraces]);

  // ⚡ BOLT: [Render Optimization] Use custom usePolling hook for consistent lifecycle management and memory leak prevention.
  // Randomized Selection from Top 5 High-Impact Targets (Network Category)
  usePolling(() => loadTraces(false), isLive ? 3000 : null);
=======
  };

  useEffect(() => {
    loadTraces(true);
  }, []);

  useEffect(() => {
      let interval: NodeJS.Timeout;
      if (isLive) {
          interval = setInterval(() => {
              loadTraces(false);
          }, 3000);
      }
      return () => clearInterval(interval);
  }, [isLive]);
>>>>>>> 4f039895e (⚡ Bolt: Render Optimization for System Status Banner (#6544))

  const selectedTrace = traces.find(t => t.id === selectedId) || null;

  if (loading) {
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
                isLive={isLive}
                onToggleLive={setIsLive}
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
