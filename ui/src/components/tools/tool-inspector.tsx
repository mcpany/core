/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useMemo, useEffect } from "react";
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Textarea } from "@/components/ui/textarea";
import { Label } from "@/components/ui/label";
import { ToolDefinition, apiClient } from "@/lib/client";
import { ScrollArea } from "@/components/ui/scroll-area";
import { PlayCircle, Loader2, Zap, BarChart3, Activity, History as HistoryIcon } from "lucide-react";
import { Area, AreaChart, ResponsiveContainer, Tooltip as ChartTooltip, XAxis, YAxis, CartesianGrid } from "recharts";
import { cn } from "@/lib/utils";
import { Badge } from "@/components/ui/badge";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { SchemaViewer } from "./schema-viewer";

import { Switch } from "@/components/ui/switch";

interface ToolInspectorProps {
  tool: ToolDefinition | null;
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

/**
 * ToolInspector.
 *
 * @param onOpenChange - The onOpenChange.
 */
export function ToolInspector({ tool, open, onOpenChange }: ToolInspectorProps) {
  const [input, setInput] = useState("{}");
  const [output, setOutput] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);
  const [loadingHistory, setLoadingHistory] = useState(false);
  const [isDryRun, setIsDryRun] = useState(false);
  const [execHistory, setExecHistory] = useState<{ time: string; latency: number; status: string }[]>([]);

  useEffect(() => {
      if (open && tool) {
          setLoadingHistory(true);
          apiClient.listAuditLogs({ tool_name: tool.name, limit: 50 })
              .then((res: any) => {
                  const entries = res.entries || [];
                  // Sort by timestamp ascending
                  entries.sort((a: any, b: any) => new Date(a.timestamp).getTime() - new Date(b.timestamp).getTime());

                  const history = entries.map((e: any) => ({
                      time: new Date(e.timestamp).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', second: '2-digit' }),
                      latency: e.durationMs || 0,
                      status: e.error ? "error" : "success"
                  }));
                  setExecHistory(history);
              })
              .catch(err => console.error("Failed to load history", err))
              .finally(() => setLoadingHistory(false));
      } else if (!open) {
          setExecHistory([]);
      }
  }, [open, tool]);

  const stats = useMemo(() => {
    const total = execHistory.length;
    const successes = execHistory.filter(h => h.status === "success").length;
    const failures = total - successes;
    const avgLatency = total > 0 ? Math.round(execHistory.reduce((acc, curr) => acc + curr.latency, 0) / total) : 0;
    const successRate = total > 0 ? Math.round((successes / total) * 100) : 100;

    return { total, successes, failures, avgLatency, successRate };
  }, [execHistory]);

  if (!tool) return null;


