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
  Activity,
  Clock,
  CheckCircle2,
  AlertTriangle,
  Search
} from "lucide-react"

import { Prism as SyntaxHighlighter } from 'react-syntax-highlighter';
import { vscDarkPlus } from 'react-syntax-highlighter/dist/esm/styles/prism';

import { cn } from "@/lib/utils"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import { ScrollArea } from "@/components/ui/scroll-area"
import { Input } from "@/components/ui/input"
import { ResizableHandle, ResizablePanel, ResizablePanelGroup } from "@/components/ui/resizable"

interface TrafficPayload {
  method: string
  duration: string
  request: any
  result?: any
  error?: string
}

interface LogEntry {
  id: string
  timestamp: string
  level: string
  message: string
  source?: string
  metadata?: {
    payload?: TrafficPayload
  }
}

export function Inspector() {
  const [logs, setLogs] = React.useState<LogEntry[]>([])
  const [selectedId, setSelectedId] = React.useState<string | null>(null)
  const [isPaused, setIsPaused] = React.useState(false)
  const [isConnected, setIsConnected] = React.useState(false)
  const [searchQuery, setSearchQuery] = React.useState("")

  const wsRef = React.useRef<WebSocket | null>(null)
  const isPausedRef = React.useRef(isPaused)
  const scrollRef = React.useRef<HTMLDivElement>(null)

  React.useEffect(() => {
    isPausedRef.current = isPaused
  }, [isPaused])

  React.useEffect(() => {
    const connect = () => {
      const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
      const host = window.location.host
      const wsUrl = `${protocol}//${host}/api/v1/ws/logs`
      const ws = new WebSocket(wsUrl)

      ws.onopen = () => setIsConnected(true)

      ws.onmessage = (event) => {
        if (isPausedRef.current) return

        try {
          const newLog: LogEntry = JSON.parse(event.data)
          // Filter for INSPECTOR source
          if (newLog.source === "INSPECTOR") {
             setLogs((prev) => {
                const newLogs = [...prev, newLog]
                if (newLogs.length > 500) return newLogs.slice(newLogs.length - 500)
                return newLogs
             })

             // Auto-select if none selected
            //  if (!selectedId) setSelectedId(newLog.id)
          }
        } catch (e) {
          console.error("Failed to parse log message", e)
        }
      }

      ws.onclose = () => {
        setIsConnected(false)
        setTimeout(connect, 3000)
      }

      wsRef.current = ws
    }

    connect()
    return () => wsRef.current?.close()
  }, [])

  // Auto-scroll
  React.useEffect(() => {
    if (!isPaused && scrollRef.current) {
        const viewport = scrollRef.current.querySelector('[data-radix-scroll-area-viewport]') as HTMLElement;
        if (viewport) {
             viewport.scrollTop = viewport.scrollHeight;
        }
    }
  }, [logs, isPaused])

  const filteredLogs = React.useMemo(() => {
      if (!searchQuery) return logs;
      const lower = searchQuery.toLowerCase();
      return logs.filter(l =>
          l.metadata?.payload?.method?.toLowerCase().includes(lower) ||
          JSON.stringify(l.metadata?.payload).toLowerCase().includes(lower)
      )
  }, [logs, searchQuery])

  const selectedLog = logs.find(l => l.id === selectedId)

  return (
    <div className="h-[calc(100vh-4rem)] flex flex-col bg-background">
      {/* Header */}
      <div className="flex items-center justify-between p-4 border-b">
        <div className="flex items-center gap-4">
             <div className="flex items-center gap-2">
                <Activity className="w-5 h-5 text-primary" />
                <h1 className="font-semibold text-lg">Network Inspector</h1>
            </div>
            <div className="relative w-64">
                <Search className="absolute left-2 top-2.5 h-4 w-4 text-muted-foreground" />
                <Input
                    placeholder="Filter traffic..."
                    className="pl-8 h-9 bg-muted/50"
                    value={searchQuery}
                    onChange={(e) => setSearchQuery(e.target.value)}
                />
            </div>
        </div>
        <div className="flex items-center gap-2">
            <Badge variant={isConnected ? "outline" : "destructive"} className="font-mono text-xs">
                {isConnected ? "Connected" : "Disconnected"}
            </Badge>
            <Button variant="outline" size="sm" onClick={() => setIsPaused(!isPaused)}>
                {isPaused ? <Play className="mr-2 h-4 w-4" /> : <Pause className="mr-2 h-4 w-4" />}
                {isPaused ? "Resume" : "Pause"}
            </Button>
            <Button variant="outline" size="sm" onClick={() => { setLogs([]); setSelectedId(null); }}>
                <Trash2 className="mr-2 h-4 w-4" /> Clear
            </Button>
        </div>
      </div>

      <ResizablePanelGroup direction="horizontal">
        {/* List Pane */}
        <ResizablePanel defaultSize={40} minSize={30}>
            <ScrollArea className="h-full" ref={scrollRef}>
                <div className="flex flex-col">
                    {filteredLogs.map(log => {
                        const payload = log.metadata?.payload
                        if (!payload) return null
                        const isError = !!payload.error

                        return (
                            <div
                                key={log.id}
                                onClick={() => setSelectedId(log.id)}
                                className={cn(
                                    "flex flex-col p-3 border-b cursor-pointer hover:bg-muted/50 transition-colors text-sm",
                                    selectedId === log.id && "bg-muted border-l-4 border-l-primary pl-2"
                                )}
                            >
                                <div className="flex items-center justify-between mb-1">
                                    <span className="font-mono font-semibold text-primary">{payload.method}</span>
                                    <span className="text-xs text-muted-foreground flex items-center gap-1">
                                        <Clock className="h-3 w-3" />
                                        {payload.duration}
                                    </span>
                                </div>
                                <div className="flex items-center justify-between">
                                     <span className="text-xs text-muted-foreground">
                                        {new Date(log.timestamp).toLocaleTimeString()}
                                     </span>
                                     <Badge variant={isError ? "destructive" : "secondary"} className="text-[10px] px-1 py-0 h-5">
                                        {isError ? "ERROR" : "OK"}
                                     </Badge>
                                </div>
                            </div>
                        )
                    })}
                    {filteredLogs.length === 0 && (
                        <div className="p-8 text-center text-muted-foreground text-sm">
                            No traffic captured. Make sure to generate some MCP requests.
                        </div>
                    )}
                </div>
            </ScrollArea>
        </ResizablePanel>

        <ResizableHandle />

        {/* Detail Pane */}
        <ResizablePanel defaultSize={60}>
            {selectedLog ? (
                <div className="h-full flex flex-col bg-muted/10">
                    <div className="p-4 border-b bg-background">
                         <div className="flex items-center gap-2 mb-2">
                            {selectedLog.metadata?.payload?.error ?
                                <AlertTriangle className="h-5 w-5 text-destructive" /> :
                                <CheckCircle2 className="h-5 w-5 text-green-500" />
                            }
                            <h2 className="font-bold text-lg">{selectedLog.metadata?.payload?.method}</h2>
                         </div>
                         <div className="grid grid-cols-2 gap-4 text-sm">
                            <div className="flex flex-col">
                                <span className="text-muted-foreground text-xs">Timestamp</span>
                                <span className="font-mono">{selectedLog.timestamp}</span>
                            </div>
                             <div className="flex flex-col">
                                <span className="text-muted-foreground text-xs">Duration</span>
                                <span className="font-mono">{selectedLog.metadata?.payload?.duration}</span>
                            </div>
                         </div>
                    </div>
                    <ScrollArea className="flex-1">
                        <div className="p-4 space-y-6">
                            <div>
                                <h3 className="text-sm font-semibold mb-2 text-muted-foreground">Request Parameters</h3>
                                <div className="rounded-md overflow-hidden border">
                                    <SyntaxHighlighter
                                        language="json"
                                        style={vscDarkPlus}
                                        customStyle={{ margin: 0, fontSize: '12px' }}
                                        wrapLongLines={true}
                                    >
                                        {JSON.stringify(selectedLog.metadata?.payload?.request?.Params || selectedLog.metadata?.payload?.request, null, 2)}
                                    </SyntaxHighlighter>
                                </div>
                            </div>

                            <div>
                                <h3 className="text-sm font-semibold mb-2 text-muted-foreground">Response Result</h3>
                                <div className="rounded-md overflow-hidden border">
                                     <SyntaxHighlighter
                                        language="json"
                                        style={vscDarkPlus}
                                        customStyle={{ margin: 0, fontSize: '12px' }}
                                        wrapLongLines={true}
                                    >
                                        {JSON.stringify(
                                            selectedLog.metadata?.payload?.error ?
                                            { error: selectedLog.metadata?.payload?.error } :
                                            selectedLog.metadata?.payload?.result,
                                            null,
                                            2
                                        )}
                                    </SyntaxHighlighter>
                                </div>
                            </div>

                            {/* Full Raw Payload Debug */}
                            <div className="pt-8 opacity-50 hover:opacity-100 transition-opacity">
                                <h4 className="text-xs font-semibold mb-1 uppercase tracking-wider">Raw Log Entry</h4>
                                <pre className="text-[10px] bg-black/50 p-2 rounded overflow-x-auto text-muted-foreground">
                                    {JSON.stringify(selectedLog, null, 2)}
                                </pre>
                            </div>
                        </div>
                    </ScrollArea>
                </div>
            ) : (
                <div className="h-full flex items-center justify-center text-muted-foreground text-sm flex-col gap-2">
                    <Activity className="h-10 w-10 opacity-20" />
                    <p>Select a request to view details</p>
                </div>
            )}
        </ResizablePanel>
      </ResizablePanelGroup>
    </div>
  )
}
