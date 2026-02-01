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

    // ⚡ Bolt Optimization: Buffer for incoming traces to support batch processing
    // Randomized Selection from Top 5 High-Impact Targets
    const traceBufferRef = useRef<Trace[]>([]);

    useEffect(() => {
        isPausedRef.current = isPaused;
    }, [isPaused]);

    // ⚡ Bolt Optimization: Periodic flush of trace buffer to reduce render cycles
    useEffect(() => {
        const flushInterval = setInterval(() => {
            if (traceBufferRef.current.length > 0) {
                // Capture buffer and clear ref immediately to avoid race conditions and side effects in setTraces
                const buffer = [...traceBufferRef.current];
                traceBufferRef.current = [];

                setTraces((prev) => {
                    if (buffer.length === 0) return prev;

                    // Optimization: Efficient merge strategy
                    // 1. Identify updates vs new items
                    const prevIds = new Set(prev.map(t => t.id));
                    const updates = new Map<string, Trace>();
                    const newItemsMap = new Map<string, Trace>();

                    // Process buffer (latest update in buffer wins)
                    buffer.forEach(t => {
                        if (prevIds.has(t.id)) {
                            updates.set(t.id, t);
                        } else {
                            newItemsMap.set(t.id, t);
                        }
                    });

                    // Convert new items map to array and reverse (newest first)
                    // Map preserves insertion order, so we reverse to put latest buffer items at top
                    const uniqueNewItems = Array.from(newItemsMap.values()).reverse();

                    // 2. Apply updates to existing items (preserve order)
                    let newPrev = prev;
                    if (updates.size > 0) {
                        newPrev = prev.map(t => updates.get(t.id) || t);
                    }

                    return [...uniqueNewItems, ...newPrev];
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
                // ⚡ Bolt Optimization: Push to buffer instead of setting state directly
                traceBufferRef.current.push(trace);
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
