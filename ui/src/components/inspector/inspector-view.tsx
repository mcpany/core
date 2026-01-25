/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import React, { useEffect, useRef, useState, useMemo } from 'react';
import {
    Activity,
    ArrowDown,
    ArrowUp,
    CheckCircle2,
    Clock,
    Pause,
    Play,
    Search,
    Trash2,
    XCircle
} from "lucide-react";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { ResizableHandle, ResizablePanel, ResizablePanelGroup } from "@/components/ui/resizable";
import { ScrollArea } from "@/components/ui/scroll-area";
import { cn } from "@/lib/utils";
import { JsonView } from "@/components/ui/json-view";

// Types
interface LogEntry {
    id: string;
    timestamp: string;
    level: string;
    message: string;
    metadata?: Record<string, any>;
}

interface TrafficEvent {
    id: string;
    timestamp: string;
    method: string;
    duration?: string;
    status: 'success' | 'error';
    requestPayload?: any;
    responsePayload?: any;
    error?: string;
}

export function InspectorView() {
    const [events, setEvents] = useState<TrafficEvent[]>([]);
    const [selectedEventId, setSelectedEventId] = useState<string | null>(null);
    const [isPaused, setIsPaused] = useState(false);
    const [searchQuery, setSearchQuery] = useState("");
    const [isConnected, setIsConnected] = useState(false);

    const wsRef = useRef<WebSocket | null>(null);
    const isPausedRef = useRef(isPaused);

    useEffect(() => {
        isPausedRef.current = isPaused;
    }, [isPaused]);

    useEffect(() => {
        const connect = () => {
            const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
            const host = window.location.host;
            const wsUrl = `${protocol}//${host}/api/v1/ws/logs`;
            const ws = new WebSocket(wsUrl);

            ws.onopen = () => setIsConnected(true);
            ws.onclose = () => {
                setIsConnected(false);
                setTimeout(connect, 3000);
            };
            ws.onmessage = (event) => {
                if (isPausedRef.current) return;
                try {
                    const log: LogEntry = JSON.parse(event.data);
                    // Filter for traffic logs
                    // We look for "Request completed" or "Request failed"
                    if (log.message === "Request completed" || log.message === "Request failed") {
                        const trafficEvent: TrafficEvent = {
                            id: log.id,
                            timestamp: log.timestamp,
                            method: log.metadata?.method as string || "unknown",
                            duration: log.metadata?.duration as string,
                            status: log.message === "Request completed" ? "success" : "error",
                            requestPayload: log.metadata?.request_payload ? parsePayload(log.metadata.request_payload) : undefined,
                            responsePayload: log.metadata?.response_payload ? parsePayload(log.metadata.response_payload) : undefined,
                            error: log.metadata?.error as string,
                        };
                        setEvents(prev => [trafficEvent, ...prev].slice(0, 1000));
                    }
                } catch (e) {
                    console.error("Failed to parse log", e);
                }
            };
            wsRef.current = ws;
        };
        connect();
        return () => wsRef.current?.close();
    }, []);

    const parsePayload = (payload: any) => {
        if (typeof payload === 'string') {
            try {
                return JSON.parse(payload);
            } catch {
                return payload;
            }
        }
        return payload;
    };

    const filteredEvents = useMemo(() => {
        if (!searchQuery) return events;
        const q = searchQuery.toLowerCase();
        return events.filter(e =>
            e.method.toLowerCase().includes(q) ||
            e.status.includes(q)
        );
    }, [events, searchQuery]);

    const selectedEvent = events.find(e => e.id === selectedEventId) || events[0] || null;

    return (
        <div className="flex flex-col h-full bg-background">
            {/* Header */}
            <div className="flex items-center justify-between p-4 border-b">
                <div className="flex items-center gap-2">
                    <Activity className="h-5 w-5 text-primary" />
                    <h1 className="text-lg font-semibold">Traffic Inspector</h1>
                    <Badge variant={isConnected ? "outline" : "destructive"} className="ml-2 font-mono text-xs">
                        {isConnected ? "Live" : "Disconnected"}
                    </Badge>
                    <Badge variant="secondary" className="font-mono text-xs">
                        {events.length} events
                    </Badge>
                </div>
                <div className="flex items-center gap-2">
                    <div className="relative w-64">
                        <Search className="absolute left-2 top-2.5 h-4 w-4 text-muted-foreground" />
                        <Input
                            placeholder="Filter requests..."
                            className="pl-8 h-9"
                            value={searchQuery}
                            onChange={(e) => setSearchQuery(e.target.value)}
                        />
                    </div>
                    <Button variant="outline" size="sm" onClick={() => setIsPaused(!isPaused)}>
                        {isPaused ? <Play className="h-4 w-4 mr-2" /> : <Pause className="h-4 w-4 mr-2" />}
                        {isPaused ? "Resume" : "Pause"}
                    </Button>
                    <Button variant="outline" size="sm" onClick={() => setEvents([])}>
                        <Trash2 className="h-4 w-4 mr-2" /> Clear
                    </Button>
                </div>
            </div>

            {/* Content */}
            <ResizablePanelGroup direction="horizontal" className="flex-1">
                <ResizablePanel defaultSize={35} minSize={20} maxSize={50}>
                    <ScrollArea className="h-full">
                        <div className="flex flex-col p-2 gap-1">
                            {filteredEvents.length === 0 && (
                                <div className="text-center text-muted-foreground py-8 text-sm">
                                    No traffic events captured.
                                    <br />
                                    Make some requests to see them here.
                                </div>
                            )}
                            {filteredEvents.map(event => (
                                <div
                                    key={event.id}
                                    onClick={() => setSelectedEventId(event.id)}
                                    className={cn(
                                        "flex flex-col p-3 rounded-md cursor-pointer transition-colors border border-transparent hover:bg-muted/50",
                                        (selectedEventId === event.id || (!selectedEventId && event === events[0])) ? "bg-muted border-border" : ""
                                    )}
                                >
                                    <div className="flex items-center justify-between mb-1">
                                        <div className="flex items-center gap-2">
                                            {event.status === 'success' ?
                                                <CheckCircle2 className="h-4 w-4 text-green-500" /> :
                                                <XCircle className="h-4 w-4 text-red-500" />
                                            }
                                            <span className="font-mono text-sm font-medium truncate w-40" title={event.method}>{event.method}</span>
                                        </div>
                                        <span className="text-xs text-muted-foreground font-mono">{event.duration}</span>
                                    </div>
                                    <div className="flex items-center justify-between">
                                        <span className="text-[10px] text-muted-foreground">
                                            {new Date(event.timestamp).toLocaleTimeString()}
                                        </span>
                                        <Badge variant="outline" className="text-[10px] px-1 py-0 h-4">
                                            RPC
                                        </Badge>
                                    </div>
                                </div>
                            ))}
                        </div>
                    </ScrollArea>
                </ResizablePanel>

                <ResizableHandle />

                <ResizablePanel defaultSize={65}>
                    {selectedEvent ? (
                        <div className="h-full flex flex-col">
                            <div className="p-4 border-b bg-muted/10">
                                <div className="flex items-center gap-2 mb-2">
                                    <Badge variant={selectedEvent.status === 'success' ? 'default' : 'destructive'}>
                                        {selectedEvent.status.toUpperCase()}
                                    </Badge>
                                    <span className="font-mono text-lg font-medium break-all">{selectedEvent.method}</span>
                                </div>
                                <div className="flex gap-4 text-sm text-muted-foreground">
                                    <span className="flex items-center gap-1">
                                        <Clock className="h-3 w-3" /> {selectedEvent.duration}
                                    </span>
                                    <span className="flex items-center gap-1 font-mono">
                                        ID: {selectedEvent.id}
                                    </span>
                                </div>
                                {selectedEvent.error && (
                                    <div className="mt-2 p-2 bg-red-50 dark:bg-red-900/20 text-red-600 dark:text-red-400 rounded text-sm font-mono border border-red-200 dark:border-red-800">
                                        Error: {selectedEvent.error}
                                    </div>
                                )}
                            </div>

                            <div className="flex-1 overflow-hidden grid grid-rows-2">
                                <div className="flex flex-col border-b overflow-hidden">
                                    <div className="px-4 py-2 bg-muted/20 border-b text-xs font-medium flex items-center gap-2">
                                        <ArrowDown className="h-3 w-3 text-blue-500" /> Request Payload
                                    </div>
                                    <ScrollArea className="flex-1 p-4">
                                        <JsonView data={selectedEvent.requestPayload || {}} />
                                    </ScrollArea>
                                </div>
                                <div className="flex flex-col overflow-hidden">
                                    <div className="px-4 py-2 bg-muted/20 border-b text-xs font-medium flex items-center gap-2">
                                        <ArrowUp className="h-3 w-3 text-green-500" /> Response Payload
                                    </div>
                                    <ScrollArea className="flex-1 p-4">
                                        <JsonView data={selectedEvent.responsePayload || {}} />
                                    </ScrollArea>
                                </div>
                            </div>
                        </div>
                    ) : (
                        <div className="h-full flex items-center justify-center text-muted-foreground flex-col gap-2">
                            <Activity className="h-12 w-12 opacity-20" />
                            <p>Select an event to view details</p>
                        </div>
                    )}
                </ResizablePanel>
            </ResizablePanelGroup>
        </div>
    );
}
