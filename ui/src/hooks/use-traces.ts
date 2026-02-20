/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { useEffect, useState, useRef } from "react";
import { Trace } from "@/types/trace";

interface UseTracesOptions {
    initialPaused?: boolean;
}

const MAX_TRACES = 1000;

/**
 * Hook to manage trace subscriptions via WebSocket.
 *
 * @param options - Configuration options for the trace hook.
 * @param options.initialPaused - Whether to start in a paused state.
 * @returns An object containing the current traces, loading state, connection status, and controls.
 */
export function useTraces(options: UseTracesOptions = {}) {
    const [traces, setTraces] = useState<Trace[]>([]);
    const [loading, setLoading] = useState(true);
    const [isConnected, setIsConnected] = useState(false);
    const [isPaused, setIsPaused] = useState(options.initialPaused || false);
    const wsRef = useRef<WebSocket | null>(null);
    const isPausedRef = useRef(isPaused);
    const isMountedRef = useRef(true);
    const reconnectTimeoutRef = useRef<NodeJS.Timeout | null>(null);
    const bufferRef = useRef<Trace[]>([]);

    useEffect(() => {
        isPausedRef.current = isPaused;
    }, [isPaused]);

    useEffect(() => {
        const interval = setInterval(() => {
            if (bufferRef.current.length === 0) return;

            const buffer = bufferRef.current;
            bufferRef.current = [];

            setTraces((prev) => {
                const updatesMap = new Map<string, Trace>();
                for (const t of buffer) {
                    updatesMap.set(t.id, t);
                }

                const existingIds = new Set(prev.map(t => t.id));
                const inserts: Trace[] = [];
                const updatesForExisting = new Map<string, Trace>();

                for (const t of updatesMap.values()) {
                    if (existingIds.has(t.id)) {
                        updatesForExisting.set(t.id, t);
                    } else {
                        inserts.push(t);
                    }
                }

                let nextTraces = prev;
                if (updatesForExisting.size > 0) {
                    nextTraces = prev.map(t => {
                        if (updatesForExisting.has(t.id)) {
                            return updatesForExisting.get(t.id)!;
                        }
                        return t;
                    });
                }

                const merged = [...inserts.reverse(), ...nextTraces];

                if (merged.length > MAX_TRACES) {
                    return merged.slice(0, MAX_TRACES);
                }
                return merged;
            });
        }, 100);

        return () => clearInterval(interval);
    }, []);

    const connect = () => {
        if (!isMountedRef.current) return;

        setLoading(true);
        if (typeof window === 'undefined') return;

        const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
        const host = window.location.host;
        const wsUrl = `${protocol}//${host}/api/v1/ws/traces`;

        if (wsRef.current) {
            wsRef.current.close();
        }

        const ws = new WebSocket(wsUrl);

        ws.onopen = () => {
            if (!isMountedRef.current) {
                ws.close();
                return;
            }
            setIsConnected(true);
            setLoading(false);
        };

        ws.onmessage = (event) => {
            if (!isMountedRef.current) return;
            if (isPausedRef.current) return;
            try {
                const trace: Trace = JSON.parse(event.data);
                bufferRef.current.push(trace);
            } catch (e) {
                console.error("Failed to parse trace", e);
            }
        };

        ws.onclose = () => {
            if (!isMountedRef.current) return;
            setIsConnected(false);
            if (reconnectTimeoutRef.current) clearTimeout(reconnectTimeoutRef.current);
            reconnectTimeoutRef.current = setTimeout(connect, 3000);
        };

        ws.onerror = (err) => {
            console.error("WebSocket error", err);
            ws.close();
        };

        wsRef.current = ws;
    };

    useEffect(() => {
        isMountedRef.current = true;
        connect();
        return () => {
            isMountedRef.current = false;
            if (wsRef.current) {
                wsRef.current.onclose = null;
                wsRef.current.close();
            }
            if (reconnectTimeoutRef.current) {
                clearTimeout(reconnectTimeoutRef.current);
            }
        };
    }, []);

    const clearTraces = () => setTraces([]);

    const refresh = () => {
        setTraces([]);
        connect();
    };

    return {
        traces,
        loading,
        isConnected,
        isPaused,
        setIsPaused,
        clearTraces,
        refresh
    };
}
