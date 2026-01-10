/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import React, { useState, useEffect } from "react";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
} from "@/components/ui/sheet";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { ScrollArea } from "@/components/ui/scroll-area";
import { RefreshCw, Play, Pause, Search, X } from "lucide-react";
import { Input } from "@/components/ui/input";

// Types matching the backend DebugEntry
interface DebugEntry {
  id: string;
  timestamp: string;
  method: string;
  path: string;
  status: number;
  duration: number; // in nanoseconds
  request_headers: Record<string, string[]>;
  response_headers: Record<string, string[]>;
  request_body?: string;
  response_body?: string;
}

const TrafficInspector: React.FC = () => {
  const [entries, setEntries] = useState<DebugEntry[]>([]);
  const [selectedEntry, setSelectedEntry] = useState<DebugEntry | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const [autoRefresh, setAutoRefresh] = useState(false);
  const [filter, setFilter] = useState("");

  const fetchEntries = async () => {
    setIsLoading(true);
    try {
      const response = await fetch("/debug/entries");
      if (response.ok) {
        const data = await response.json();
        setEntries(data || []);
      }
    } catch (error) {
      console.error("Failed to fetch debug entries:", error);
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    fetchEntries();
  }, []);

  useEffect(() => {
    let interval: NodeJS.Timeout;
    if (autoRefresh) {
      interval = setInterval(fetchEntries, 2000);
    }
    return () => clearInterval(interval);
  }, [autoRefresh]);

  const formatDuration = (nanos: number) => {
    const ms = nanos / 1000000;
    if (ms < 1) return "< 1ms";
    return `${ms.toFixed(2)}ms`;
  };

  const getStatusColor = (status: number) => {
    if (status >= 500) return "destructive";
    if (status >= 400) return "destructive"; // Use warning if available, but shadcn usually has default/destructive/secondary/outline
    if (status >= 300) return "secondary";
    return "default"; // Success/200
  };

  const filteredEntries = entries.filter((entry) =>
    entry.path.toLowerCase().includes(filter.toLowerCase()) ||
    entry.method.toLowerCase().includes(filter.toLowerCase()) ||
    entry.id.includes(filter)
  );

  return (
    <div className="p-6 space-y-6">
      <div className="flex flex-col gap-4 md:flex-row md:items-center md:justify-between">
        <div>
          <h2 className="text-3xl font-bold tracking-tight">Traffic Inspector</h2>
          <p className="text-muted-foreground">
            Live inspection of HTTP traffic processed by the server.
          </p>
        </div>
        <div className="flex items-center gap-2">
           <div className="relative w-64">
            <Search className="absolute left-2 top-2.5 h-4 w-4 text-muted-foreground" />
            <Input
              placeholder="Filter by path, method..."
              value={filter}
              onChange={(e) => setFilter(e.target.value)}
              className="pl-8"
            />
          </div>
          <Button
            variant="outline"
            size="icon"
            onClick={() => setAutoRefresh(!autoRefresh)}
            className={autoRefresh ? "bg-green-100 dark:bg-green-900 border-green-500" : ""}
          >
            {autoRefresh ? <Pause className="h-4 w-4" /> : <Play className="h-4 w-4" />}
          </Button>
          <Button variant="outline" size="icon" onClick={fetchEntries} disabled={isLoading}>
            <RefreshCw className={`h-4 w-4 ${isLoading ? "animate-spin" : ""}`} />
          </Button>
        </div>
      </div>

      <Card>
        <CardContent className="p-0">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead className="w-[100px]">Time</TableHead>
                <TableHead className="w-[80px]">Method</TableHead>
                <TableHead className="w-[80px]">Status</TableHead>
                <TableHead>Path</TableHead>
                <TableHead className="w-[100px] text-right">Duration</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {filteredEntries.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={5} className="text-center h-24 text-muted-foreground">
                    No traffic recorded yet.
                  </TableCell>
                </TableRow>
              ) : (
                filteredEntries.map((entry) => (
                  <TableRow
                    key={entry.id}
                    className="cursor-pointer hover:bg-muted/50"
                    onClick={() => setSelectedEntry(entry)}
                  >
                    <TableCell className="font-mono text-xs whitespace-nowrap">
                      {new Date(entry.timestamp).toLocaleTimeString()}
                    </TableCell>
                    <TableCell>
                      <Badge variant="outline">{entry.method}</Badge>
                    </TableCell>
                    <TableCell>
                      <Badge variant={getStatusColor(entry.status)}>{entry.status}</Badge>
                    </TableCell>
                    <TableCell className="font-mono text-sm break-all">
                        {entry.path.length > 50 ? entry.path.substring(0, 50) + "..." : entry.path}
                    </TableCell>
                    <TableCell className="text-right font-mono text-xs">
                      {formatDuration(entry.duration)}
                    </TableCell>
                  </TableRow>
                ))
              )}
            </TableBody>
          </Table>
        </CardContent>
      </Card>

      <Sheet open={!!selectedEntry} onOpenChange={(open) => !open && setSelectedEntry(null)}>
        <SheetContent className="sm:max-w-2xl overflow-y-auto">
          <SheetHeader className="mb-6">
            <SheetTitle className="flex items-center gap-2">
              <Badge variant="outline">{selectedEntry?.method}</Badge>
              <span className="font-mono text-sm">{selectedEntry?.path}</span>
            </SheetTitle>
            <SheetDescription>
                ID: {selectedEntry?.id} <br/>
                Time: {selectedEntry && new Date(selectedEntry.timestamp).toLocaleString()} <br/>
                Duration: {selectedEntry && formatDuration(selectedEntry.duration)}
            </SheetDescription>
          </SheetHeader>

          {selectedEntry && (
            <div className="space-y-6">
              <div>
                <h3 className="text-sm font-semibold mb-2">Request</h3>
                <div className="space-y-2">
                  <details className="text-sm">
                    <summary className="cursor-pointer text-muted-foreground hover:text-foreground">Headers</summary>
                    <div className="mt-2 bg-muted p-2 rounded-md font-mono text-xs overflow-x-auto">
                        {Object.entries(selectedEntry.request_headers).map(([key, val]) => (
                            <div key={key}>
                                <span className="font-semibold">{key}:</span> {val.join(", ")}
                            </div>
                        ))}
                    </div>
                  </details>
                  {selectedEntry.request_body && (
                      <div className="bg-muted p-2 rounded-md font-mono text-xs overflow-x-auto whitespace-pre-wrap max-h-[300px] overflow-y-auto">
                        {tryFormatJson(selectedEntry.request_body)}
                      </div>
                  )}
                </div>
              </div>

              <div className="border-t pt-4">
                <h3 className="text-sm font-semibold mb-2 flex items-center gap-2">
                    Response
                    <Badge variant={getStatusColor(selectedEntry.status)} className="ml-2">{selectedEntry.status}</Badge>
                </h3>
                <div className="space-y-2">
                  <details className="text-sm">
                    <summary className="cursor-pointer text-muted-foreground hover:text-foreground">Headers</summary>
                    <div className="mt-2 bg-muted p-2 rounded-md font-mono text-xs overflow-x-auto">
                        {Object.entries(selectedEntry.response_headers).map(([key, val]) => (
                            <div key={key}>
                                <span className="font-semibold">{key}:</span> {val.join(", ")}
                            </div>
                        ))}
                    </div>
                  </details>
                  {selectedEntry.response_body && (
                      <div className="bg-muted p-2 rounded-md font-mono text-xs overflow-x-auto whitespace-pre-wrap max-h-[300px] overflow-y-auto">
                        {tryFormatJson(selectedEntry.response_body)}
                      </div>
                  )}
                </div>
              </div>
            </div>
          )}
        </SheetContent>
      </Sheet>
    </div>
  );
};

function tryFormatJson(str: string) {
  try {
    const parsed = JSON.parse(str);
    return JSON.stringify(parsed, null, 2);
  } catch (e) {
    return str;
  }
}

export default TrafficInspector;
