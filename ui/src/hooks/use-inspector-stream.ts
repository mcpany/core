/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useEffect, useRef, useCallback } from "react";
import { LogEntry } from "@/components/logs/log-stream";

export interface MCPMessage {
    id: string;
    timestamp: string;
    direction: "inbound" | "outbound";
    method?: string;
    payload: Record<string, unknown>;
    isError?: boolean;
    source?: string;
}

export function useInspectorStream() {
    const [messages, setMessages] = useState<MCPMessage[]>([]);
    const [isConnected, setIsConnected] = useState(false);
    const [isPaused, setIsPaused] = useState(false);
    const wsRef = useRef<WebSocket | null>(null);

    // Simulation mode state
    const [isSimulating, setIsSimulating] = useState(false);
    const simulationIntervalRef = useRef<NodeJS.Timeout | null>(null);

    const parseLogToMCP = useCallback((log: LogEntry): MCPMessage | null => {
        try {
            // Attempt to parse the message as JSON
            let payload: any; // eslint-disable-line @typescript-eslint/no-explicit-any
            if (typeof log.message === 'string' && (log.message.startsWith('{') || log.message.startsWith('['))) {
                 try {
                     payload = JSON.parse(log.message);
                 } catch {
                     return null;
                 }
            } else {
                return null;
            }

            // Check if it looks like JSON-RPC
            // MCP messages usually have "jsonrpc": "2.0"
            if (!payload || typeof payload !== 'object') return null;

            // Heuristic for MCP messages
            const isJsonRpc = payload.jsonrpc === "2.0";
            if (!isJsonRpc) return null;

            // Determine direction based on log source or content if possible
            const direction: "inbound" | "outbound" = "inbound";
            const method = payload.method;

            let methodDisplay = method;
            if (!methodDisplay) {
                if (payload.result) {
                    methodDisplay = "Response";
                } else if (payload.error) {
                    methodDisplay = "Error";
                } else {
                    methodDisplay = "Notification";
                }
            }

            return {
                id: payload.id || `notify-${Date.now()}-${Math.random()}`,
                timestamp: log.timestamp,
                direction: direction,
                method: methodDisplay,
                payload: payload,
                isError: !!payload.error,
                source: log.source
            };

        } catch (_e) {
            return null;
        }
    }, []);

    useEffect(() => {
        if (isSimulating) return;

        const connect = () => {
            const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
            const host = window.location.host;
            // Fallback for dev environment if needed, but assuming relative path works
            const wsUrl = `${protocol}//${host}/api/v1/ws/logs`;

            const ws = new WebSocket(wsUrl);

            ws.onopen = () => setIsConnected(true);
            ws.onclose = () => {
                setIsConnected(false);
                if (!isSimulating) setTimeout(connect, 3000);
            };
            ws.onerror = (e) => console.error("Inspector WS Error", e);

            ws.onmessage = (event) => {
                if (isPaused) return;
                try {
                    const log: LogEntry = JSON.parse(event.data);
                    const mcpMsg = parseLogToMCP(log);
                    if (mcpMsg) {
                        setMessages(prev => [mcpMsg, ...prev].slice(0, 1000));
                    }
                } catch (_e) {
                    // ignore parse errors
                }
            };

            wsRef.current = ws;
        };

        connect();

        return () => {
            wsRef.current?.close();
        };
    }, [isPaused, isSimulating, parseLogToMCP]);

    const clearMessages = () => setMessages([]);

    const togglePause = () => setIsPaused(!isPaused);

    const startSimulation = () => {
        if (wsRef.current) wsRef.current.close();
        setIsSimulating(true);

        let simId = 1;
        simulationIntervalRef.current = setInterval(() => {
            if (isPaused) return;

            const methods = ["initialize", "tools/list", "tools/call", "resources/list", "prompts/list"];
            const method = methods[Math.floor(Math.random() * methods.length)];
            const isRequest = Math.random() > 0.5;

            const msg: MCPMessage = {
                id: isRequest ? `req-${simId++}` : `resp-${simId++}`,
                timestamp: new Date().toISOString(),
                direction: Math.random() > 0.5 ? "inbound" : "outbound",
                method: isRequest ? method : undefined,
                payload: isRequest ? {
                    jsonrpc: "2.0",
                    id: simId,
                    method: method,
                    params: {}
                } : {
                    jsonrpc: "2.0",
                    id: simId,
                    result: { status: "success" }
                },
                isError: Math.random() > 0.9,
                source: "simulation"
            };

            if (msg.isError) {
                msg.payload = {
                    jsonrpc: "2.0",
                    id: simId,
                    error: {
                        code: -32603,
                        message: "Internal error"
                    }
                };
            }

            setMessages(prev => [msg, ...prev].slice(0, 1000));
        }, 1000);
    };

    const stopSimulation = () => {
        if (simulationIntervalRef.current) clearInterval(simulationIntervalRef.current);
        setIsSimulating(false);
        // Reconnect logic will trigger via useEffect
    };

    return {
        messages,
        isConnected,
        isPaused,
        isSimulating,
        clearMessages,
        togglePause,
        startSimulation,
        stopSimulation
    };
}
