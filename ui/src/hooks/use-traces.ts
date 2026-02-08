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
 * Hook to manage trace subscriptions via WebSocket with historical data fetching.
 *
 * @param options - Configuration options for the trace hook.
 * @param options.initialPaused - Whether to start in a paused state.
 * @param options.fetchHistory - Whether to fetch initial history from REST API.
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
        if (!fetchHistory) return;
        try {
            const res = await fetch('/api/traces');
            if (res.ok) {
                const history: Trace[] = await res.json();
                if (isMountedRef.current) {
                    setTraces(prev => {
                         // Simple merge: Replace history, but what if WS updates came in?
                         // Ideally we want to merge.
                         // But for simplicity on "refresh" or "init", we can assume history is the base.
                         return history;
                    });
                }
            }
        } catch (e) {
            console.error("Failed to fetch historical traces", e);
        }
    }, [fetchHistory]);

    const connect = useCallback(() => {
        if (!isMountedRef.current) return;

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
                setTraces((prev) => {
                    // Deduplicate by ID
                    const index = prev.findIndex(t => t.id === trace.id);
                    if (index !== -1) {
                        const newTraces = [...prev];
                        newTraces[index] = trace;
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
            if (reconnectTimeoutRef.current) clearTimeout(reconnectTimeoutRef.current);
            reconnectTimeoutRef.current = setTimeout(connect, 3000);
        };

        ws.onerror = (err) => {
            console.error("WebSocket error", err);
            ws.close();
        };

        wsRef.current = ws;
    }, []);

    useEffect(() => {
        isMountedRef.current = true;

        const init = async () => {
            if (fetchHistory) {
                setLoading(true);
                await fetchHistoricalTraces();
            }
            connect();
        };

        init();

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
    }, [connect, fetchHistoricalTraces, fetchHistory]);

    const clearTraces = () => setTraces([]);

    const refresh = () => {
        setLoading(true);
        // Do not clear traces immediately to avoid flash?
        // Actually clear is better for explicit refresh.
        setTraces([]);

        // Re-fetch history
        fetchHistoricalTraces().then(() => {
             // Check connection status
             if (!wsRef.current || wsRef.current.readyState !== WebSocket.OPEN) {
                 connect();
             } else {
                 setLoading(false);
             }
        });
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
