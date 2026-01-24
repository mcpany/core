/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client"

import * as React from "react"
import {
  Activity,
  ArrowDown,
  ArrowUp,
  Ban,
  CheckCircle2,
  Clock,
  Download,
  Pause,
  Play,
  Search,
  Trash2,
  XCircle,
  Zap
} from "lucide-react"

import { Prism as SyntaxHighlighter } from 'react-syntax-highlighter';
import { vscDarkPlus } from 'react-syntax-highlighter/dist/esm/styles/prism';

import { cn } from "@/lib/utils"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Input } from "@/components/ui/input"
import { ResizableHandle, ResizablePanel, ResizablePanelGroup } from "@/components/ui/resizable"
import { ScrollArea } from "@/components/ui/scroll-area"

interface TrafficEvent {
  id: string
  timestamp: string
  method: string
  duration: string
  status: "success" | "error"
  request: any
  result?: any
  error?: string
  payload_json: string // raw string for search
}

interface LogEntry {
  id: string
  timestamp: string
  level: string
  message: string
  source?: string
  metadata?: Record<string, unknown>
}

export function Inspector() {
  const [events, setEvents] = React.useState<TrafficEvent[]>([])
  const [selectedEventId, setSelectedEventId] = React.useState<string | null>(null)
  const [isPaused, setIsPaused] = React.useState(false)
  const [searchQuery, setSearchQuery] = React.useState("")
  const [isConnected, setIsConnected] = React.useState(false)

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

      ws.onopen = () => {
        setIsConnected(true)
      }

      ws.onmessage = (event) => {
        if (isPausedRef.current) return
        if (document.hidden) return

        try {
          const log: LogEntry = JSON.parse(event.data)

          // Filter for Inspector traffic
          if (log.source !== "INSPECTOR") return

          // Parse the message which contains the JSON payload
          let payload: any
          try {
            payload = JSON.parse(log.message)
          } catch (e) {
            console.error("Failed to parse traffic payload", e)
            return
          }

          const newEvent: TrafficEvent = {
            id: log.id, // Use log ID or generate new one? Log ID is unique enough.
            timestamp: payload.timestamp || log.timestamp,
            method: payload.method || "unknown",
            duration: payload.duration || "0s",
            status: payload.status || "success",
            request: payload.request,
            result: payload.result,
            error: payload.error,
            payload_json: log.message.toLowerCase()
          }

          setEvents((prev) => {
            const MAX_EVENTS = 500
            const newEvents = [newEvent, ...prev]
            if (newEvents.length > MAX_EVENTS) {
              return newEvents.slice(0, MAX_EVENTS)
            }
            return newEvents
          })
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

    return () => {
      wsRef.current?.close()
    }
  }, [])

  const filteredEvents = React.useMemo(() => {
    if (!searchQuery) return events
    const lowerQuery = searchQuery.toLowerCase()
    return events.filter(e =>
      e.method.toLowerCase().includes(lowerQuery) ||
      e.payload_json.includes(lowerQuery)
    )
  }, [events, searchQuery])

  const selectedEvent = React.useMemo(() =>
    events.find(e => e.id === selectedEventId) || filteredEvents[0] || null
  , [events, selectedEventId, filteredEvents])

  const clearEvents = () => setEvents([])

  return (
    <div className="flex flex-col h-full bg-background text-foreground">
      {/* Header */}
      <div className="flex items-center justify-between p-4 border-b h-16 shrink-0">
        <div className="flex items-center gap-4">
          <div className="flex items-center gap-2">
            <Activity className="w-5 h-5 text-primary" />
            <h1 className="text-lg font-semibold">Inspector</h1>
          </div>
          <Badge variant={isConnected ? "outline" : "destructive"} className="gap-1.5 font-mono text-xs">
             <span className={cn("relative flex h-2 w-2", isConnected ? "bg-green-500" : "bg-red-500", "rounded-full")}>
               {isConnected && <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-green-400 opacity-75"></span>}
             </span>
             {isConnected ? "Live" : "Disconnected"}
          </Badge>
          <div className="h-6 w-px bg-border mx-2" />
          <div className="relative w-64">
            <Search className="absolute left-2 top-2.5 h-4 w-4 text-muted-foreground" />
            <Input
              placeholder="Filter traffic..."
              className="pl-8 h-9"
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
            />
          </div>
        </div>
        <div className="flex items-center gap-2">
          <Button
            variant="ghost"
            size="sm"
            onClick={() => setIsPaused(!isPaused)}
            className={cn(isPaused && "text-amber-500")}
          >
            {isPaused ? <Play className="mr-2 h-4 w-4" /> : <Pause className="mr-2 h-4 w-4" />}
            {isPaused ? "Resume" : "Pause"}
          </Button>
          <Button variant="ghost" size="sm" onClick={clearEvents}>
            <Trash2 className="mr-2 h-4 w-4" /> Clear
          </Button>
        </div>
      </div>

      {/* Content */}
      <ResizablePanelGroup direction="horizontal" className="flex-1">

        {/* List Pane */}
        <ResizablePanel defaultSize={40} minSize={30} maxSize={60} className="flex flex-col border-r bg-muted/10">
          <ScrollArea className="flex-1">
            <div className="flex flex-col">
              {filteredEvents.length === 0 && (
                 <div className="p-8 text-center text-muted-foreground">
                    <Zap className="h-10 w-10 mx-auto mb-2 opacity-50" />
                    <p>No traffic detected</p>
                    <p className="text-xs opacity-70">Waiting for MCP requests...</p>
                 </div>
              )}
              {filteredEvents.map((event) => (
                <button
                  key={event.id}
                  onClick={() => setSelectedEventId(event.id)}
                  className={cn(
                    "flex flex-col gap-1 p-3 text-left border-b transition-colors hover:bg-muted/50",
                    (selectedEventId === event.id || (!selectedEventId && filteredEvents[0]?.id === event.id)) && "bg-muted border-l-2 border-l-primary"
                  )}
                >
                  <div className="flex items-center justify-between w-full">
                    <span className="font-mono text-sm font-semibold text-primary">{event.method}</span>
                    <Badge
                      variant="outline"
                      className={cn(
                        "text-[10px] h-5 px-1.5",
                        event.status === "error" ? "border-red-500 text-red-500" : "border-green-500 text-green-500"
                      )}
                    >
                      {event.status === "error" ? "ERR" : "OK"}
                    </Badge>
                  </div>
                  <div className="flex items-center justify-between text-xs text-muted-foreground">
                    <span className="flex items-center gap-1">
                      <Clock className="h-3 w-3" />
                      {new Date(event.timestamp).toLocaleTimeString()}
                    </span>
                    <span>{event.duration}</span>
                  </div>
                </button>
              ))}
            </div>
          </ScrollArea>
        </ResizablePanel>

        <ResizableHandle />

        {/* Detail Pane */}
        <ResizablePanel defaultSize={60}>
          {selectedEvent ? (
             <div className="flex flex-col h-full overflow-hidden">
                <div className="p-4 border-b bg-muted/5 flex items-center justify-between">
                    <div>
                        <h2 className="text-lg font-bold font-mono text-primary flex items-center gap-2">
                            {selectedEvent.method}
                            {selectedEvent.status === "error" ? (
                                <XCircle className="h-5 w-5 text-red-500" />
                            ) : (
                                <CheckCircle2 className="h-5 w-5 text-green-500" />
                            )}
                        </h2>
                        <div className="flex items-center gap-4 text-xs text-muted-foreground mt-1 font-mono">
                            <span>ID: {selectedEvent.id.slice(0, 8)}</span>
                            <span>Time: {new Date(selectedEvent.timestamp).toLocaleString()}</span>
                            <span>Dur: {selectedEvent.duration}</span>
                        </div>
                    </div>
                    {/* Actions could go here */}
                </div>

                <div className="flex-1 overflow-auto p-0">
                    <div className="grid grid-rows-2 h-full">
                         {/* Request */}
                         <div className="row-span-1 border-b flex flex-col min-h-[200px]">
                            <div className="px-4 py-2 bg-muted/10 border-b text-xs font-semibold text-muted-foreground flex items-center gap-2">
                                <ArrowUp className="h-3 w-3" /> Request Params
                            </div>
                            <div className="flex-1 relative overflow-hidden bg-[#1e1e1e]">
                                <ScrollArea className="h-full w-full">
                                    <SyntaxHighlighter
                                        language="json"
                                        style={vscDarkPlus}
                                        customStyle={{ margin: 0, padding: '1rem', background: 'transparent' }}
                                        wrapLongLines={true}
                                    >
                                        {JSON.stringify(selectedEvent.request, null, 2)}
                                    </SyntaxHighlighter>
                                </ScrollArea>
                            </div>
                         </div>

                         {/* Response */}
                         <div className="row-span-1 flex flex-col min-h-[200px]">
                            <div className="px-4 py-2 bg-muted/10 border-b text-xs font-semibold text-muted-foreground flex items-center gap-2">
                                <ArrowDown className="h-3 w-3" />
                                {selectedEvent.error ? "Error" : "Response Result"}
                            </div>
                            <div className="flex-1 relative overflow-hidden bg-[#1e1e1e]">
                                <ScrollArea className="h-full w-full">
                                    <SyntaxHighlighter
                                        language="json"
                                        style={vscDarkPlus}
                                        customStyle={{ margin: 0, padding: '1rem', background: 'transparent' }}
                                        wrapLongLines={true}
                                    >
                                        {JSON.stringify(selectedEvent.error ? { error: selectedEvent.error } : selectedEvent.result, null, 2)}
                                    </SyntaxHighlighter>
                                </ScrollArea>
                            </div>
                         </div>
                    </div>
                </div>
             </div>
          ) : (
            <div className="flex items-center justify-center h-full text-muted-foreground">
                <div className="text-center">
                    <Activity className="h-12 w-12 mx-auto mb-4 opacity-20" />
                    <p>Select an event to view details</p>
                </div>
            </div>
          )}
        </ResizablePanel>

      </ResizablePanelGroup>
    </div>
  )
}
