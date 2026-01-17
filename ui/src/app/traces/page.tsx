/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useEffect, useState } from "react";
import { apiClient } from "@/lib/client";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { ScrollArea } from "@/components/ui/scroll-area";
import { PlayCircle, RefreshCw, Clock, ArrowRight, ArrowLeft } from "lucide-react";
import { formatDistanceToNow } from "date-fns";
import { useToast } from "@/hooks/use-toast";

// Interface for Debug Entry (from server/docs/features/debugger.md)
interface DebugEntry {
  id: string;
  timestamp: string;
  method: string;
  path: string;
  status: number;
  duration: number; // in nanoseconds? Doc says 15000000.
  request_headers: Record<string, string[]>;
  response_headers: Record<string, string[]>;
  request_body: string;
  response_body: string;
}

export default function TracesPage() {
  const [traces, setTraces] = useState<DebugEntry[]>([]);
  const [selectedTrace, setSelectedTrace] = useState<DebugEntry | null>(null);
  const [loading, setLoading] = useState(false);
  const { toast } = useToast();

  const fetchTraces = async () => {
    setLoading(true);
    try {
      const data = await apiClient.getDebugEntries();
      // Sort by timestamp desc
      const sorted = (Array.isArray(data) ? data : []).sort(
        (a: DebugEntry, b: DebugEntry) => new Date(b.timestamp).getTime() - new Date(a.timestamp).getTime()
      );
      setTraces(sorted);
    } catch (e) {
      console.error("Failed to fetch traces", e);
      toast({
        title: "Error",
        description: "Failed to load traces. Ensure the debugger middleware is enabled.",
        variant: "destructive",
      });
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchTraces();
    // Auto refresh every 5s? Maybe too aggressive for now.
  }, []);

  const handleReplay = (trace: DebugEntry) => {
      // For now, just copy to clipboard or show a toast
      navigator.clipboard.writeText(trace.request_body);
      toast({
          title: "Payload Copied",
          description: "Request body copied to clipboard. Paste in Playground to replay.",
      });
  };

  return (
    <div className="flex flex-col h-[calc(100vh-4rem)] p-4 gap-4">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold tracking-tight">Traffic Inspector</h1>
          <p className="text-muted-foreground">
            Inspect and replay agent traffic.
          </p>
        </div>
        <Button variant="outline" size="sm" onClick={fetchTraces} disabled={loading}>
          <RefreshCw className={`mr-2 h-4 w-4 ${loading ? "animate-spin" : ""}`} />
          Refresh
        </Button>
      </div>

      <div className="flex flex-1 gap-4 overflow-hidden">
        {/* Trace List */}
        <Card className="w-1/3 flex flex-col">
          <CardHeader className="p-4 border-b">
            <CardTitle className="text-sm font-medium">Request History</CardTitle>
          </CardHeader>
          <CardContent className="p-0 flex-1 overflow-hidden">
            <ScrollArea className="h-full">
              <div className="flex flex-col divide-y">
                {traces.length === 0 && !loading && (
                    <div className="p-4 text-center text-sm text-muted-foreground">No traces found.</div>
                )}
                {traces.map((trace) => (
                  <button
                    key={trace.id}
                    className={`flex flex-col items-start p-3 text-left hover:bg-muted/50 transition-colors ${
                      selectedTrace?.id === trace.id ? "bg-muted" : ""
                    }`}
                    onClick={() => setSelectedTrace(trace)}
                  >
                    <div className="flex items-center justify-between w-full mb-1">
                      <Badge variant={trace.status >= 400 ? "destructive" : "default"}>
                        {trace.method}
                      </Badge>
                      <span className="text-xs text-muted-foreground">
                        {formatDistanceToNow(new Date(trace.timestamp), { addSuffix: true })}
                      </span>
                    </div>
                    <div className="text-xs font-mono truncate w-full mb-1" title={trace.path}>
                      {trace.path}
                    </div>
                    <div className="flex items-center gap-2 text-xs text-muted-foreground">
                       <span className={`px-1 rounded ${trace.status >= 400 ? 'text-red-500' : 'text-green-500'}`}>
                           {trace.status}
                       </span>
                       <span>•</span>
                       <span>{(trace.duration / 1000000).toFixed(2)}ms</span>
                    </div>
                  </button>
                ))}
              </div>
            </ScrollArea>
          </CardContent>
        </Card>

        {/* Trace Details */}
        <Card className="flex-1 flex flex-col overflow-hidden">
          {selectedTrace ? (
            <>
              <CardHeader className="p-4 border-b flex flex-row items-center justify-between space-y-0">
                <div>
                   <CardTitle className="text-lg flex items-center gap-2">
                       {selectedTrace.method} <span className="text-muted-foreground font-normal">{selectedTrace.path}</span>
                   </CardTitle>
                   <CardDescription>
                       ID: {selectedTrace.id} • {new Date(selectedTrace.timestamp).toLocaleString()}
                   </CardDescription>
                </div>
                <Button size="sm" onClick={() => handleReplay(selectedTrace)}>
                    <PlayCircle className="mr-2 h-4 w-4" />
                    Replay
                </Button>
              </CardHeader>
              <CardContent className="flex-1 overflow-auto p-0">
                <div className="flex flex-col gap-0">
                    {/* Timeline Visualization Placeholder */}
                    <div className="bg-muted/30 p-4 border-b">
                         <div className="text-xs font-semibold mb-2">Timeline</div>
                         <div className="flex items-center gap-2 text-xs font-mono">
                             <div className="w-24 text-right text-muted-foreground">0ms</div>
                             <div className="flex-1 h-2 bg-gray-200 rounded overflow-hidden">
                                 <div className="h-full bg-blue-500" style={{ width: '100%' }}></div>
                             </div>
                             <div className="w-24">{(selectedTrace.duration / 1000000).toFixed(2)}ms</div>
                         </div>
                    </div>

                    <div className="grid grid-cols-2 h-full divide-x">
                        <div className="p-4 flex flex-col gap-4">
                            <h3 className="font-semibold text-sm flex items-center gap-2">
                                <ArrowRight className="h-4 w-4 text-blue-500" /> Request
                            </h3>
                            <div>
                                <div className="text-xs font-semibold mb-1">Headers</div>
                                <pre className="bg-muted p-2 rounded text-xs overflow-auto max-h-40">
                                    {JSON.stringify(selectedTrace.request_headers, null, 2)}
                                </pre>
                            </div>
                            <div className="flex-1">
                                <div className="text-xs font-semibold mb-1">Body</div>
                                <pre className="bg-muted p-2 rounded text-xs overflow-auto max-h-[300px] whitespace-pre-wrap font-mono">
                                    {selectedTrace.request_body || "<no body>"}
                                </pre>
                            </div>
                        </div>
                        <div className="p-4 flex flex-col gap-4">
                            <h3 className="font-semibold text-sm flex items-center gap-2">
                                <ArrowLeft className="h-4 w-4 text-green-500" /> Response
                            </h3>
                            <div>
                                <div className="text-xs font-semibold mb-1">Headers</div>
                                <pre className="bg-muted p-2 rounded text-xs overflow-auto max-h-40">
                                    {JSON.stringify(selectedTrace.response_headers, null, 2)}
                                </pre>
                            </div>
                            <div className="flex-1">
                                <div className="text-xs font-semibold mb-1">Body</div>
                                <pre className="bg-muted p-2 rounded text-xs overflow-auto max-h-[300px] whitespace-pre-wrap font-mono">
                                    {selectedTrace.response_body || "<no body>"}
                                </pre>
                            </div>
                        </div>
                    </div>
                </div>
              </CardContent>
            </>
          ) : (
            <div className="flex items-center justify-center h-full text-muted-foreground">
              Select a trace to view details
            </div>
          )}
        </Card>
      </div>
    </div>
  );
}
