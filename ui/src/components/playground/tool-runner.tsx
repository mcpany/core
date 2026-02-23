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
import { UniversalSchemaForm as SchemaForm, Schema } from "@/components/shared/universal-schema-form";
import { RichResultViewer } from "@/components/tools/rich-result-viewer";
import { Switch } from "@/components/ui/switch";
import { generateCurlCommand, generatePythonCode } from "@/lib/code-generator";
import { useToast } from "@/hooks/use-toast";
import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";

interface ToolRunnerProps {
  tool: ToolDefinition;
  onClose?: () => void;
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

/**
 * ToolRunner is a component that provides an interface for executing tools
 * and viewing their results, metrics, and schema.
 *
 * @param props - The props.
 * @param props.tool - The tool.
 * @param props.onClose - The onClose callback.
 * @returns The rendered component.
 */
export function ToolRunner({ tool, onClose }: ToolRunnerProps) {
  const [input, setInput] = useState("{}");
  const [output, setOutput] = useState<any>(null);
  const [loading, setLoading] = useState(false);
  const [isDryRun, setIsDryRun] = useState(false);
  const { toast } = useToast();

  // Real data state
  const [historicalStats, setHistoricalStats] = useState<ToolAnalytics | null>(null);
  const [auditLogs, setAuditLogs] = useState<AuditLogEntry[]>([]);
  const [metricsLoading, setMetricsLoading] = useState(false);

  // Reset state when tool changes
  useEffect(() => {
      setInput("{}");
      setOutput(null);
      setHistoricalStats(null);
      setAuditLogs([]);
      fetchMetrics();
  }, [tool.name]);

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
      const sorted = [...auditLogs].sort((a, b) => new Date(a.timestamp).getTime() - new Date(b.timestamp).getTime());
      const chartData = sorted.map(l => ({
          time: new Date(l.timestamp).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', second: '2-digit' }),
          latency: l.durationMs,
          status: l.error ? "error" : "success"
      }));

      return { total, successes, failures, avgLatency, chartData };
  }, [auditLogs]);

  const fetchMetrics = async () => {
      setMetricsLoading(true);
      try {
          const [usage, logs] = await Promise.all([
              apiClient.getToolUsage(),
              apiClient.listAuditLogs({ tool_name: tool.name, limit: 50 })
          ]);
          const stats = usage.find(u => u.name === tool.name && u.serviceId === tool.serviceId);
          setHistoricalStats(stats || null);
          setAuditLogs(logs.entries || []);
      } catch (e) {
          console.error("Failed to fetch metrics", e);
      } finally {
          setMetricsLoading(false);
      }
  };

  const handleExecute = async () => {
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
      setOutput({ error: e.message || String(e) });
      setTimeout(fetchMetrics, 500);
    } finally {
      setLoading(false);
    }
  };

  const handleCopyCode = (type: 'curl' | 'python') => {
        const data = parsedInput || {};
        const baseUrl = typeof window !== 'undefined' ? window.location.origin : 'http://localhost:8080';
        const code = type === 'curl'
            ? generateCurlCommand({ toolName: tool.name, args: data, baseUrl })
            : generatePythonCode({ toolName: tool.name, args: data, baseUrl });

        navigator.clipboard.writeText(code);
        toast({ title: "Copied to clipboard", description: `${type === 'curl' ? 'Curl command' : 'Python code'} copied.` });
    };

  return (
    <div className="flex flex-col h-full bg-background">
        <div className="px-6 py-4 border-b flex items-center justify-between sticky top-0 bg-background z-10">
             <div className="flex items-center gap-3">
                 <div className="bg-primary/10 p-2 rounded-md">
                    <Zap className="h-5 w-5 text-primary" />
                 </div>
                 <div>
                    <h2 className="text-lg font-semibold flex items-center gap-2">
                        {tool.name}
                        <Badge variant="outline" className="font-normal text-muted-foreground">{tool.serviceId || 'core'}</Badge>
                    </h2>
                    <p className="text-xs text-muted-foreground line-clamp-1 max-w-md" title={tool.description}>
                        {tool.description}
                    </p>
                 </div>
             </div>
             <div className="flex items-center gap-2">
                 <div className="flex items-center space-x-2 mr-4 border-r pr-4">
                      <Switch id="dry-run" checked={isDryRun} onCheckedChange={setIsDryRun} />
                      <Label htmlFor="dry-run" className="text-sm font-normal">Dry Run</Label>
                 </div>
                 <DropdownMenu>
                      <DropdownMenuTrigger asChild>
                          <Button variant="outline" size="sm" className="gap-2">
                              <Code className="size-4" />
                              Copy Code
                          </Button>
                      </DropdownMenuTrigger>
                      <DropdownMenuContent align="end">
                          <DropdownMenuItem onClick={() => handleCopyCode('curl')}>
                              <Terminal className="mr-2 size-4" />
                              Copy as Curl
                          </DropdownMenuItem>
                          <DropdownMenuItem onClick={() => handleCopyCode('python')}>
                              <Code className="mr-2 size-4" />
                              Copy as Python
                          </DropdownMenuItem>
                      </DropdownMenuContent>
                  </DropdownMenu>
                  <Button onClick={handleExecute} disabled={loading} size="sm">
                    {loading ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : <PlayCircle className="mr-2 h-4 w-4" />}
                    Execute
                  </Button>
                  {onClose && (
                      <Button variant="ghost" size="sm" onClick={onClose}>
                          Close
                      </Button>
                  )}
             </div>
        </div>

        <Tabs defaultValue="builder" className="flex-1 flex flex-col overflow-hidden">
             <div className="border-b px-6 bg-muted/5">
                <TabsList className="h-10 bg-transparent p-0">
                    <TabsTrigger value="builder" className="data-[state=active]:bg-transparent data-[state=active]:border-b-2 data-[state=active]:border-primary data-[state=active]:shadow-none rounded-none h-10 px-4">
                        Request Builder
                    </TabsTrigger>
                    <TabsTrigger value="metrics" className="data-[state=active]:bg-transparent data-[state=active]:border-b-2 data-[state=active]:border-primary data-[state=active]:shadow-none rounded-none h-10 px-4">
                        Metrics & History
                    </TabsTrigger>
                     <TabsTrigger value="schema" className="data-[state=active]:bg-transparent data-[state=active]:border-b-2 data-[state=active]:border-primary data-[state=active]:shadow-none rounded-none h-10 px-4">
                        Schema
                    </TabsTrigger>
                </TabsList>
             </div>

             <TabsContent value="builder" className="flex-1 overflow-hidden p-0 m-0 flex flex-col md:flex-row">
                 {/* Left/Top: Input */}
                 <div className="flex-1 flex flex-col border-b md:border-b-0 md:border-r min-h-[300px]">
                      <div className="p-3 border-b bg-muted/10 flex justify-between items-center">
                          <Label className="text-xs font-semibold text-muted-foreground uppercase">Arguments</Label>
                      </div>
                      <div className="flex-1 overflow-y-auto p-4">
                           <Tabs defaultValue="form" className="w-full">
                                <TabsList className="grid w-[200px] grid-cols-2 h-8 mb-4">
                                    <TabsTrigger value="form" className="text-xs">Form</TabsTrigger>
                                    <TabsTrigger value="json" className="text-xs">JSON</TabsTrigger>
                                </TabsList>
                                <TabsContent value="form" className="mt-0">
                                    {parsedInput === undefined ? (
                                        <div className="text-destructive text-sm p-4 border border-destructive/50 rounded bg-destructive/10">
                                            Invalid JSON. Please fix errors in JSON view to use the Form builder.
                                        </div>
                                    ) : (
                                        <SchemaForm
                                            schema={tool.inputSchema as Schema}
                                            value={parsedInput}
                                            onChange={(v) => setInput(JSON.stringify(v, null, 2))}
                                        />
                                    )}
                                </TabsContent>
                                <TabsContent value="json" className="mt-0">
                                     <Textarea
                                        value={input}
                                        onChange={(e) => setInput(e.target.value)}
                                        className="font-mono text-sm h-[400px]"
                                        placeholder="{}"
                                    />
                                </TabsContent>
                           </Tabs>
                      </div>
                 </div>

                 {/* Right/Bottom: Output */}
                 <div className="flex-1 flex flex-col bg-muted/5 min-h-[300px]">
                      <div className="p-3 border-b bg-muted/10 flex justify-between items-center">
                          <Label className="text-xs font-semibold text-muted-foreground uppercase">Result</Label>
                          {output && (
                              <Badge variant={output.isError || output.error ? "destructive" : "outline"} className={cn("text-[10px]", output.isError || output.error ? "" : "text-green-600 border-green-200 bg-green-50")}>
                                  {output.isError || output.error ? "Failed" : "Success"}
                              </Badge>
                          )}
                      </div>
                      <div className="flex-1 overflow-y-auto p-4">
                          {output ? (
                              <RichResultViewer result={output} />
                          ) : (
                              <div className="h-full flex flex-col items-center justify-center text-muted-foreground space-y-2 opacity-50">
                                  <PlayCircle className="h-12 w-12 stroke-[1]" />
                                  <p className="text-sm">Run the tool to see results here.</p>
                              </div>
                          )}
                      </div>
                 </div>
             </TabsContent>

             <TabsContent value="metrics" className="flex-1 overflow-y-auto p-6 m-0">
                <div className="flex justify-end mb-4">
                    <Button variant="ghost" size="sm" onClick={fetchMetrics} disabled={metricsLoading} className="h-8 gap-1">
                        <RefreshCw className={cn("h-3 w-3", metricsLoading && "animate-spin")} />
                        Refresh
                    </Button>
                </div>
                <div className="grid grid-cols-4 gap-4 mb-8">
                    <div className="space-y-1 p-4 border rounded-lg bg-card">
                        <p className="text-[10px] uppercase text-muted-foreground font-semibold">Total Calls</p>
                        <p className="text-2xl font-bold">{historicalStats?.totalCalls ?? recentStats.total ?? 0}</p>
                    </div>
                    <div className="space-y-1 p-4 border rounded-lg bg-card">
                        <p className="text-[10px] uppercase text-muted-foreground font-semibold">Success Rate</p>
                        <p className="text-2xl font-bold text-green-500">{historicalStats?.successRate ?? 100}%</p>
                    </div>
                    <div className="space-y-1 p-4 border rounded-lg bg-card">
                        <p className="text-[10px] uppercase text-muted-foreground font-semibold">Avg Latency (50)</p>
                        <p className="text-2xl font-bold">{recentStats.avgLatency}ms</p>
                    </div>
                    <div className="space-y-1 p-4 border rounded-lg bg-card">
                        <p className="text-[10px] uppercase text-muted-foreground font-semibold">Error Count (50)</p>
                        <p className="text-2xl font-bold text-destructive">{recentStats.failures}</p>
                    </div>
                </div>

                <div className="grid gap-6">
                    <div className="space-y-2">
                        <Label className="flex items-center gap-2 font-semibold">
                            <Activity className="h-4 w-4" /> Execution Latency (ms)
                        </Label>
                        <div className="h-[250px] w-full border rounded-md p-4 bg-card">
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

                    <div className="space-y-2">
                        <Label className="flex items-center gap-2 font-semibold">
                            <HistoryIcon className="h-4 w-4" /> Recent Timeline
                        </Label>
                        <div className="rounded-md border divide-y">
                            {recentStats.chartData.length === 0 && (
                                <div className="text-xs text-muted-foreground p-4 text-center">
                                    No recent executions.
                                </div>
                            )}
                            {[...recentStats.chartData].reverse().slice(0, 10).map((h, i) => (
                                <div key={i} className="flex items-center justify-between text-sm p-3 hover:bg-muted/50 transition-colors">
                                    <div className="flex items-center gap-3">
                                        <div className={cn("h-2.5 w-2.5 rounded-full", h.status === "success" ? "bg-green-500" : "bg-destructive")} />
                                        <span className="font-medium">{h.time}</span>
                                    </div>
                                    <span className="text-muted-foreground font-mono text-xs">{h.latency}ms</span>
                                </div>
                            ))}
                        </div>
                    </div>
                </div>
             </TabsContent>

             <TabsContent value="schema" className="flex-1 overflow-hidden p-6 m-0">
                  <Tabs defaultValue="visual" className="w-full h-full flex flex-col">
                      <TabsList className="w-[200px] mb-4">
                        <TabsTrigger value="visual">Visual</TabsTrigger>
                        <TabsTrigger value="json">JSON</TabsTrigger>
                      </TabsList>
                      <TabsContent value="visual" className="flex-1 overflow-hidden rounded-md border bg-muted/20 mt-0">
                         <ScrollArea className="h-full w-full p-4">
                            <SchemaViewer schema={tool.inputSchema as any} />
                         </ScrollArea>
                      </TabsContent>
                      <TabsContent value="json" className="flex-1 overflow-hidden rounded-md border bg-muted/50 mt-0">
                        <ScrollArea className="h-full w-full p-4">
                            <pre className="text-xs font-mono">{JSON.stringify(tool.inputSchema, null, 2)}</pre>
                        </ScrollArea>
                      </TabsContent>
                    </Tabs>
             </TabsContent>
        </Tabs>
    </div>
  );
}
