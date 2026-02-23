/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { InspectorTable } from "@/components/inspector/inspector-table";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { RefreshCcw, Bug, Unplug, Pause, Play, Trash2, Database, Zap } from "lucide-react";
import { useTraces } from "@/hooks/use-traces";
import { apiClient } from "@/lib/client";
import { useToast } from "@/hooks/use-toast";
import { useState } from "react";

/**
 * InspectorPage component.
 * @returns The rendered component.
 */
export default function InspectorPage() {
  const {
      traces,
      loading,
      isConnected,
      isPaused,
      setIsPaused,
      clearTraces,
      refresh
  } = useTraces();
  const { toast } = useToast();
  const [seeding, setSeeding] = useState(false);

  const handleSeedTrace = async () => {
      setSeeding(true);
      const now = Date.now();
      const trace = {
        id: "trace-seed-" + Math.floor(Math.random() * 10000),
        timestamp: new Date().toISOString(),
        totalDuration: 1250,
        status: "success",
        trigger: "user",
        rootSpan: {
            id: "span-1",
            name: "orchestrator-task",
            type: "core",
            startTime: now,
            endTime: now + 1250,
            status: "success",
            input: { query: "Analyze Q3 financial report", context: "user-session-123" },
            output: { summary: "Revenue up 15%", confidence: 0.98 },
            children: [
            {
                id: "span-2",
                name: "search-tool",
                type: "tool",
                startTime: now + 50,
                endTime: now + 450,
                status: "success",
                input: { query: "Q3 2024 financials" },
                output: { results: ["report_q3.pdf", "data_q3.xlsx"] },
                children: [
                    {
                        id: "span-2-1",
                        name: "google-search-api",
                        serviceName: "google",
                        type: "service",
                        startTime: now + 100,
                        endTime: now + 400,
                        status: "success",
                        input: { q: "Q3 2024 financials site:sec.gov" },
                        output: { items: [{ title: "10-Q", link: "..." }] }
                    }
                ]
            },
            {
                id: "span-3",
                name: "data-analyzer",
                type: "tool",
                startTime: now + 500,
                endTime: now + 1200,
                status: "success",
                input: { files: ["data_q3.xlsx"] },
                output: { analysis: "Growth detected", metrics: { revenue: 1.15 } },
                children: [
                    {
                        id: "span-3-1",
                        name: "python-interpreter",
                        serviceName: "local-python",
                        type: "service",
                        startTime: now + 550,
                        endTime: now + 1150,
                        status: "success",
                        input: { code: "import pandas as pd\ndf = pd.read_excel('data_q3.xlsx')\nprint(df.revenue.sum())" },
                        output: { stdout: "115000000" }
                    }
                ]
            }
            ]
        }
      };

      try {
          await apiClient.seedTrace(trace);
          toast({ title: "Trace Seeded", description: "Injected a complex test trace." });
          refresh();
      } catch (e) {
          toast({ title: "Seeding Failed", variant: "destructive", description: String(e) });
      } finally {
          setSeeding(false);
      }
  };

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
                onClick={handleSeedTrace}
                disabled={seeding}
                className="gap-2 hidden sm:flex"
            >
                <Zap className="h-4 w-4 text-amber-500" /> Seed Trace
            </Button>
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

      <div className="flex-1 min-h-0">
        <InspectorTable traces={traces} loading={loading && traces.length === 0} />
      </div>
    </div>
  );
}
