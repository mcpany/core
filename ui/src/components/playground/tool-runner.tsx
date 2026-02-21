/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useMemo, useEffect } from "react";
import { Button } from "@/components/ui/button";
import { Textarea } from "@/components/ui/textarea";
import { Label } from "@/components/ui/label";
import { ToolDefinition, apiClient, ToolAnalytics } from "@/lib/client";
import { ScrollArea } from "@/components/ui/scroll-area";
import { PlayCircle, Loader2, Zap, Activity, History as HistoryIcon, RefreshCw, Code, Terminal } from "lucide-react";
import { Area, AreaChart, ResponsiveContainer, Tooltip as ChartTooltip, XAxis, YAxis, CartesianGrid } from "recharts";
import { cn } from "@/lib/utils";
import { Badge } from "@/components/ui/badge";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { SchemaViewer } from "@/components/tools/schema-viewer";
import { SchemaForm, Schema } from "@/components/tools/schema-form";
import { RichResultViewer } from "@/components/tools/rich-result-viewer";
import { Switch } from "@/components/ui/switch";
import { useToast } from "@/hooks/use-toast";
import { generateCurlCommand, generatePythonCode } from "@/lib/code-generator";
import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";

interface ToolRunnerProps {
  tool: ToolDefinition | null;
  onClose?: () => void;
  className?: string;
}

interface AuditLogEntry {
    timestamp: string;
    toolName: string;
    userId: string;
    profileId: string;
    arguments: string;
    result: string;
    error: string;
    duration: string;
    durationMs: number;
}

