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
  Terminal
} from "lucide-react"

import { cn } from "@/lib/utils"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Badge } from "@/components/ui/badge"
import { ScrollArea } from "@/components/ui/scroll-area"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import { Card, CardContent, CardHeader } from "@/components/ui/card"

export type LogLevel = "INFO" | "WARN" | "ERROR" | "DEBUG"

export interface LogEntry {
  id: string
  timestamp: string
  level: LogLevel
  message: string
  source?: string
  // Optimization: Pre-computed lowercase string for search performance
  searchStr?: string
}

const SAMPLE_MESSAGES = {
  INFO: [
    "Server started on port 8080",
    "Request received: GET /api/v1/tools",
    "Health check passed",
    "Configuration reloaded",
    "User authenticated successfully",
    "Backup completed",
  ],
  WARN: [
    "Response time > 500ms",
    "Rate limit approaching for user",
    "Deprecated API usage detected",
    "Memory usage at 75%",
  ],
  ERROR: [
    "Database connection timeout",
    "Failed to parse JSON body",
    "Upstream service unavailable",
    "Permission denied: /etc/hosts",
  ],
  DEBUG: [
    "Payload size: 1024 bytes",
    "Executing query: SELECT * FROM users",
    "Cache miss for key: user_123",
    "Context switch",
  ],
}

const SOURCES = ["gateway", "auth-service", "db-worker", "api-server", "scheduler"]

const getLevelColor = (level: LogLevel) => {
  switch (level) {
    case "INFO": return "text-blue-400"
    case "WARN": return "text-yellow-400"
    case "ERROR": return "text-red-400"
    case "DEBUG": return "text-gray-400"
    default: return "text-foreground"
  }
}

// Optimization: Memoize LogRow to prevent unnecessary re-renders when list updates
const LogRow = React.memo(({ log }: { log: LogEntry }) => {
  return (
    <div
      className="group flex flex-col sm:flex-row sm:items-start gap-1 sm:gap-3 hover:bg-white/5 p-2 sm:p-1 rounded transition-colors break-words border-b border-white/5 sm:border-0"
      // Optimization: content-visibility allows the browser to skip rendering work for off-screen rows.
      // This significantly improves performance when the log list grows large.
      style={{ contentVisibility: 'auto', containIntrinsicSize: '0 32px' } as React.CSSProperties}
    >
      <div className="flex items-center gap-2 sm:contents">
          <span className="text-muted-foreground whitespace-nowrap opacity-50 text-[10px] sm:text-xs sm:mt-0.5">
            {new Date(log.timestamp).toLocaleTimeString()}
          </span>
          <span className={cn("font-bold w-12 text-[10px] sm:text-xs sm:mt-0.5", getLevelColor(log.level))}>
            {log.level}
          </span>
          {log.source && (
            <span className="text-cyan-600 dark:text-cyan-400 sm:hidden inline-block truncate text-[10px] flex-1 text-right" title={log.source}>
              [{log.source}]
            </span>
          )}
      </div>

      {log.source && (
        <span className="text-cyan-600 dark:text-cyan-400 hidden sm:inline-block w-24 truncate text-xs mt-0.5" title={log.source}>
          [{log.source}]
        </span>
      )}
      <span className="text-gray-300 flex-1 text-xs sm:text-sm pl-0 sm:pl-0">
        {log.message}
      </span>
    </div>
  )
})
LogRow.displayName = 'LogRow'

