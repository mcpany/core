/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { useEffect, useState, useRef } from "react";
import { Trace } from "@/types/trace";

interface UseTracesOptions {
    initialPaused?: boolean;
}

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

    // ⚡ BOLT: Buffer for incoming traces to avoid rapid re-renders.
    // Randomized Selection from Top 5 High-Impact Targets
    const bufferRef = useRef<Trace[]>([]);

    useEffect(() => {
        isPausedRef.current = isPaused;
    }, [isPaused]);

    useEffect(() => {
        const flushInterval = setInterval(() => {
            if (bufferRef.current.length > 0) {
                const incoming = [...bufferRef.current];
                bufferRef.current = []; // Clear buffer immediately

                setTraces((prev) => {
                    // Optimization: Deduplicate efficiently using Map to avoid O(N*M) complexity
                    // Create a Map of incoming traces (latest wins if duplicate in buffer)
                    const incomingMap = new Map<string, Trace>();
                    incoming.forEach(t => incomingMap.set(t.id, t));

                    // Update existing traces in place if they are in the incoming batch
                    const nextTraces = prev.map(t => {
                        if (incomingMap.has(t.id)) {
                            const updated = incomingMap.get(t.id)!;
                            incomingMap.delete(t.id);
                            return updated;
                        }
                        return t;
                    });

                    // Prepend new traces (remaining in incomingMap)
                    // We reverse newTraces because incomingMap follows insertion order (oldest first),
                    // but we want the newest traces at the top of the list.
                    const newTraces = Array.from(incomingMap.values()).reverse();

                    // Optimization: Limit total traces to avoid unbounded memory growth
                    const MAX_TRACES = 5000;
                    const combined = [...newTraces, ...nextTraces];

                    if (combined.length > MAX_TRACES) {
                        return combined.slice(0, MAX_TRACES);
                    }
                    return combined;
                });
            }
        }, 100); // Flush every 100ms

        return () => clearInterval(flushInterval);
    }, []);

    const connect = () => {
        if (!isMountedRef.current) return;

        setLoading(true);
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
                // Optimization: Push to buffer instead of setting state directly
                bufferRef.current.push(trace);
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
    };

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
