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

    const loadHistory = useCallback(async () => {
        try {
            const res = await fetch('/api/traces');
            if (!res.ok) throw new Error('Failed to fetch traces');
            const data: Trace[] = await res.json();
            if (isMountedRef.current) {
                setTraces(data);
            }
        } catch (e) {
            console.error("Failed to load trace history", e);
        }
    }, []);

    const connect = useCallback(async () => {
        if (!isMountedRef.current) return;

        setLoading(true);

        if (fetchHistory) {
             await loadHistory();
        }

        if (!isMountedRef.current) return;

        // Use relative URL for client-side navigation, but handle both dev and prod
        // If window is undefined (SSR), don't connect
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
                        // Sort by timestamp descending
                        return newTraces.sort((a, b) => new Date(b.timestamp).getTime() - new Date(a.timestamp).getTime());
                    }
                    const newTraces = [trace, ...prev];
                    return newTraces.sort((a, b) => new Date(b.timestamp).getTime() - new Date(a.timestamp).getTime());
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
    }, [fetchHistory, loadHistory]);

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
        // Reload history + reconnect
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
