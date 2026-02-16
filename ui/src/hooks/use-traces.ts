/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { useEffect, useState, useRef, useMemo } from "react";
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
    // ⚡ BOLT: Switched to Map for O(1) updates and efficient deduplication.
    // Randomized Selection from Top 5 High-Impact Targets
    const [traceMap, setTraceMap] = useState<Map<string, Trace>>(new Map());

    // Derived state: Convert map to array and reverse for newest-first display.
    // This is still O(N) but avoids the O(N) deduplication steps in the update loop.
    const traces = useMemo(() => Array.from(traceMap.values()).reverse(), [traceMap]);

    const [loading, setLoading] = useState(true);
    const [isConnected, setIsConnected] = useState(false);
    const [isPaused, setIsPaused] = useState(options.initialPaused || false);
    const wsRef = useRef<WebSocket | null>(null);
    const isPausedRef = useRef(isPaused);
    const isMountedRef = useRef(true);
    const reconnectTimeoutRef = useRef<NodeJS.Timeout | null>(null);

    // ⚡ BOLT: Buffer for batched updates to avoid main thread blocking
    // Randomized Selection from Top 5 High-Impact Targets
    const bufferRef = useRef<Trace[]>([]);

    useEffect(() => {
        isPausedRef.current = isPaused;
    }, [isPaused]);

    // ⚡ BOLT: Flush buffer periodically
    useEffect(() => {
        const interval = setInterval(() => {
            if (bufferRef.current.length === 0) return;

            // Take current buffer and clear it immediately
            const buffer = bufferRef.current;
            bufferRef.current = [];

            setTraceMap((prev) => {
                // Optimization: Map cloning is O(N) but faster than multiple array allocs.
                // V8 optimizes Map cloning.
                const next = new Map(prev);

                for (const t of buffer) {
                    // Map preserves insertion order for new keys.
                    // For existing keys, it updates value in-place without moving key.
                    // This means:
                    // - New traces (from buffer) are appended to the end of the Map.
                    // - Updated traces (from buffer) stay in their original position.
                    // When we reverse the array later, new traces appear at the top,
                    // and updated traces stay where they were relative to others.
                    next.set(t.id, t);
                }
                return next;
            });
        }, 100);

        return () => clearInterval(interval);
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
                // ⚡ BOLT: Push to buffer instead of updating state directly
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

    const clearTraces = () => setTraceMap(new Map());

    const refresh = () => {
        setTraceMap(new Map());
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
