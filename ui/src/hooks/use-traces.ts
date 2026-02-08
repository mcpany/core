/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { useEffect, useState, useRef, useCallback } from "react";
import { Trace } from "@/types/trace";

interface UseTracesOptions {
    initialPaused?: boolean;
    fetchHistory?: boolean;
}

/**
 * Hook to manage trace subscriptions via WebSocket.
 *
 * @param options - Configuration options for the trace hook.
 * @param options.initialPaused - Whether to start in a paused state.
 * @param options.fetchHistory - Whether to fetch historical traces on connect (default: true).
 * @returns An object containing the current traces, loading state, connection status, and controls.
 */
export function useTraces(options: UseTracesOptions = {}) {
    const { initialPaused = false, fetchHistory = true } = options;
    const [traces, setTraces] = useState<Trace[]>([]);
    const [loading, setLoading] = useState(true);
    const [isConnected, setIsConnected] = useState(false);
    const [isPaused, setIsPaused] = useState(initialPaused);
    const wsRef = useRef<WebSocket | null>(null);
    const isPausedRef = useRef(isPaused);
    const isMountedRef = useRef(true);
    const reconnectTimeoutRef = useRef<NodeJS.Timeout | null>(null);

    useEffect(() => {
        isPausedRef.current = isPaused;
    }, [isPaused]);

    const fetchHistoricalTraces = useCallback(async () => {
        try {
            const res = await fetch('/api/traces');
            if (res.ok) {
                const data: Trace[] = await res.json();
                if (isMountedRef.current) {
                    setTraces(prev => {
                        const map = new Map<string, Trace>();
                        // Keep existing live traces if any
                        prev.forEach(t => map.set(t.id, t));
                        // Merge history (only if not already present from WS)
                        data.forEach(t => {
                            if (!map.has(t.id)) {
                                map.set(t.id, t);
                            }
                        });
                        return Array.from(map.values()).sort((a, b) =>
                            new Date(b.timestamp).getTime() - new Date(a.timestamp).getTime()
                        );
                    });
                }
            }
        } catch (e) {
            console.error("Failed to fetch trace history", e);
        }
    }, []);

    const connect = useCallback(() => {
        if (!isMountedRef.current) return;

        setLoading(true);

        // Fetch history first if enabled
        if (fetchHistory) {
            fetchHistoricalTraces().finally(() => {
                if (isMountedRef.current) setLoading(false);
            });
        }

        // Use relative URL for client-side navigation, but handle both dev and prod
        if (typeof window === 'undefined') return;

        const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
        const host = window.location.host;
        const wsUrl = `${protocol}//${host}/api/v1/ws/traces`;

        // Cleanup previous
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
            if (!fetchHistory) setLoading(false);
        };

        ws.onmessage = (event) => {
            if (!isMountedRef.current) return;
            if (isPausedRef.current) return;
            try {
                const trace: Trace = JSON.parse(event.data);
                setTraces((prev) => {
                    // Deduplicate by ID
                    const index = prev.findIndex(t => t.id === trace.id);
                    if (index !== -1) {
                        const newTraces = [...prev];
                        newTraces[index] = trace;
                        // Re-sort? Probably not needed if we assume new traces are newer,
                        // but updates might change order if sort by duration (unlikely) or status.
                        // But we sort by timestamp. Trace updates (spans added) don't change start timestamp.
                        return newTraces;
                    }
                    return [trace, ...prev];
                });
            } catch (e) {
                console.error("Failed to parse trace", e);
            }
        };

        ws.onclose = () => {
            if (!isMountedRef.current) return;
            setIsConnected(false);
            // Reconnect after 3s
            if (reconnectTimeoutRef.current) clearTimeout(reconnectTimeoutRef.current);
            reconnectTimeoutRef.current = setTimeout(connect, 3000);
        };

        ws.onerror = (err) => {
            console.error("WebSocket error", err);
            ws.close();
        };

        wsRef.current = ws;
    }, [fetchHistory, fetchHistoricalTraces]);

    useEffect(() => {
        isMountedRef.current = true;
        connect();
        return () => {
            isMountedRef.current = false;
            if (wsRef.current) {
                wsRef.current.onclose = null; // Prevent reconnect trigger on manual close
                wsRef.current.close();
            }
            if (reconnectTimeoutRef.current) {
                clearTimeout(reconnectTimeoutRef.current);
            }
        };
    }, [connect]);

    const clearTraces = () => setTraces([]);

    const refresh = () => {
        // Clear traces and reconnect (which triggers history fetch again)
        setTraces([]);

        // Clear any pending reconnect
        if (reconnectTimeoutRef.current) {
            clearTimeout(reconnectTimeoutRef.current);
        }

        // Force reconnect logic immediately
        if (wsRef.current) {
            wsRef.current.onclose = null; // Prevent triggering onclose logic
            wsRef.current.close();
        }
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
