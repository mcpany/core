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

            setTraces((prev) => {
                // ⚡ BOLT: Batched updates logic

                // 1. Deduplicate buffer (last write wins)
                const updatesMap = new Map<string, Trace>();
                for (const t of buffer) {
                    updatesMap.set(t.id, t);
                }

                // 2. Identify existing IDs for O(1) lookup
                const existingIds = new Set(prev.map(t => t.id));

                // 3. Separate new inserts from updates
                const inserts: Trace[] = [];
                const updatesForExisting = new Map<string, Trace>();

                for (const t of updatesMap.values()) {
                    if (existingIds.has(t.id)) {
                        updatesForExisting.set(t.id, t);
                    } else {
                        inserts.push(t);
                    }
                }

                // 4. Apply updates in-place to preserve order of existing items
                const nextTraces = prev.map(t => {
                    if (updatesForExisting.has(t.id)) {
                        return updatesForExisting.get(t.id)!;
                    }
                    return t;
                });

                // 5. Prepend new inserts (newest first).
                // Buffer is oldest->newest. We want newest at top of list.
                // So we reverse inserts.
                const merged = [...inserts.reverse(), ...nextTraces];

                // ⚡ BOLT: Cap the size of the traces array to prevent memory leaks.
                // Randomized Selection from Top 5 High-Impact Targets
                const MAX_TRACES = 1000;
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
