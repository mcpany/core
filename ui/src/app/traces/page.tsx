/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useEffect, useState } from "react";
import { TraceList } from "@/components/traces/trace-list";
import { TraceDetail } from "@/components/traces/trace-detail";
import { Trace } from "@/app/api/traces/route";
import { ResizableHandle, ResizablePanel, ResizablePanelGroup } from "@/components/ui/resizable";
import { Loader2 } from "lucide-react";

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
  const loadTraces = async (isFirstLoad = false) => {
      try {
        // ⚡ Bolt Optimization: Request summary only for list to save bandwidth/CPU
        const res = await fetch('/api/traces?summary=true');
        const data = await res.json();

        setTraces(prev => {
             // Merge strategy to preserve full details if we already fetched them
             const traceMap = new Map(prev.map(t => [t.id, t]));
             return data.map((newTrace: Trace) => {
                 const oldTrace = traceMap.get(newTrace.id);
                 // If old trace has details (not summary) and new trace is summary, keep details
                 if (oldTrace && !oldTrace.isSummary && newTrace.isSummary) {
                     return {
                         ...newTrace,
                         isSummary: false, // Keep it as full trace
                         rootSpan: {
                             ...newTrace.rootSpan,
                             input: oldTrace.rootSpan.input,
                             output: oldTrace.rootSpan.output
                         }
                     };
                 }
                 return newTrace;
             });
        });

        if (isFirstLoad && data.length > 0 && !selectedId) {
            setSelectedId(data[0].id);
        }
      } catch (err) {
        console.error("Failed to load traces", err);
      } finally {
        if (isFirstLoad) setLoading(false);
      }
  };

  useEffect(() => {
    loadTraces(true);
  }, []);

  // ⚡ Bolt Optimization: Fetch full details only for the selected trace
  useEffect(() => {
      if (selectedId) {
         const trace = traces.find(t => t.id === selectedId);
         // If it's a summary trace, fetch full details
         if (trace && trace.isSummary) {
              fetch(`/api/traces/${selectedId}`)
                  .then(res => {
                      if (!res.ok) throw new Error("Failed to fetch trace details");
                      return res.json();
                  })
                  .then((fullTrace: Trace) => {
                      setTraces(prev => prev.map(t => t.id === fullTrace.id ? fullTrace : t));
                  })
                  .catch(err => console.error("Failed to load full trace", err));
         }
      }
  }, [selectedId, traces]);

  useEffect(() => {
      let interval: NodeJS.Timeout;
      if (isLive) {
          interval = setInterval(() => {
              loadTraces(false);
          }, 3000);
      }
      return () => clearInterval(interval);
  }, [isLive]);

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
