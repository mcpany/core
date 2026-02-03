/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { InspectorTable } from "@/components/inspector/inspector-table";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { RefreshCcw, Bug, Unplug, Pause, Play, Trash2 } from "lucide-react";
import { useTraces } from "@/hooks/use-traces";
import { useSearchParams } from "next/navigation";
import { Suspense } from "react";

function InspectorContent() {
  const {
      traces,
      loading,
      isConnected,
      isPaused,
      setIsPaused,
      clearTraces,
      refresh
  } = useTraces();
  const searchParams = useSearchParams();
  const traceId = searchParams.get("traceId");

  return (
    <div className="flex flex-col h-[calc(100vh-4rem)] p-4 md:p-8 space-y-4">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold tracking-tight flex items-center gap-2">
            <Bug className="h-6 w-6" /> Inspector
          </h1>
          <div className="flex items-center gap-2 mt-1">
             <p className="text-muted-foreground">
                Debug JSON-RPC traffic and tool executions.
            </p>
            <Badge variant={isConnected ? "outline" : "destructive"} className="font-mono text-xs gap-1 ml-2">
                {isConnected ? (
                    <>
                    <span className="relative flex h-2 w-2">
                        <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-green-400 opacity-75"></span>
                        <span className="relative inline-flex rounded-full h-2 w-2 bg-green-500"></span>
                    </span>
                    Live
                    </>
                ) : (
                    <>
                    <Unplug className="h-3 w-3" /> Disconnected
                    </>
                )}
            </Badge>
          </div>
        </div>
        <div className="flex items-center gap-2">
             <Button
                variant="outline"
                size="sm"
                onClick={() => setIsPaused(!isPaused)}
            >
                {isPaused ? <><Play className="mr-2 h-4 w-4" /> Resume</> : <><Pause className="mr-2 h-4 w-4" /> Pause</>}
            </Button>
             <Button variant="outline" size="sm" onClick={clearTraces}>
                <Trash2 className="mr-2 h-4 w-4" /> Clear
            </Button>
            <Button variant="outline" size="sm" onClick={refresh} disabled={loading && !isConnected}>
            <RefreshCcw className={`mr-2 h-4 w-4 ${loading && !isConnected ? 'animate-spin' : ''}`} />
            Refresh
            </Button>
        </div>
      </div>

      <div className="flex-1 overflow-y-auto rounded-md border bg-card">
        <InspectorTable traces={traces} loading={loading && traces.length === 0} initialSelectedId={traceId} />
      </div>
    </div>
  );
}

/**
 * InspectorPage component.
 * @returns The rendered component.
 */
export default function InspectorPage() {
  return (
    <Suspense fallback={<div className="p-8">Loading inspector...</div>}>
      <InspectorContent />
    </Suspense>
  );
}
