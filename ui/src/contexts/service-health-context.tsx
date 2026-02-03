/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { createContext, useContext, useEffect, useState, ReactNode, useRef, useCallback, useMemo, useSyncExternalStore } from 'react';
import { Graph, NodeStatus } from '@/types/topology';

/**
 * MetricPoint represents a single data point for service health metrics at a specific time.
 */
export interface MetricPoint {
    /** The timestamp of the metric point in milliseconds. */
    timestamp: number;
    /** The latency in milliseconds. */
    latencyMs: number;
    /** The error rate (0-1). */
    errorRate: number;
    /** Queries per second. */
    qps: number;
    /** The status of the node. */
    status: NodeStatus;
}

interface ServiceHealthContextType {
    /** @deprecated Use useServiceHistory(serviceId) instead for better performance. */
    getServiceHistory: (serviceId: string) => MetricPoint[];
    /** @deprecated Use useServiceHistory(serviceId) instead. */
    getServiceCurrentHealth: (serviceId: string) => MetricPoint | null;
    latestTopology: Graph | null;
    refreshTopology: () => Promise<void>;
}

interface TopologyContextType {
    latestTopology: Graph | null;
    refreshTopology: () => Promise<void>;
}

const ServiceHealthContext = createContext<ServiceHealthContextType | undefined>(undefined);
const TopologyContext = createContext<TopologyContextType | undefined>(undefined);

// Internal Store Context to allow useServiceHistory to access the store
const ServiceHealthStoreContext = createContext<HealthStore | undefined>(undefined);

/** Maximum number of history points to keep (30 points * 5s = 2.5 minutes). */
const MAX_HISTORY_POINTS = 30;

/** Polling interval in milliseconds (5 seconds). */
const POLLING_INTERVAL = 5000;

const EMPTY_HISTORY: MetricPoint[] = [];

type Listener = () => void;

class HealthStore {
    private history: Record<string, MetricPoint[]> = {};
    private listeners: Set<Listener> = new Set();

    getHistory(serviceId: string): MetricPoint[] {
        return this.history[serviceId] || EMPTY_HISTORY;
    }

    getAllHistory(): Record<string, MetricPoint[]> {
        return this.history;
    }

    update(newPoints: Record<string, MetricPoint>) {
        let changed = false;
        Object.entries(newPoints).forEach(([id, point]) => {
            const currentPoints = this.history[id] || EMPTY_HISTORY;
            // Optimization: Only update if strictly needed, but for history we always append.
            // But we can check if the new point is different from the last one?
            // Usually we always append time series data.

            const newHistory = [...currentPoints, point];
            if (newHistory.length > MAX_HISTORY_POINTS) {
                newHistory.shift();
            }
            this.history[id] = newHistory;
            changed = true;
        });

        // Also ensure we keep existing history for services not in newPoints?
        // The original code did: next = { ...prev }; update...
        // So yes, we keep others.

        // We technically create a new object for this.history to be safe?
        // Or just mutate and notify. useSyncExternalStore handles snapshots.
        // To be safe with concurrent rendering, we should treat this.history as immutable or snapshot.
        if (changed) {
            this.history = { ...this.history }; // Shallow copy to ensure reference change
            this.notify();
        }
    }

    subscribe(listener: Listener) {
        this.listeners.add(listener);
        return () => this.listeners.delete(listener);
    }

    private notify() {
        this.listeners.forEach(l => l());
    }
}

/**
 * ServiceHealthProvider component.
 * @param props - The component props.
 * @param props.children - The child components.
 * @returns The rendered component.
 */
