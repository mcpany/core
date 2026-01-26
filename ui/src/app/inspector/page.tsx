/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useEffect, useState, useRef } from "react";
import { InspectorTable } from "@/components/inspector/inspector-table";
import { Trace } from "@/types/trace";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { RefreshCcw, Bug, Unplug, Pause, Play, Trash2 } from "lucide-react";

/**
 * InspectorPage component.
 * @returns The rendered component.
 */
export default function InspectorPage() {
  const [traces, setTraces] = useState<Trace[]>([]);
  const [loading, setLoading] = useState(true);
  const [isConnected, setIsConnected] = useState(false);
  const [isPaused, setIsPaused] = useState(false);
  const wsRef = useRef<WebSocket | null>(null);
  const isPausedRef = useRef(isPaused);

  useEffect(() => {
    isPausedRef.current = isPaused;
  }, [isPaused]);

  const connect = () => {
    setLoading(true);
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const host = window.location.host;
    const wsUrl = `${protocol}//${host}/api/v1/ws/traces`;

    // Cleanup previous
    if (wsRef.current) {
        wsRef.current.close();
    }

    const ws = new WebSocket(wsUrl);

    ws.onopen = () => {
      setIsConnected(true);
      setLoading(false);
    };

    ws.onmessage = (event) => {
        if (isPausedRef.current) return;
        try {
            const trace: Trace = JSON.parse(event.data);
            setTraces((prev) => {
                // Check if trace already exists (update) or new
                // Backend might stream history which overlaps with current if we reconnect
                // But generally history is sent once.
                // We prepend new traces to show newest at top

                // Deduplicate by ID
                const index = prev.findIndex(t => t.id === trace.id);
                if (index !== -1) {
                    const newTraces = [...prev];
                    newTraces[index] = trace;
                    return newTraces;
                }
                return [trace, ...prev];
            });
        } catch (e) {
            console.error("Failed to parse trace", e);
        }
    };

    ws.onclose = () => {
      setIsConnected(false);
      // Reconnect after 3s
      setTimeout(connect, 3000);
    };

    ws.onerror = (err) => {
      console.error("WebSocket error", err);
      ws.close();
    };

    wsRef.current = ws;
  };

  useEffect(() => {
    connect();
    return () => {
        wsRef.current?.close();
    };
  }, []);

  const clearTraces = () => setTraces([]);

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
            <Button variant="outline" size="sm" onClick={() => {
                setTraces([]);
                connect(); // Reconnect to refresh/fetch history
            }} disabled={loading && !isConnected}>
            <RefreshCcw className={`mr-2 h-4 w-4 ${loading && !isConnected ? 'animate-spin' : ''}`} />
            Refresh
            </Button>
        </div>
      </div>

      <div className="flex-1 overflow-y-auto rounded-md border bg-card">
        <InspectorTable traces={traces} loading={loading && traces.length === 0} />
      </div>
    </div>
  );
}
