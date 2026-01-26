/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useEffect, useState, useRef } from "react";
import { InspectorTable } from "@/components/inspector/inspector-table";
import { Trace } from "@/types/trace";
import { Button } from "@/components/ui/button";
import { RefreshCcw, Bug, Unplug } from "lucide-react";
import { Badge } from "@/components/ui/badge";

/**
 * InspectorPage component.
 * @returns The rendered component.
 */
export default function InspectorPage() {
  const [traces, setTraces] = useState<Trace[]>([]);
  const [loading, setLoading] = useState(true);
  const [isConnected, setIsConnected] = useState(false);
  const wsRef = useRef<WebSocket | null>(null);

  const loadTraces = async () => {
      setLoading(true);
      try {
        const res = await fetch('/api/v1/traces');
        if (!res.ok) throw new Error("Failed to fetch");
        const data = await res.json();
        setTraces(data);
      } catch (err) {
        console.error("Failed to load traces", err);
      } finally {
        setLoading(false);
      }
  };

  const connect = () => {
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const host = window.location.host;
    const wsUrl = `${protocol}//${host}/api/v1/ws/traces`;

    const ws = new WebSocket(wsUrl);

    ws.onopen = () => {
      setIsConnected(true);
      setLoading(false);
    };

    ws.onmessage = (event) => {
      try {
        const newTrace: Trace = JSON.parse(event.data);
        setTraces((prev) => {
            // Deduplicate based on ID to avoid displaying the same trace twice
            // (e.g. if loaded via REST and then received via WS history)
            if (prev.some(t => t.id === newTrace.id)) return prev;
            return [newTrace, ...prev];
        });
      } catch (e) {
        console.error("Failed to parse trace", e);
      }
    };

    ws.onclose = () => {
      setIsConnected(false);
      // Attempt reconnection after 3 seconds
      setTimeout(connect, 3000);
    };

    wsRef.current = ws;
  };

  useEffect(() => {
    // We load via REST for immediate content (SSR/fast load) and then connect WS for updates
    loadTraces();
    connect();

    return () => {
      wsRef.current?.close();
    };
  }, []);

  return (
    <div className="flex flex-col h-[calc(100vh-4rem)] p-4 md:p-8 space-y-4">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold tracking-tight flex items-center gap-2">
            <Bug className="h-6 w-6" /> Inspector
          </h1>
          <p className="text-muted-foreground mt-1">
            Debug JSON-RPC traffic and tool executions.
          </p>
        </div>
        <div className="flex items-center gap-2">
           <Badge variant={isConnected ? "outline" : "destructive"} className="gap-1">
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
          <Button variant="outline" size="sm" onClick={loadTraces} disabled={loading}>
            <RefreshCcw className={`mr-2 h-4 w-4 ${loading ? 'animate-spin' : ''}`} />
            Refresh
          </Button>
        </div>
      </div>

      <div className="flex-1 overflow-y-auto">
        <InspectorTable traces={traces} loading={loading} />
      </div>
    </div>
  );
}
