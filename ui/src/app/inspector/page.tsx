/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useEffect, useState, useMemo } from "react";
import { InspectorTable } from "@/components/inspector/inspector-table";
import { Trace } from "@/types/trace";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Switch } from "@/components/ui/switch";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { RefreshCcw, Bug, Search, Filter } from "lucide-react";

/**
 * InspectorPage component.
 * @returns The rendered component.
 */
export default function InspectorPage() {
  const [traces, setTraces] = useState<Trace[]>([]);
  const [loading, setLoading] = useState(true);
  const [isLive, setIsLive] = useState(false);
  const [filterText, setFilterText] = useState("");
  const [filterStatus, setFilterStatus] = useState("all");

  const loadTraces = async (silent = false) => {
      if (!silent) setLoading(true);
      try {
        const res = await fetch('/api/traces');
        const data = await res.json();
        setTraces(data);
      } catch (err) {
        console.error("Failed to load traces", err);
      } finally {
        if (!silent) setLoading(false);
      }
  };

  useEffect(() => {
    loadTraces();
  }, []);

  // Polling effect
  useEffect(() => {
    let interval: NodeJS.Timeout;
    if (isLive) {
      interval = setInterval(() => {
        loadTraces(true);
      }, 2000);
    }
    return () => clearInterval(interval);
  }, [isLive]);

  // Filtering logic
  const filteredTraces = useMemo(() => {
    return traces.filter(trace => {
      const matchesText = filterText === "" ||
        trace.rootSpan.name.toLowerCase().includes(filterText.toLowerCase()) ||
        trace.id.includes(filterText);

      const matchesStatus = filterStatus === "all" ||
        (filterStatus === "success" && trace.status === "success") ||
        (filterStatus === "error" && trace.status === "error");

      return matchesText && matchesStatus;
    });
  }, [traces, filterText, filterStatus]);

  return (
    <div className="flex flex-col h-[calc(100vh-4rem)] p-4 md:p-8 space-y-4">
      <div className="flex flex-col gap-4">
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
             <div className="flex items-center gap-2 bg-muted/50 p-1.5 rounded-md border mr-2">
                <Switch
                  id="live-mode"
                  checked={isLive}
                  onCheckedChange={setIsLive}
                />
                <Label htmlFor="live-mode" className="text-sm font-medium cursor-pointer">
                  Live Mode
                </Label>
             </div>
            <Button variant="outline" size="sm" onClick={() => loadTraces()} disabled={loading && !isLive}>
              <RefreshCcw className={`mr-2 h-4 w-4 ${loading ? 'animate-spin' : ''}`} />
              Refresh
            </Button>
          </div>
        </div>

        {/* Toolbar */}
        <div className="flex items-center gap-4 bg-card p-3 rounded-md border shadow-sm">
            <div className="flex-1 max-w-sm relative">
                <Search className="absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" />
                <Input
                  placeholder="Filter by name or ID..."
                  className="pl-9 h-9"
                  value={filterText}
                  onChange={(e) => setFilterText(e.target.value)}
                />
            </div>

            <div className="flex items-center gap-2">
                <Filter className="h-4 w-4 text-muted-foreground" />
                <Select value={filterStatus} onValueChange={setFilterStatus}>
                  <SelectTrigger className="w-[140px] h-9">
                    <SelectValue placeholder="Status" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="all">All Status</SelectItem>
                    <SelectItem value="success">Success</SelectItem>
                    <SelectItem value="error">Error</SelectItem>
                  </SelectContent>
                </Select>
            </div>

            <div className="ml-auto text-xs text-muted-foreground">
                Showing {filteredTraces.length} of {traces.length} traces
            </div>
        </div>
      </div>

      <div className="flex-1 overflow-y-auto">
        <InspectorTable traces={filteredTraces} loading={loading && traces.length === 0} />
      </div>
    </div>
  );
}