export function ServiceHealthProvider({ children }: { children: ReactNode }) {
    // ⚡ Bolt: Use a stable store ref to manage history without triggering provider re-renders.
    // Randomized Selection from Top 5 High-Impact Targets.
    const store = useRef(new HealthStore()).current;

    const [latestTopology, setLatestTopology] = useState<Graph | null>(null);
    const lastTopologyText = useRef<string>('');
    const lastGraph = useRef<Graph | null>(null);
    const lastEtag = useRef<string | null>(null);

    const fetchTopology = useCallback(async () => {
        try {
            // Handle relative URL for fetch in jsdom/test env
            const url = typeof window !== 'undefined' ? '/api/v1/topology' : 'http://localhost/api/v1/topology';

            const headers: HeadersInit = {};
            if (lastEtag.current) {
                headers['If-None-Match'] = lastEtag.current;
            }

            // Inject Auth Token if available
            if (typeof window !== 'undefined') {
                const token = localStorage.getItem('mcp_auth_token');
                if (token) {
                    headers['Authorization'] = `Basic ${token}`;
                }
            }

            const res = await fetch(url, { headers });

            if (res.status === 304 && lastGraph.current) {
                return;
            }

            if (!res.ok) return;

            const etag = res.headers.get('ETag');
            if (etag) {
                lastEtag.current = etag;
            }

            const text = await res.text();
            let graph: Graph;

            if (text === lastTopologyText.current && lastGraph.current) {
                graph = lastGraph.current;
            } else {
                graph = JSON.parse(text);
                lastTopologyText.current = text;
                lastGraph.current = graph;
                setLatestTopology(graph);
            }

            const now = Date.now();
            const newPoints: Record<string, MetricPoint> = {};

            // Helper to extract service nodes
            const extractServiceNodes = (nodes: any[]) => {
                nodes.forEach(node => {
                    if (node.type === 'NODE_TYPE_SERVICE') {
                        newPoints[node.id] = {
                            timestamp: now,
                            latencyMs: node.metrics?.latencyMs || 0,
                            errorRate: node.metrics?.errorRate || 0,
                            qps: node.metrics?.qps || 0,
                            status: node.status || 'NODE_STATUS_UNSPECIFIED'
                        };
                    }
                    if (node.children) {
                        extractServiceNodes(node.children);
                    }
                });
            };

            if (graph.core) {
                extractServiceNodes([graph.core]);
                if (graph.core.children) extractServiceNodes(graph.core.children);
            }

            // Update store instead of state
            store.update(newPoints);

        } catch (e) {
            console.error("Failed to fetch topology for health history", e);
        }
    }, [store]);

    useEffect(() => {
        void fetchTopology();
        const interval = setInterval(() => {
             if (!document.hidden) {
                 void fetchTopology();
             }
        }, POLLING_INTERVAL);

        const onVisibilityChange = () => {
            if (!document.hidden) void fetchTopology();
        };
        document.addEventListener("visibilitychange", onVisibilityChange);

        return () => {
            clearInterval(interval);
            document.removeEventListener("visibilitychange", onVisibilityChange);
        };
    }, [fetchTopology]);

    // Backward compatibility helpers
    // These WILL trigger re-renders if used, but we'll migrate to useServiceHistory.
    // However, to make them reactive, we need to subscribe the Context value?
    // Or we just return functions that get from store?
    // If we return functions that get from store, they won't trigger re-render when data changes,
    // so the consumer won't update!
    // So for backward compatibility, we can't fully support reactivity via these methods unless
    // the consumer uses `useServiceHistory` or we simulate the old behavior.
    //
    // BUT, `useServiceHealth` hook returns this context.
    // If we want `useServiceHealth` to still cause re-renders for old code, we can put a listener in `useServiceHealth`.
    //
    // For now, let's keep the context value stable (except topology) and fix `useServiceHealth` to wrap the store.

    const value = useMemo(() => ({
        // These are just accessors now. They don't make the component reactive by themselves.
        // Reactivity will be handled by the hook wrapper.
        getServiceHistory: (serviceId: string) => store.getHistory(serviceId),
        getServiceCurrentHealth: (serviceId: string) => {
            const points = store.getHistory(serviceId);
            return points && points.length > 0 ? points[points.length - 1] : null;
        },
        latestTopology,
        refreshTopology: fetchTopology
    }), [latestTopology, fetchTopology, store]);

    const topologyValue = useMemo(() => ({
        latestTopology,
        refreshTopology: fetchTopology
    }), [latestTopology, fetchTopology]);

    return (
        <ServiceHealthStoreContext.Provider value={store}>
            <ServiceHealthContext.Provider value={value}>
                <TopologyContext.Provider value={topologyValue}>
                    {children}
                </TopologyContext.Provider>
            </ServiceHealthContext.Provider>
        </ServiceHealthStoreContext.Provider>
    );
}

/**
 * useServiceHealth is a hook to access service health history and current status.
 * @returns The service health context.
 * @throws Error if used outside of a ServiceHealthProvider.
 * @deprecated Use `useServiceHistory` for specific services or `useTopology` for graph.
 */
export function useServiceHealth() {
    const context = useContext(ServiceHealthContext);
    const store = useContext(ServiceHealthStoreContext);

    if (!context || !store) {
        throw new Error("useServiceHealth must be used within a ServiceHealthProvider");
    }

    // ⚡ Bolt: Maintain backward compatibility reactivity.
    // We subscribe to the store and force a re-render when ANY history changes.
    // This mimics the old behavior (global re-render) but allows us to migrate components one by one.
    const subscribe = useCallback((cb: Listener) => store.subscribe(cb), [store]);
    const history = useSyncExternalStore(
        subscribe,
        () => store.getAllHistory(),
        () => store.getAllHistory()
    );

    // We wrap the context methods to use the reactive history we just subscribed to.
    // Actually, `history` here is just to force re-render.
    // The methods in `context` read from `store` directly.

    return useMemo(() => ({
        ...context,
        // We override these to be explicit, though context already has them.
        // The dependency on `history` ensures we return a new object when history changes,
        // causing the consumer to re-render.
        getServiceHistory: (serviceId: string) => history[serviceId] || [],
        getServiceCurrentHealth: (serviceId: string) => {
            const points = history[serviceId] || [];
             return points.length > 0 ? points[points.length - 1] : null;
        }
    }), [context, history]);
}

/**
 * useTopology is a hook to access network topology.
 * It is optimized to not re-render when health metrics update.
 * @returns The topology context.
 * @throws Error if used outside of a ServiceHealthProvider.
 */
export function useTopology() {
    const context = useContext(TopologyContext);
    if (!context) {
        throw new Error("useTopology must be used within a ServiceHealthProvider");
    }
    return context;
}

/**
 * useServiceHistory is a highly optimized hook to subscribe to a single service's history.
 * It only triggers a re-render when the specific service's history changes.
 * @param serviceId - The ID of the service to monitor.
 * @returns The list of metric points for the service.
 */
export function useServiceHistory(serviceId: string) {
    const store = useContext(ServiceHealthStoreContext);
    if (!store) {
        throw new Error("useServiceHistory must be used within a ServiceHealthProvider");
    }

    const subscribe = useCallback((cb: Listener) => store.subscribe(cb), [store]);
    const history = useSyncExternalStore(
        subscribe,
        () => store.getHistory(serviceId),
        () => EMPTY_HISTORY // Server snapshot
    );

    return history;
}
