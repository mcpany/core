/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useEffect, useMemo } from "react";
import { TraceList } from "@/components/traces/trace-list";
import { TraceDetail } from "@/components/traces/trace-detail";
import { Trace } from "@/lib/trace-types";
import { Card, CardContent } from "@/components/ui/card";
import { Loader2 } from "lucide-react";

interface ServiceInspectorProps {
  serviceId: string;
  toolNames: string[];
}

export function ServiceInspector({ serviceId, toolNames }: ServiceInspectorProps) {
  const [traces, setTraces] = useState<Trace[]>([]);
  const [selectedId, setSelectedId] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [searchQuery, setSearchQuery] = useState("");
  const [isLive, setIsLive] = useState(true);

  // Poll for traces
  useEffect(() => {
    let isMounted = true;
    const fetchTraces = async () => {
      try {
        const res = await fetch('/api/traces');
        if (res.ok) {
          const data = await res.json();
          if (isMounted) {
            setTraces(data);
            setIsLoading(false);
          }
        }
      } catch (error) {
        console.error("Failed to fetch traces", error);
        if (isMounted) setIsLoading(false);
      }
    };

    fetchTraces(); // Initial fetch

    const interval = setInterval(() => {
      if (isLive) {
        fetchTraces();
      }
    }, 2000);

    return () => {
      isMounted = false;
      clearInterval(interval);
    };
  }, [isLive]);

  // Filter traces relevant to this service
  const filteredTraces = useMemo(() => {
    return traces.filter(trace => {
      // 1. Check if the tool name (in input params) matches one of our tools
      const inputName = trace.rootSpan.input?.params?.name;
      if (inputName && toolNames.includes(inputName)) {
        return true;
      }

      // 2. Check if the trace name contains the service name (heuristic)
      // This is useful if the tool name is namespaced like "service_tool"
      // or if the trace name is "POST /mcp/tools/call" but somehow tagged (less reliable without explicit tags)

      // 3. Check if the trace id matches selected (always show selected)
      if (trace.id === selectedId) return true;

      return false;
    });
  }, [traces, toolNames, selectedId]);

  const selectedTrace = useMemo(() =>
    traces.find(t => t.id === selectedId) || null,
  [traces, selectedId]);

  if (isLoading && traces.length === 0) {
    return (
      <div className="flex h-[400px] items-center justify-center text-muted-foreground">
        <Loader2 className="mr-2 h-4 w-4 animate-spin" /> Loading traces...
      </div>
    );
  }

  return (
    <Card className="h-[600px] border-none shadow-none bg-transparent">
      <CardContent className="p-0 h-full flex border rounded-md overflow-hidden bg-background">
        <div className="w-1/3 min-w-[300px] h-full">
          <TraceList
            traces={filteredTraces}
            selectedId={selectedId}
            onSelect={setSelectedId}
            searchQuery={searchQuery}
            onSearchChange={setSearchQuery}
            isLive={isLive}
            onToggleLive={setIsLive}
          />
        </div>
        <div className="flex-1 h-full border-l">
          <TraceDetail trace={selectedTrace} />
        </div>
      </CardContent>
    </Card>
  );
}
