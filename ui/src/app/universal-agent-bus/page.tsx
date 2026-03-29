/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import React, { useState, useEffect } from "react";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle
} from "@/components/ui/card";
import {
  GitMerge,
  Clock,
  Search,
  Network,
  Activity,
  Play
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { useToast } from "@/hooks/use-toast";
const fetchWrapper = fetch;

/**
 * Intent: Document UniversalAgentBusPage
 *
 * Params:
 *   - None
 *
 * Errors:
 *   - None
 *
 * UniversalAgentBusPage component for managing the Universal Agent Bus.
 *
 * Summary: Renders a dashboard with metrics for the Universal Agent Bus.
 *
 * Parameters:
 *   - None.
 *
 * Returns:
 *   - JSX.Element: The rendered layout for the agent bus dashboard.
 *
 * Errors/Throws:
 *   - None explicitly thrown by the component.
 *
 * Side Effects:
 *   - Renders multiple dashboard cards.
 */
export default function UniversalAgentBusPage() {
  const { toast } = useToast();
  const [adapters, setAdapters] = useState<string[]>([]);
  const [loadingAdapters, setLoadingAdapters] = useState(true);
  const [taskResult, setTaskResult] = useState<any>(null);
  const [taskLoading, setTaskLoading] = useState(false);
  const [intentInput, setIntentInput] = useState("adaptive_reasoning");

  useEffect(() => {
    const fetchAdapters = async () => {
      try {
        const response = await fetchWrapper("/api/v1/interop/adapters");
        if (response.ok) {
          const data = await response.json();
          setAdapters(data.adapters || []);
        } else {
          toast({
            title: "Error fetching adapters",
            description: "Failed to fetch adapters list.",
            variant: "destructive"
          });
        }
      } catch (e) {
        console.error(e);
      } finally {
        setLoadingAdapters(false);
      }
    };
    fetchAdapters();
  }, [toast]);

  const handleTestTask = async () => {
    setTaskLoading(true);
    setTaskResult(null);
    try {
      const response = await fetchWrapper("/api/v1/interop/task", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          id: "test-" + Date.now(),
          framework: "OpenClaw",
          intent: intentInput,
          payload: { test: "data" }
        })
      });
      if (response.ok) {
        const data = await response.json();
        setTaskResult(data);
        toast({
          title: "Task executed",
          description: "Task returned " + data.status,
        });
      } else {
        const errData = await response.text();
        toast({
          title: "Task failed",
          description: errData,
          variant: "destructive"
        });
      }
    } catch (e) {
      console.error(e);
      toast({
        title: "Task error",
        description: "Failed to execute task.",
        variant: "destructive"
      });
    } finally {
      setTaskLoading(false);
    }
  };

  return (
    <div className="p-6 space-y-6">
      <div>
        <h1 className="text-3xl font-bold tracking-tight">Universal Agent Bus</h1>
        <p className="text-muted-foreground mt-2">
          Manage and map subagents dynamically. Visualize, configure, and orchestrate interactions between complex AI swarms.
        </p>
      </div>

      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
        {/* Active Adapters */}
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Registered Hub Adapters</CardTitle>
            <GitMerge className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{loadingAdapters ? "..." : adapters.length}</div>
            <p className="text-xs text-muted-foreground">
              {adapters.join(", ")}
            </p>
          </CardContent>
        </Card>

        {/* Task Executor */}
        <Card className="md:col-span-2">
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Test Interop Task</CardTitle>
            <Play className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="flex space-x-2 mt-4">
              <Input
                value={intentInput}
                onChange={(e) => setIntentInput(e.target.value)}
                placeholder="Intent (e.g. adaptive_reasoning)"
              />
              <Button onClick={handleTestTask} disabled={taskLoading}>
                {taskLoading ? "Executing..." : "Execute Task"}
              </Button>
            </div>
            {taskResult && (
              <div className="p-4 bg-muted rounded-md overflow-x-auto text-xs font-mono">
                {JSON.stringify(taskResult, null, 2)}
              </div>
            )}
          </CardContent>
        </Card>

        {/* Multi-Agent Session Timeline */}
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Multi-Agent Session Timeline</CardTitle>
            <Clock className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">0 Sessions</div>
            <p className="text-xs text-muted-foreground">
              Visual tracking of agent handoffs and shared tool state.
            </p>
          </CardContent>
        </Card>

        {/* Unified Discovery Manager */}
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Unified Discovery Manager</CardTitle>
            <Network className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">0 Transports</div>
            <p className="text-xs text-muted-foreground">
              UI for managing and auto-discovering MCP servers across transports.
            </p>
          </CardContent>
        </Card>

        {/* Agent Chain Tracer (A2A) */}
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Agent Chain Tracer (A2A)</CardTitle>
            <Activity className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">0 Handoffs</div>
            <p className="text-xs text-muted-foreground">
              Visual timeline of multi-agent handoffs and message passing.
            </p>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
