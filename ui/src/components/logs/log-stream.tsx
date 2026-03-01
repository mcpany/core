/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client"

import * as React from "react"
import {
  Pause,
  Play,
  Trash2,
  Search,
  Download,
  Filter,
  Terminal,
  Unplug,
  Monitor,
} from "lucide-react"

import { useSearchParams } from "next/navigation"

import { cn } from "@/lib/utils"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Badge } from "@/components/ui/badge"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import {
  DropdownMenu,
  DropdownMenuTrigger,
  DropdownMenuContent,
  DropdownMenuCheckboxItem,
} from "@/components/ui/dropdown-menu"
import { Card, CardContent, CardHeader } from "@/components/ui/card"
import { LogViewer, LogEntry, timeFormatter } from "./log-viewer"

/**
 * LogStream component.
 * @param props - The component props.
 * @param props.source - Optional source to filter by initially.
 * @param props.traceId - Optional trace ID to filter logs by (requires support in log metadata).
 * @param props.traceStartTime - Optional start time of the trace for time-window filtering.
 * @param props.traceEndTime - Optional end time of the trace for time-window filtering.
 * @returns The rendered component.
 */
export function LogStream({
  source,
  traceId,
  traceStartTime,
  traceEndTime
}: {
  source?: string;
  traceId?: string;
  traceStartTime?: number;
  traceEndTime?: number;
}) {
  const [logs, setLogs] = React.useState<LogEntry[]>([])
  const [isPaused, setIsPaused] = React.useState(false)
  // Optimization: Use a ref to access the latest isPaused state inside the WebSocket closure
  // without triggering a reconnection or having a stale closure.
  const isPausedRef = React.useRef(isPaused)

  React.useEffect(() => {
    isPausedRef.current = isPaused
  }, [isPaused])

  const searchParams = useSearchParams()
  // Use prop source if provided, otherwise fallback to URL param or "ALL"
  const initialSource = source || searchParams.get("source") || "ALL"

  const initialLevel = searchParams.get("level") || "ALL"
  const [filterLevels, setFilterLevels] = React.useState<string[]>(initialLevel === "ALL" ? ["INFO", "WARN", "ERROR", "DEBUG"] : initialLevel.split(","))
  const [filterSource, setFilterSource] = React.useState<string>(initialSource)
  const [searchQuery, setSearchQuery] = React.useState("")
  const [isConnected, setIsConnected] = React.useState(false)
  // Optimization: Defer the search query to keep the UI responsive while filtering large lists
  const deferredSearchQuery = React.useDeferredValue(searchQuery)

  // Optimization: Pre-compile regex for highlighting to avoid repeated RegExp creation in render loop
  const highlightRegex = React.useMemo(() => {
    if (!deferredSearchQuery) return null;
    const escaped = deferredSearchQuery.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
    return new RegExp(`(${escaped})`, 'gi');
  }, [deferredSearchQuery]);

  const wsRef = React.useRef<WebSocket | null>(null)
  // Optimization: Buffer for incoming logs to support batch processing
  const logBufferRef = React.useRef<LogEntry[]>([])

  React.useEffect(() => {
    // Optimization: Flush buffer periodically to limit re-renders
    const flushInterval = setInterval(() => {
      if (logBufferRef.current.length > 0) {
        setLogs((prev) => {
          const buffer = logBufferRef.current
          logBufferRef.current = [] // Clear buffer
          const MAX_LOGS = 2000 // Increased limit to allow for more history

          // Optimization: Efficient array handling to minimize memory allocation and gc pressure.
          // Avoiding large intermediate arrays reduces garbage collection overhead during rapid logging.

          // Case 1: Total logs fit within limit - simple concat
          if (prev.length + buffer.length <= MAX_LOGS) {
            return [...prev, ...buffer]
          }

          // Case 2: Buffer itself exceeds limit (unlikely but possible) - take last MAX_LOGS
          if (buffer.length >= MAX_LOGS) {
            return buffer.slice(buffer.length - MAX_LOGS)
          }

          // Case 3: Need to trim from prev to make room for buffer
          // We need (MAX_LOGS - buffer.length) from the end of prev
          const keepCount = MAX_LOGS - buffer.length
          return [...prev.slice(-keepCount), ...buffer]
        })
      }
    }, 100) // Flush every 100ms

    const connect = () => {
      // Determine protocol (ws or wss)
      // Check if we are running in a browser environment
      if (typeof window === 'undefined') return;

      const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
      const host = window.location.host

      let wsUrl = `${protocol}//${host}/api/v1/ws/logs`

      // Append auth token if available to authenticate the WebSocket connection
      if (typeof window !== 'undefined') {
        const token = localStorage.getItem('mcp_auth_token')
        if (token) {
            const separator = wsUrl.includes('?') ? '&' : '?';
            wsUrl += `${separator}auth_token=${encodeURIComponent(token)}`;
        }
      }

      const ws = new WebSocket(wsUrl)

      ws.onopen = () => {
        setIsConnected(true)
      }

      ws.onmessage = (event) => {
        // Optimization: check against the ref to ensure we respect the current pause state
        if (isPausedRef.current) return
        if (document.hidden) return

        try {
          const newLog: LogEntry = JSON.parse(event.data)
          // Optimization: Pre-compute formatted time to avoid expensive Date parsing during render.
          // Uses cached timeFormatter for better performance.
          newLog.formattedTime = timeFormatter
            ? timeFormatter.format(new Date(newLog.timestamp))
            : new Date(newLog.timestamp).toLocaleTimeString()

          // Optimization: Add to buffer instead of calling setLogs directly
          logBufferRef.current.push(newLog)
        } catch (e) {
          console.error("Failed to parse log message", e)
        }
      }

      ws.onclose = () => {
        setIsConnected(false)
        setTimeout(connect, 3000)
      }

      ws.onerror = (err) => {
        console.error("WebSocket error", err)
        ws.close()
      }

      wsRef.current = ws
    }

    connect()

    return () => {
      wsRef.current?.close()
      clearInterval(flushInterval)
    }
  }, []) // Empty dependency array -> run once

  // Optimization: Extract unique sources from logs efficiently
  const uniqueSources = React.useMemo(() => {
    const sources = new Set<string>()
    logs.forEach(log => {
      if (log.source) {
        sources.add(log.source)
      }
    })
    return Array.from(sources).sort()
  }, [logs])

  // Optimization: Memoize filtered logs and pre-calculate lowercase search query
  // to avoid O(N) redundant string operations during filtering
  const filteredLogs = React.useMemo(() => {
    const lowerSearchQuery = deferredSearchQuery.toLowerCase()

    return logs.filter((log) => {
      // 1. Trace ID Filtering (Highest Priority)
      if (traceId) {
        // Check metadata for trace_id
        const logTraceId = log.metadata?.trace_id as string | undefined;
        if (logTraceId && logTraceId === traceId) {
            return true;
        }
        // Fallback: Check if message contains trace ID
        if (log.message.includes(traceId)) {
            return true;
        }

        // 2. Time Window Filtering (Heuristic fallback if explicit traceId match fails but we are in trace mode)
        if (traceStartTime && traceEndTime) {
             const logTime = new Date(log.timestamp).getTime();
             // Allow 500ms buffer around trace
             if (logTime >= traceStartTime - 500 && logTime <= traceEndTime + 500) {
                 return true;
             }
        }

        // If we are in "Trace Mode" (traceId passed) and neither ID match nor time match, exclude.
        // Unless user wants to see ALL logs? Usually correlation implies filtering.
        return false;
      }

      // Normal Filtering
      const matchesLevel = filterLevels.includes(log.level)
      const matchesSource = filterSource === "ALL" || log.source === filterSource

      // ⚡ BOLT: Optimized memory usage by removing eager search string allocation.
      // Randomized Selection from Top 5 High-Impact Targets
      // We calculate matches on demand to avoid O(N) memory overhead for search strings.
      const matchesSearch = !deferredSearchQuery ||
        log.message.toLowerCase().includes(lowerSearchQuery) ||
        log.source?.toLowerCase().includes(lowerSearchQuery)

      return matchesLevel && matchesSource && matchesSearch
    })
  }, [logs, filterLevels, filterSource, deferredSearchQuery, traceId, traceStartTime, traceEndTime])

  const clearLogs = () => setLogs([])

  const downloadLogs = () => {
    const content = filteredLogs.map(l => `[${l.timestamp}] [${l.level}] [${l.source}] ${l.message}`).join('\n')
    const blob = new Blob([content], { type: 'text/plain' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `logs-${new Date().toISOString()}.txt`
    a.click()
  }

  // If running in trace correlation mode, show a simpler header or no header?
  // We should still allow searching within the correlated logs.
  const isEmbedded = !!traceId;

  return (
    <div className="flex flex-col h-full gap-4">
      {!isEmbedded && (
        <div className="flex flex-col gap-4 md:flex-row md:items-center justify-between">
            <div className="flex items-center justify-between md:justify-start gap-2">
                <div className="flex items-center gap-2">
                    <h1 className="text-2xl font-bold tracking-tight flex items-center gap-2">
                        <Terminal className="w-6 h-6" /> Live Logs
                    </h1>
                    <Badge variant={isConnected ? "outline" : "destructive"} className="font-mono text-xs gap-1">
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
                    <Badge variant="secondary" className="font-mono text-xs">
                        {filteredLogs.length} events
                    </Badge>
                </div>
                {/* Mobile-only pause/resume for better access */}
                <div className="md:hidden">
                <Button
                    variant="ghost"
                    size="icon"
                    onClick={() => setIsPaused(!isPaused)}
                >
                    {isPaused ? <Play className="h-4 w-4" /> : <Pause className="h-4 w-4" />}
                </Button>
                </div>
            </div>

            <div className="flex items-center gap-2 justify-end">
            <Button
                variant="outline"
                size="sm"
                onClick={() => setIsPaused(!isPaused)}
                className="w-24 hidden md:flex"
            >
                {isPaused ? <><Play className="mr-2 h-4 w-4" /> Resume</> : <><Pause className="mr-2 h-4 w-4" /> Pause</>}
            </Button>
            <Button variant="outline" size="sm" onClick={clearLogs} className="flex-1 md:flex-none">
                <Trash2 className="mr-2 h-4 w-4" /> Clear
            </Button>
            <Button variant="outline" size="sm" onClick={downloadLogs} className="flex-1 md:flex-none">
                <Download className="mr-2 h-4 w-4" /> Export
            </Button>
            </div>
        </div>
      )}

      <Card className={cn("flex-1 flex flex-col overflow-hidden border-muted/50 shadow-sm bg-background/50 backdrop-blur-sm", isEmbedded && "border-0 shadow-none bg-transparent")}>
        <CardHeader className={cn("p-4 border-b bg-muted/20", isEmbedded && "p-2 bg-transparent border-b")}>
             <div className="flex flex-col md:flex-row gap-4 justify-between">
                <div className="relative flex-1 max-w-sm w-full">
                    <Search className="absolute left-2 top-2.5 h-4 w-4 text-muted-foreground" />
                    <Input
                    placeholder={isEmbedded ? "Search within correlated logs..." : "Search logs..."}
                    className="pl-8 bg-background w-full"
                    value={searchQuery}
                    onChange={(e) => setSearchQuery(e.target.value)}
                    />
                </div>
                {isEmbedded && (
                    <div className="text-[10px] text-muted-foreground flex items-center justify-end px-2 italic">
                        Note: Showing live logs only. Historical logs may not be available.
                    </div>
                )}
                {!isEmbedded && (
                    <div className="flex items-center gap-2 justify-end">
                        <Monitor className="h-4 w-4 text-muted-foreground" />
                        <Select value={filterSource} onValueChange={setFilterSource}>
                            <SelectTrigger className="w-[140px] bg-background">
                                <SelectValue placeholder="Source" />
                            </SelectTrigger>
                            <SelectContent>
                                <SelectItem value="ALL">All Sources</SelectItem>
                                {uniqueSources.map(source => (
                                    <SelectItem key={source} value={source}>{source}</SelectItem>
                                ))}
                            </SelectContent>
                        </Select>

                        <Filter className="h-4 w-4 text-muted-foreground ml-2" />
                        <DropdownMenu>
                            <DropdownMenuTrigger asChild>
                                <Button variant="outline" className="w-[120px] justify-start bg-background font-normal overflow-hidden truncate px-3 py-2 text-sm">
                                    {filterLevels.length === 4 ? "All Levels" : filterLevels.join(", ") || "None"}
                                </Button>
                            </DropdownMenuTrigger>
                            <DropdownMenuContent className="w-[120px]">
                                {["INFO", "WARN", "ERROR", "DEBUG"].map(level => (
                                    <DropdownMenuCheckboxItem
                                        key={level}
                                        checked={filterLevels.includes(level)}
                                        onCheckedChange={(checked) => {
                                            setFilterLevels(prev => {
                                                if (checked) {
                                                    return [...prev, level];
                                                }
                                                return prev.filter(l => l !== level);
                                            });
                                        }}
                                    >
                                        {level === "INFO" ? "Info" :
                                         level === "WARN" ? "Warning" :
                                         level === "ERROR" ? "Error" : "Debug"}
                                    </DropdownMenuCheckboxItem>
                                ))}
                            </DropdownMenuContent>
                        </DropdownMenu>
                    </div>
                )}
             </div>
        </CardHeader>
        <CardContent className="flex-1 p-0 overflow-hidden bg-black/90 font-mono text-sm relative">
             <LogViewer logs={filteredLogs} highlightRegex={highlightRegex} isPaused={isPaused} />
             {filteredLogs.length === 0 && (
                <div className="absolute inset-0 flex items-center justify-center pointer-events-none">
                    <div className="text-muted-foreground text-center italic">
                        {isEmbedded
                            ? "No correlated logs found."
                            : (isConnected ? "No logs found matching your criteria..." : "Waiting for connection...")}
                    </div>
                </div>
             )}
        </CardContent>
      </Card>
    </div>
  )
}
