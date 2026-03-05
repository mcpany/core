/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { InspectorTable } from "@/components/inspector/inspector-table";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Input } from "@/components/ui/input";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { RefreshCcw, Bug, Unplug, Pause, Play, Trash2, Zap, Search, Filter } from "lucide-react";
import { useTraces } from "@/hooks/use-traces";
import { apiClient } from "@/lib/client";
import { useToast } from "@/hooks/use-toast";
import { useState, useMemo } from "react";

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

  // Filter State
  const [searchQuery, setSearchQuery] = useState("");
  const [statusFilter, setStatusFilter] = useState("all");
  const [typeFilter, setTypeFilter] = useState("all");

  const handleSeedTrace = async () => {
      setSeeding(true);

      try {
          await apiClient.seedTrace();
          toast({ title: "Trace Seeded", description: "Injected a complex test trace." });
          refresh();
      } catch (e) {
          toast({ title: "Seeding Failed", variant: "destructive", description: String(e) });
      } finally {
          setSeeding(false);
      }
  };

  const filteredTraces = useMemo(() => {
      return traces.filter((trace) => {
        // Filter by Status
        if (statusFilter !== "all" && trace.status !== statusFilter) return false;

        // Filter by Type (root span type)
        if (typeFilter !== "all" && trace.rootSpan.type !== typeFilter) return false;

        // Filter by Search (ID or Name)
        if (searchQuery) {
            const query = searchQuery.toLowerCase();
            return (
                trace.id.toLowerCase().includes(query) ||
                trace.rootSpan.name.toLowerCase().includes(query)
            );
        }

        return true;
      });
  }, [traces, statusFilter, typeFilter, searchQuery]);

  return (
    <div className="flex flex-col h-[calc(100vh-4rem)] p-4 md:p-8 space-y-4">
      <div className="flex flex-col gap-4 md:flex-row md:items-center justify-between">
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
        <div className="flex items-center gap-2 flex-wrap md:flex-nowrap justify-end">
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

      {/* Filtering Toolbar */}
      <div className="flex flex-col md:flex-row gap-4 items-center bg-muted/20 p-2 rounded-lg border border-muted/50">
          <div className="relative flex-1 w-full">
              <Search className="absolute left-2 top-2.5 h-4 w-4 text-muted-foreground" />
              <Input
                  placeholder="Search traces (ID, Name)..."
                  className="pl-8 w-full bg-background"
                  value={searchQuery}
                  onChange={(e) => setSearchQuery(e.target.value)}
              />
          </div>
          <div className="flex gap-2 w-full md:w-auto">
              <Select value={statusFilter} onValueChange={setStatusFilter}>
                  <SelectTrigger className="w-[140px] bg-background">
                      <Filter className="mr-2 h-4 w-4 text-muted-foreground" />
                      <SelectValue placeholder="Status" />
                  </SelectTrigger>
                  <SelectContent>
                      <SelectItem value="all">All Statuses</SelectItem>
                      <SelectItem value="success">Success</SelectItem>
                      <SelectItem value="error">Error</SelectItem>
                      <SelectItem value="pending">Pending</SelectItem>
                  </SelectContent>
              </Select>

              <Select value={typeFilter} onValueChange={setTypeFilter}>
                  <SelectTrigger className="w-[140px] bg-background">
                      <Filter className="mr-2 h-4 w-4 text-muted-foreground" />
                      <SelectValue placeholder="Type" />
                  </SelectTrigger>
                  <SelectContent>
                      <SelectItem value="all">All Types</SelectItem>
                      <SelectItem value="tool">Tool</SelectItem>
                      <SelectItem value="service">Service</SelectItem>
                      <SelectItem value="core">Core</SelectItem>
                      <SelectItem value="resource">Resource</SelectItem>
                  </SelectContent>
              </Select>
          </div>
          <div className="text-xs text-muted-foreground whitespace-nowrap px-2">
              {filteredTraces.length} / {traces.length}
          </div>
      </div>

      <div className="flex-1 min-h-0">
        <InspectorTable traces={filteredTraces} loading={loading && traces.length === 0} />
      </div>
    </div>
  );
}
