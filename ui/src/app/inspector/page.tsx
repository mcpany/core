/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useEffect, useState } from "react";
import { InspectorTable } from "@/components/inspector/inspector-table";
import { Trace } from "@/app/api/traces/route";
import { Button } from "@/components/ui/button";
import { RefreshCcw, Bug } from "lucide-react";

/**
 * InspectorPage component.
 * @returns The rendered component.
 */
export default function InspectorPage() {
  const [traces, setTraces] = useState<Trace[]>([]);
  const [loading, setLoading] = useState(true);

  const loadTraces = async () => {
      setLoading(true);
      try {
        const res = await fetch('/api/traces');
        const data = await res.json();
        setTraces(data);
      } catch (err) {
        console.error("Failed to load traces", err);
      } finally {
        setLoading(false);
      }
  };

  useEffect(() => {
    loadTraces();
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
        <Button variant="outline" size="sm" onClick={loadTraces} disabled={loading}>
          <RefreshCcw className={`mr-2 h-4 w-4 ${loading ? 'animate-spin' : ''}`} />
          Refresh
        </Button>
      </div>

      <div className="flex-1 overflow-y-auto">
        <InspectorTable traces={traces} loading={loading} />
      </div>
    </div>
  );
}