export function ToolRunner({ tool, onClose, className }: ToolRunnerProps) {
  const [input, setInput] = useState("{}");
  const [output, setOutput] = useState<any>(null);
  const [loading, setLoading] = useState(false);
  const [isDryRun, setIsDryRun] = useState(false);
  const { toast } = useToast();

  // Real data state
  const [historicalStats, setHistoricalStats] = useState<ToolAnalytics | null>(null);
  const [auditLogs, setAuditLogs] = useState<AuditLogEntry[]>([]);
  const [metricsLoading, setMetricsLoading] = useState(false);

  const parsedInput = useMemo(() => {
    try {
        return JSON.parse(input);
    } catch {
        return undefined;
    }
  }, [input]);

    // Computed stats from audit logs (recent 50)
  const recentStats = useMemo(() => {
      const total = auditLogs.length;
      const successes = auditLogs.filter(l => !l.error).length;
      const failures = total - successes;
      const avgLatency = total > 0 ? Math.round(auditLogs.reduce((acc, curr) => acc + curr.durationMs, 0) / total) : 0;

      // Map for chart
      // Sort by timestamp asc
      const sorted = [...auditLogs].sort((a, b) => new Date(a.timestamp).getTime() - new Date(b.timestamp).getTime());
      const chartData = sorted.map(l => ({
          time: new Date(l.timestamp).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', second: '2-digit' }),
          latency: l.durationMs,
          status: l.error ? "error" : "success"
      }));

      return { total, successes, failures, avgLatency, chartData };
  }, [auditLogs]);

  const fetchMetrics = async () => {
      if (!tool) return;
      setMetricsLoading(true);
      try {
          const [usage, logs] = await Promise.all([
              apiClient.getToolUsage(),
              apiClient.listAuditLogs({ tool_name: tool.name, limit: 50 })
          ]);
          // Find stats for this tool
          const stats = usage.find(u => u.name === tool.name && u.serviceId === tool.serviceId);
          setHistoricalStats(stats || null);
          setAuditLogs(logs.entries || []);
      } catch (e) {
          console.error("Failed to fetch metrics", e);
      } finally {
          setMetricsLoading(false);
      }
  };

  useEffect(() => {
      if (tool) {
          fetchMetrics();
          setInput("{}");
          setOutput(null);
      }
  }, [tool]);

  const handleExecute = async () => {
    if (!tool) return;
    setLoading(true);
    setOutput(null);
    try {
      const args = JSON.parse(input);
      const res = await apiClient.executeTool({
          name: tool.name,
          arguments: args
      }, isDryRun);
      setOutput(res);
      setTimeout(fetchMetrics, 500);
    } catch (e: any) {
      setOutput(`Error: ${e.message}`);
      setTimeout(fetchMetrics, 500);
    } finally {
      setLoading(false);
    }
  };

  const handleCopyCode = (type: 'curl' | 'python') => {
        if (!tool || !parsedInput) return;
        const baseUrl = typeof window !== 'undefined' ? window.location.origin : 'http://localhost:8080';
        const code = type === 'curl'
            ? generateCurlCommand({ toolName: tool.name, args: parsedInput, baseUrl })
            : generatePythonCode({ toolName: tool.name, args: parsedInput, baseUrl });

        navigator.clipboard.writeText(code);
        toast({ title: "Copied to clipboard", description: `${type === 'curl' ? 'Curl command' : 'Python code'} copied.` });
    };

  if (!tool) {
      return (
          <div className="flex items-center justify-center h-full text-muted-foreground flex-col gap-2">
              <Zap className="h-10 w-10 opacity-20" />
              <p>Select a tool to configure and execute.</p>
          </div>
      );
  }

  return (
    <div className={cn("flex flex-col h-full bg-background", className)}>
        <div className="flex items-center justify-between p-4 border-b bg-muted/20 sticky top-0 z-10 backdrop-blur">
             <div className="flex flex-col gap-1">
                 <div className="flex items-center gap-2">
                    <h2 className="text-lg font-semibold flex items-center gap-2">
                        {tool.name}
                    </h2>
                    <Badge variant="outline">{tool.serviceId}</Badge>
                 </div>
                 <p className="text-xs text-muted-foreground line-clamp-1">{tool.description}</p>
             </div>
             <div className="flex items-center gap-2">
                  <Switch id="dry-run" checked={isDryRun} onCheckedChange={setIsDryRun} className="scale-75" />
                  <Label htmlFor="dry-run" className="text-xs">Dry Run</Label>

                 <DropdownMenu>
                      <DropdownMenuTrigger asChild>
                          <Button variant="outline" size="sm" className="h-8 gap-1 ml-2">
                              <Code className="h-3 w-3" /> Code
                          </Button>
                      </DropdownMenuTrigger>
                      <DropdownMenuContent align="end">
                          <DropdownMenuItem onClick={() => handleCopyCode('curl')}>
                              <Terminal className="mr-2 h-3 w-3" /> Curl
                          </DropdownMenuItem>
                          <DropdownMenuItem onClick={() => handleCopyCode('python')}>
                              <Code className="mr-2 h-3 w-3" /> Python
                          </DropdownMenuItem>
                      </DropdownMenuContent>
                  </DropdownMenu>

                  <Button size="sm" onClick={handleExecute} disabled={loading} className="h-8 gap-1">
                    {loading ? <Loader2 className="h-3 w-3 animate-spin" /> : <PlayCircle className="h-3 w-3" />}
                    Run
                  </Button>
             </div>
        </div>

        <Tabs defaultValue="testing" className="flex-1 flex flex-col overflow-hidden">
            <div className="px-4 pt-2 border-b bg-muted/5">
                <TabsList className="grid w-[300px] grid-cols-2 h-8">
                    <TabsTrigger value="testing" className="text-xs">Test & Execute</TabsTrigger>
                    <TabsTrigger value="metrics" className="text-xs">Performance</TabsTrigger>
                </TabsList>
            </div>

            <TabsContent value="testing" className="flex-1 overflow-y-auto p-4 space-y-4 mt-0">
                 <div className="grid md:grid-cols-2 gap-4 h-full">
                    <div className="flex flex-col gap-2 h-full overflow-hidden">
                        <Label>Arguments</Label>
                         <Tabs defaultValue="form" className="flex-1 flex flex-col overflow-hidden border rounded-md">
                            <div className="bg-muted px-2 py-1 flex items-center justify-between border-b">
                                <TabsList className="h-6 bg-muted">
                                    <TabsTrigger value="form" className="text-[10px] h-5 px-2">Form</TabsTrigger>
                                    <TabsTrigger value="json" className="text-[10px] h-5 px-2">JSON</TabsTrigger>
                                </TabsList>
                            </div>
                            <TabsContent value="form" className="flex-1 overflow-y-auto p-4 mt-0 bg-muted/5">
                                {parsedInput === undefined ? (
                                    <div className="text-destructive text-xs">Invalid JSON</div>
                                ) : (
                                    <SchemaForm
                                        schema={tool.inputSchema as Schema}
                                        value={parsedInput}
                                        onChange={(v) => setInput(JSON.stringify(v, null, 2))}
                                    />
                                )}
                            </TabsContent>
                            <TabsContent value="json" className="flex-1 mt-0">
                                 <Textarea
                                    value={input}
                                    onChange={(e) => setInput(e.target.value)}
                                    className="font-mono text-xs h-full resize-none border-0 focus-visible:ring-0 rounded-none p-4"
                                />
                            </TabsContent>
                        </Tabs>
                    </div>

                    <div className="flex flex-col gap-2 h-full overflow-hidden">
                        <Label>Result</Label>
                        <div className="flex-1 border rounded-md bg-muted/10 overflow-hidden relative">
                             {output ? (
                                 <ScrollArea className="h-full">
                                     <div className="p-2">
                                        <RichResultViewer result={output} />
                                     </div>
                                 </ScrollArea>
                             ) : (
                                 <div className="absolute inset-0 flex items-center justify-center text-muted-foreground text-xs italic">
                                     Execute to see results...
                                 </div>
                             )}
                        </div>
                    </div>
                 </div>
            </TabsContent>

            <TabsContent value="metrics" className="flex-1 overflow-y-auto p-6 mt-0">
                 <div className="space-y-6">
                    <div className="flex justify-end">
                        <Button variant="ghost" size="sm" onClick={fetchMetrics} disabled={metricsLoading} className="h-8 gap-1">
                            <RefreshCw className={cn("h-3 w-3", metricsLoading && "animate-spin")} />
                            Refresh
                        </Button>
                    </div>
                    <div className="grid grid-cols-4 gap-4">
                        <div className="space-y-1">
                            <p className="text-[10px] uppercase text-muted-foreground font-semibold">Total Calls</p>
                            <p className="text-2xl font-bold">{historicalStats?.totalCalls ?? recentStats.total ?? 0}</p>
                        </div>
                        <div className="space-y-1">
                            <p className="text-[10px] uppercase text-muted-foreground font-semibold">Success Rate</p>
                            <p className="text-2xl font-bold text-green-500">{historicalStats?.successRate ?? 100}%</p>
                        </div>
                        <div className="space-y-1">
                            <p className="text-[10px] uppercase text-muted-foreground font-semibold">Avg Latency</p>
                            <p className="text-2xl font-bold">{recentStats.avgLatency}ms</p>
                        </div>
                        <div className="space-y-1">
                            <p className="text-[10px] uppercase text-muted-foreground font-semibold">Errors</p>
                            <p className="text-2xl font-bold text-destructive">{recentStats.failures}</p>
                        </div>
                    </div>

                    <div className="space-y-2">
                        <Label className="flex items-center gap-2">
                            <Activity className="h-4 w-4" /> Execution Latency (ms)
                        </Label>
                        <div className="h-[200px] w-full border rounded-md p-4 bg-muted/20">
                            {recentStats.chartData.length > 0 ? (
                                <ResponsiveContainer width="100%" height="100%">
                                    <AreaChart data={recentStats.chartData}>
                                        <CartesianGrid strokeDasharray="3 3" vertical={false} stroke="hsl(var(--muted))" />
                                        <XAxis dataKey="time" hide />
                                        <YAxis stroke="#888888" fontSize={10} tickLine={false} axisLine={false} />
                                        <ChartTooltip
                                            contentStyle={{ borderRadius: '8px', border: 'none', backgroundColor: 'hsl(var(--background))', boxShadow: '0 4px 6px -1px rgb(0 0 0 / 0.1)' }}
                                        />
                                        <Area type="monotone" dataKey="latency" stroke="hsl(var(--primary))" fill="hsl(var(--primary))" fillOpacity={0.1} />
                                    </AreaChart>
                                </ResponsiveContainer>
                            ) : (
                                 <div className="flex items-center justify-center h-full text-muted-foreground text-xs">
                                    No execution history available.
                                 </div>
                            )}
                        </div>
                    </div>
                </div>
            </TabsContent>
        </Tabs>

        {onClose && (
             <div className="p-4 border-t flex justify-end">
                <Button variant="outline" onClick={onClose}>Close</Button>
             </div>
        )}
    </div>
  );
}
