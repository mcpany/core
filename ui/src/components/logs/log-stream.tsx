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
  Activity,
  Wifi,
  WifiOff
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
import { Separator } from "@/components/ui/separator"
import { useToast } from "@/hooks/use-toast"

export type LogLevel = "INFO" | "WARN" | "ERROR" | "DEBUG"

export interface LogEntry {
  id: string
  timestamp: string
  level: LogLevel
  message: string
  source?: string
}

export function LogStream() {
  const [logs, setLogs] = React.useState<LogEntry[]>([])
  const [isPaused, setIsPaused] = React.useState(false)
  const [isConnected, setIsConnected] = React.useState(false)
  const [filterLevel, setFilterLevel] = React.useState<string>("ALL")
  const [searchQuery, setSearchQuery] = React.useState("")
  const scrollRef = React.useRef<HTMLDivElement>(null)
  const eventSourceRef = React.useRef<EventSource | null>(null)
  const isPausedRef = React.useRef(isPaused)
  const { toast } = useToast()

  // Update ref when state changes
  React.useEffect(() => {
    isPausedRef.current = isPaused
  }, [isPaused])

  // Initialize SSE connection
  React.useEffect(() => {
    connectToStream()
    return () => {
      if (eventSourceRef.current) {
        eventSourceRef.current.close()
      }
    }
  }, [])

  const connectToStream = () => {
    if (eventSourceRef.current) {
      eventSourceRef.current.close()
    }

    const eventSource = new EventSource("/api/logs")
    eventSourceRef.current = eventSource

    eventSource.onopen = () => {
      setIsConnected(true)
    }

    eventSource.onmessage = (event) => {
      if (isPausedRef.current) return
      try {
        const newLog: LogEntry = JSON.parse(event.data)
        setLogs((prev) => {
          const newLogs = [...prev, newLog]
          // Keep last 2000 logs to avoid memory issues
          if (newLogs.length > 2000) return newLogs.slice(newLogs.length - 2000)
          return newLogs
        })
      } catch (e) {
        console.error("Failed to parse log entry", e)
      }
    }

    eventSource.onerror = () => {
      setIsConnected(false)
      eventSource.close()
      // Retry connection after 5 seconds
      setTimeout(connectToStream, 5000)
    }
  }

  // Toggle Pause/Resume
  const togglePause = () => {
    setIsPaused(!isPaused)
  }

  // Auto-scroll
  React.useEffect(() => {
    if (!isPaused && scrollRef.current) {
      const scrollContainer = scrollRef.current.querySelector('[data-radix-scroll-area-viewport]')
      if (scrollContainer) {
        scrollContainer.scrollTop = scrollContainer.scrollHeight
      }
    }
  }, [logs, isPaused])

  const filteredLogs = React.useMemo(() => {
     return logs.filter((log) => {
        const matchesLevel = filterLevel === "ALL" || log.level === filterLevel
        const matchesSearch = log.message.toLowerCase().includes(searchQuery.toLowerCase()) ||
                             (log.source?.toLowerCase().includes(searchQuery.toLowerCase()) ?? false)
        return matchesLevel && matchesSearch
      })
  }, [logs, filterLevel, searchQuery])

  const clearLogs = () => {
      setLogs([])
      toast({
          title: "Logs Cleared",
          description: "All captured logs have been removed from the view.",
      })
  }

  const downloadLogs = () => {
    const content = filteredLogs.map(l => `[${l.timestamp}] [${l.level}] [${l.source}] ${l.message}`).join('\n')
    const blob = new Blob([content], { type: 'text/plain' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `mcp-logs-${new Date().toISOString()}.txt`
    a.click()
    toast({
        title: "Export Complete",
        description: "Logs have been downloaded to your device.",
    })
  }

  const getLevelColor = (level: LogLevel) => {
    switch (level) {
      case "INFO": return "text-blue-500 dark:text-blue-400"
      case "WARN": return "text-yellow-600 dark:text-yellow-400"
      case "ERROR": return "text-red-600 dark:text-red-400"
      case "DEBUG": return "text-gray-500 dark:text-gray-400"
      default: return "text-foreground"
    }
  }

  const getLevelBg = (level: LogLevel) => {
      switch (level) {
          case "INFO": return "bg-blue-500/10"
          case "WARN": return "bg-yellow-500/10"
          case "ERROR": return "bg-red-500/10"
          case "DEBUG": return "bg-gray-500/10"
          default: return ""
      }
  }

  return (
    <div className="flex flex-col h-full gap-4">
      {/* Header Bar */}
      <div className="flex flex-col gap-4 md:flex-row md:items-center justify-between">
        <div className="flex items-center gap-3">
            <div className="bg-primary/10 p-2 rounded-lg">
                <Terminal className="w-5 h-5 text-primary" />
            </div>
            <div>
                <h1 className="text-xl font-bold tracking-tight flex items-center gap-2">
                    Live Stream
                </h1>
                <div className="flex items-center gap-2 text-xs text-muted-foreground">
                    <span className={cn("flex items-center gap-1", isConnected ? "text-green-500" : "text-red-500")}>
                        {isConnected ? <Wifi className="h-3 w-3" /> : <WifiOff className="h-3 w-3" />}
                        {isConnected ? "Connected" : "Disconnected"}
                    </span>
                    <span>•</span>
                    <span>{logs.length} events captured</span>
                </div>
            </div>
        </div>

        <div className="flex items-center gap-2">
           <Button
            variant={isPaused ? "secondary" : "outline"}
            size="sm"
            onClick={togglePause}
            className={cn("w-24 transition-colors", isPaused && "bg-yellow-500/10 text-yellow-600 hover:bg-yellow-500/20")}
          >
            {isPaused ? <><Play className="mr-2 h-3 w-3" /> Resume</> : <><Pause className="mr-2 h-3 w-3" /> Pause</>}
          </Button>
          <Separator orientation="vertical" className="h-6 mx-1" />
          <Button variant="ghost" size="sm" onClick={clearLogs} className="text-muted-foreground hover:text-destructive">
            <Trash2 className="mr-2 h-3 w-3" /> Clear
          </Button>
          <Button variant="outline" size="sm" onClick={downloadLogs}>
            <Download className="mr-2 h-3 w-3" /> Export
          </Button>
        </div>
      </div>

      {/* Main Console Area */}
      <Card className="flex-1 flex flex-col overflow-hidden border-border/40 shadow-sm bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60">
        <CardHeader className="p-3 border-b bg-muted/40 flex flex-row items-center gap-4 space-y-0">
             <div className="relative flex-1 max-w-md">
                <Search className="absolute left-2.5 top-2.5 h-3.5 w-3.5 text-muted-foreground" />
                <Input
                    placeholder="Filter logs..."
                    className="pl-8 h-8 bg-background/50 border-muted-foreground/20 text-xs focus-visible:ring-1"
                    value={searchQuery}
                    onChange={(e) => setSearchQuery(e.target.value)}
                />
            </div>
            <div className="flex items-center gap-2">
                <Filter className="h-3.5 w-3.5 text-muted-foreground" />
                <Select value={filterLevel} onValueChange={setFilterLevel}>
                    <SelectTrigger className="w-[110px] h-8 text-xs bg-background/50 border-muted-foreground/20">
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
        </CardHeader>
        <CardContent className="flex-1 p-0 overflow-hidden bg-[#0c0c0c] dark:bg-[#0c0c0c] relative">
             <ScrollArea className="h-full w-full" ref={scrollRef}>
                <div className="p-4 font-mono text-[13px] leading-6 space-y-0.5 min-h-full" data-testid="log-rows-container">
                    {filteredLogs.length === 0 && (
                        <div className="h-full flex flex-col items-center justify-center text-muted-foreground/50 py-20 gap-2">
                            <Activity className="h-10 w-10 opacity-20" />
                            <p className="text-sm">Waiting for logs...</p>
                        </div>
                    )}
                    {filteredLogs.map((log) => (
                        <div key={log.id} className="group flex items-start gap-3 px-2 rounded-sm hover:bg-white/5 transition-colors">
                            <span className="text-gray-500 select-none w-[85px] shrink-0 text-[11px] pt-0.5 opacity-60">
                                {new Date(log.timestamp).toLocaleTimeString([], { hour12: false, hour: '2-digit', minute:'2-digit', second:'2-digit', fractionalSecondDigits: 3 })}
                            </span>
                            <span className={cn(
                                "text-[10px] font-bold px-1.5 rounded-sm w-[50px] text-center shrink-0 pt-[1px] h-5 inline-flex items-center justify-center",
                                getLevelColor(log.level),
                                getLevelBg(log.level)
                            )}>
                                {log.level}
                            </span>
                            <div className="flex-1 break-all flex gap-3 text-gray-300">
                                {log.source && (
                                    <span className="text-cyan-600 dark:text-cyan-500 shrink-0 opacity-80" title={log.source}>
                                        {log.source}
                                    </span>
                                )}
                                <span className={cn(log.level === "ERROR" && "text-red-400")}>
                                    {log.message}
                                </span>
                            </div>
                        </div>
                    ))}
                </div>
             </ScrollArea>
        </CardContent>
      </Card>
    </div>
  )
}