  const handleExecute = async () => {
    setLoading(true);
    setOutput(null);
    try {
      const args = JSON.parse(input);
      const start = Date.now();
      const res = await apiClient.executeTool({
          toolName: tool.name,
          arguments: args
      }, isDryRun);
      const duration = Date.now() - start;
      setOutput(JSON.stringify(res, null, 2));
      setExecHistory(prev => [...prev.slice(-19), {
        time: new Date().toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', second: '2-digit' }),
        latency: duration,
        status: "success"
      }]);
    } catch (e: any) {
      setOutput(`Error: ${e.message}`);
      setExecHistory(prev => [...prev.slice(-19), {
        time: new Date().toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', second: '2-digit' }),
        latency: 0,
        status: "error"
      }]);
    } finally {
      setLoading(false);
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[700px]">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
              {String(tool.name)}
              <Badge variant="outline">{String(tool.serviceId)}</Badge>
          </DialogTitle>
          <DialogDescription>
            {String(tool.description)}
          </DialogDescription>
        </DialogHeader>

        <Tabs defaultValue="testing" className="w-full">
            <TabsList className="grid w-full grid-cols-2 mb-4">
                <TabsTrigger value="testing" className="flex items-center gap-2">
                    <Zap className="h-4 w-4" /> Test & Execute
                </TabsTrigger>
                <TabsTrigger value="metrics" className="flex items-center gap-2">
                    <BarChart3 className="h-4 w-4" /> Performance & Analytics
                </TabsTrigger>
            </TabsList>

            <TabsContent value="testing" className="space-y-4">
                <div className="grid gap-2">
                    <Label>Schema</Label>
                    <Tabs defaultValue="visual" className="w-full">
                      <TabsList className="grid w-[200px] grid-cols-2 h-8">
                        <TabsTrigger value="visual" className="text-xs">Visual</TabsTrigger>
                        <TabsTrigger value="json" className="text-xs">JSON</TabsTrigger>
                      </TabsList>
                      <TabsContent value="visual" className="mt-2">
                         <ScrollArea className="h-[200px] w-full rounded-md border p-4 bg-muted/20">
                            <SchemaViewer schema={tool.inputSchema as any} />
                         </ScrollArea>
                      </TabsContent>
                      <TabsContent value="json" className="mt-2">
                        <ScrollArea className="h-[200px] w-full rounded-md border p-4 bg-muted/50">
                            <pre className="text-xs">{JSON.stringify(tool.inputSchema, null, 2)}</pre>
                        </ScrollArea>
                      </TabsContent>
                    </Tabs>
                </div>

                <div className="grid gap-2">
                    <Label htmlFor="args">Arguments (JSON)</Label>
                    <Textarea
                        id="args"
                        value={input}
                        onChange={(e) => setInput(e.target.value)}
                        className="font-mono text-sm"
                        rows={5}
                    />
                </div>

                {output && (
                     <div className="grid gap-2">
                        <Label>Result</Label>
                        <ScrollArea className="h-[150px] w-full rounded-md border p-4 bg-muted/50">
                            <pre className="text-xs text-green-600 dark:text-green-400 font-mono">{output}</pre>
                        </ScrollArea>
                    </div>
                )}
            </TabsContent>

            <TabsContent value="metrics" className="space-y-6">
                {loadingHistory && (
                    <div className="flex items-center gap-2 text-sm text-muted-foreground mb-4">
                        <Loader2 className="h-3 w-3 animate-spin" /> Loading historical data...
                    </div>
                )}
                <div className="grid grid-cols-4 gap-4">
                    <div className="space-y-1">
                        <p className="text-[10px] uppercase text-muted-foreground font-semibold">Total Calls</p>
                        <p className="text-2xl font-bold">{stats.total}</p>
                    </div>
                    <div className="space-y-1">
                        <p className="text-[10px] uppercase text-muted-foreground font-semibold">Success Rate</p>
                        <p className="text-2xl font-bold text-green-500">{stats.successRate}%</p>
                    </div>
                    <div className="space-y-1">
                        <p className="text-[10px] uppercase text-muted-foreground font-semibold">Avg Latency</p>
                        <p className="text-2xl font-bold">{stats.avgLatency}ms</p>
                    </div>
                    <div className="space-y-1">
                        <p className="text-[10px] uppercase text-muted-foreground font-semibold">Error Count</p>
                        <p className="text-2xl font-bold text-destructive">{stats.failures}</p>
                    </div>
                </div>

                <div className="space-y-2">
                    <Label className="flex items-center gap-2">
                        <Activity className="h-4 w-4" /> Execution Latency (ms)
                    </Label>
                    <div className="h-[200px] w-full border rounded-md p-4 bg-muted/20">
                        <ResponsiveContainer width="100%" height="100%">
                            <AreaChart data={execHistory}>
                                <CartesianGrid strokeDasharray="3 3" vertical={false} stroke="hsl(var(--muted))" />
                                <XAxis dataKey="time" hide />
                                <YAxis stroke="#888888" fontSize={10} tickLine={false} axisLine={false} />
                                <ChartTooltip
                                    contentStyle={{ borderRadius: '8px', border: 'none', backgroundColor: 'hsl(var(--background))', boxShadow: '0 4px 6px -1px rgb(0 0 0 / 0.1)' }}
                                />
                                <Area type="monotone" dataKey="latency" stroke="hsl(var(--primary))" fill="hsl(var(--primary))" fillOpacity={0.1} />
                            </AreaChart>
                        </ResponsiveContainer>
                    </div>
                </div>

                {execHistory.length > 0 && (
                    <div className="space-y-2">
                        <Label className="flex items-center gap-2">
                            <HistoryIcon className="h-4 w-4" /> Recent Timeline
                        </Label>
                        <div className="space-y-2">
                            {execHistory.slice().reverse().slice(0, 5).map((h, i) => (
                                <div key={i} className="flex items-center justify-between text-xs p-2 rounded border bg-muted/30">
                                    <div className="flex items-center gap-2">
                                        <div className={cn("h-2 w-2 rounded-full", h.status === "success" ? "bg-green-500" : "bg-destructive")} />
                                        <span className="font-medium">{h.time}</span>
                                    </div>
                                    <span className="text-muted-foreground">{h.latency}ms</span>
                                </div>
                            ))}
                        </div>
                    </div>
                )}
            </TabsContent>
        </Tabs>

        <DialogFooter className="flex justify-between items-center sm:justify-between">
          <div className="flex items-center space-x-2">
              <Switch id="dry-run" checked={isDryRun} onCheckedChange={setIsDryRun} />
              <Label htmlFor="dry-run">Dry Run</Label>
          </div>
          <div className="flex gap-2">
              <Button variant="secondary" onClick={() => onOpenChange(false)}>Close</Button>
              <Button onClick={handleExecute} disabled={loading}>
                {loading ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : <PlayCircle className="mr-2 h-4 w-4" />}
                Execute
              </Button>
          </div>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