export function LogStream() {
  const [logs, setLogs] = React.useState<LogEntry[]>([])
  const [isPaused, setIsPaused] = React.useState(false)
  const [filterLevel, setFilterLevel] = React.useState<string>("ALL")
  const [searchQuery, setSearchQuery] = React.useState("")
  // Optimization: Defer the search query to keep the UI responsive while filtering large lists
  const deferredSearchQuery = React.useDeferredValue(searchQuery)
  const scrollRef = React.useRef<HTMLDivElement>(null)

  // Mock log generation
  React.useEffect(() => {
    if (isPaused) return

    const interval = setInterval(() => {
      // Optimization: Stop generating logs when tab is hidden to save resources
      if (document.hidden) return

      const levels: LogLevel[] = ["INFO", "INFO", "INFO", "WARN", "DEBUG", "ERROR"]
      const level = levels[Math.floor(Math.random() * levels.length)]
      const messages = SAMPLE_MESSAGES[level]
      const message = messages[Math.floor(Math.random() * messages.length)]
      const source = SOURCES[Math.floor(Math.random() * SOURCES.length)]

      // Optimization: Pre-compute search string to avoid repetitive toLowerCase() during filtering
      const searchStr = (message + " " + (source || "")).toLowerCase()

      const newLog: LogEntry = {
        id: Math.random().toString(36).substring(7),
        timestamp: new Date().toISOString(),
        level,
        message,
        source,
        searchStr
      }

      setLogs((prev) => {
        // Optimization: Avoid creating an intermediate array (double copy) when limit is reached.
        // If we are at limit, slice the old array (copy N-1 items) and append the new one.
        if (prev.length >= 1000) {
          // Optimization: slice(1) returns a shallow copy. push mutates it.
          // This avoids the double allocation of [...prev.slice(1), newLog].
          const next = prev.slice(1)
          next.push(newLog)
          return next
        }
        return [...prev, newLog]
      })
    }, 800)

    return () => clearInterval(interval)
  }, [isPaused])

  // Auto-scroll
  // Optimization: Cache the viewport element to avoid frequent DOM queries (querySelector).
  const viewportRef = React.useRef<HTMLElement | null>(null)

  React.useEffect(() => {
    if (!isPaused && scrollRef.current) {
      // Lazy init or validate viewport ref
      if (!viewportRef.current || !viewportRef.current.isConnected) {
        viewportRef.current = scrollRef.current.querySelector('[data-radix-scroll-area-viewport]') as HTMLElement
      }

      const scrollContainer = viewportRef.current
      if (scrollContainer) {
        scrollContainer.scrollTop = scrollContainer.scrollHeight
      }
    }
  }, [logs, isPaused])

  // Optimization: Memoize filtered logs and pre-calculate lowercase search query
  // to avoid O(N) redundant string operations during filtering
  const filteredLogs = React.useMemo(() => {
    // Optimization: Fast path for when no filters are active.
    // This avoids iterating over the entire array (up to 1000 items) every 800ms when showing all logs.
    if (filterLevel === "ALL" && !deferredSearchQuery) {
      return logs
    }

    const lowerSearchQuery = deferredSearchQuery.toLowerCase()
    return logs.filter((log) => {
      const matchesLevel = filterLevel === "ALL" || log.level === filterLevel

      // Optimization: Use pre-computed search string if available to skip repeated toLowerCase() calls
      let matchesSearch: boolean | undefined
      if (log.searchStr) {
        matchesSearch = log.searchStr.includes(lowerSearchQuery)
      } else {
        matchesSearch =
          log.message.toLowerCase().includes(lowerSearchQuery) ||
          log.source?.toLowerCase().includes(lowerSearchQuery)
      }

      return matchesLevel && matchesSearch
    })
  }, [logs, filterLevel, deferredSearchQuery])

  const clearLogs = () => setLogs([])

  const downloadLogs = () => {
    const content = logs.map(l => `[${l.timestamp}] [${l.level}] [${l.source}] ${l.message}`).join('\n')
    const blob = new Blob([content], { type: 'text/plain' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `logs-${new Date().toISOString()}.txt`
    a.click()
  }

  return (
    <div className="flex flex-col h-full gap-4">
      <div className="flex flex-col gap-4 md:flex-row md:items-center justify-between">
        <div className="flex items-center justify-between md:justify-start gap-2">
            <div className="flex items-center gap-2">
                <h1 className="text-2xl font-bold tracking-tight flex items-center gap-2">
                    <Terminal className="w-6 h-6" /> Live Logs
                </h1>
                <Badge variant="outline" className="font-mono text-xs">
                    {logs.length} events
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

      <Card className="flex-1 flex flex-col overflow-hidden border-muted/50 shadow-sm bg-background/50 backdrop-blur-sm">
        <CardHeader className="p-4 border-b bg-muted/20">
             <div className="flex flex-col md:flex-row gap-4 justify-between">
                <div className="relative flex-1 max-w-sm w-full">
                    <Search className="absolute left-2 top-2.5 h-4 w-4 text-muted-foreground" />
                    <Input
                    placeholder="Search logs..."
                    className="pl-8 bg-background w-full"
                    value={searchQuery}
                    onChange={(e) => setSearchQuery(e.target.value)}
                    />
                </div>
                <div className="flex items-center gap-2 justify-end">
                    <Filter className="h-4 w-4 text-muted-foreground" />
                    <Select value={filterLevel} onValueChange={setFilterLevel}>
                        <SelectTrigger className="w-[120px] bg-background">
                            <SelectValue placeholder="Level" />
                        </SelectTrigger>
                        <SelectContent>
                            <SelectItem value="ALL">All Levels</SelectItem>
                            <SelectItem value="INFO">Info</SelectItem>
                            <SelectItem value="WARN">Warning</SelectItem>
                            <SelectItem value="ERROR">Error</SelectItem>
                            <SelectItem value="DEBUG">Debug</SelectItem>
                        </SelectContent>
                    </Select>
                </div>
             </div>
        </CardHeader>
        <CardContent className="flex-1 p-0 overflow-hidden bg-black/90 font-mono text-sm relative">
             <ScrollArea className="h-full w-full p-4" ref={scrollRef}>
                <div className="space-y-1" data-testid="log-rows-container">
                    {filteredLogs.length === 0 && (
                        <div className="text-muted-foreground text-center py-10 italic">
                            No logs found matching your criteria...
                        </div>
                    )}
                    {filteredLogs.map((log) => (
                      <LogRow key={log.id} log={log} />
                    ))}
                    {/* Invisible element to help auto-scroll if needed, though scroll logic handles viewport */}
                </div>
             </ScrollArea>
        </CardContent>
      </Card>
    </div>
  )
}
