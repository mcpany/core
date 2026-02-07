/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { createContext, useContext, useEffect, useState, ReactNode, useRef, useCallback, useMemo } from 'react';
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
    getServiceHistory: (serviceId: string) => MetricPoint[];
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

/** Maximum number of history points to keep (30 points * 5s = 2.5 minutes). */
const MAX_HISTORY_POINTS = 30;

/** Polling interval in milliseconds (5 seconds). */
const POLLING_INTERVAL = 5000;

/**
 * ServiceHealthProvider component.
 *
 * Summary: Provides context for service health metrics and network topology to child components.
 *
 * @param props - Object. The component props.
 * @param props.children - ReactNode. The child components to wrap.
 * @returns JSX.Element. The rendered provider component.
 *
 * Side Effects:
 *   - Polls the /api/v1/topology endpoint every 5 seconds.
 *   - Maintains state for health history and latest topology graph.
 */
export function ServiceHealthProvider({ children }: { children: ReactNode }) {
    const [history, setHistory] = useState<Record<string, MetricPoint[]>>({});
    const [latestTopology, setLatestTopology] = useState<Graph | null>(null);
    const lastTopologyText = useRef<string>('');
    const lastGraph = useRef<Graph | null>(null);
    const lastEtag = useRef<string | null>(null);

    const fetchTopology = useCallback(async () => {
        try {
            // Handle relative URL for fetch in jsdom/test env
            const url = typeof window !== 'undefined' ? '/api/v1/topology' : 'http://localhost/api/v1/topology';

            // ⚡ Bolt: Optimize Polling with ETag (If-None-Match).
            // Randomized Selection from Top 5 High-Impact Targets.
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
                // Not modified, use cached graph
                return;
            }

            if (!res.ok) return;

            const etag = res.headers.get('ETag');
            if (etag) {
                lastEtag.current = etag;
            }

            // ⚡ Bolt Optimization: Use text comparison to avoid expensive JSON operations.
            // res.text() + string comparison is much faster than res.json() + JSON.stringify().
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
            // Using 'any' for node because TopologyNode type from types/topology
            // might not match exactly what comes from API or recursion needs to be flexible
            // But we should try to be safer if possible.
            // Assuming Node type from @/types/topology
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

            // Update history
            setHistory(prev => {
                const next = { ...prev };
                Object.entries(newPoints).forEach(([id, point]) => {
                    const points = next[id] ? [...next[id], point] : [point];
                    if (points.length > MAX_HISTORY_POINTS) {
                        points.shift();
                    }
                    next[id] = points;
                });

                return next;
            });

        } catch (e) {
            console.error("Failed to fetch topology for health history", e);
        }
    }, []);

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

    const getServiceHistory = useCallback((serviceId: string) => {
        return history[serviceId] || [];
    }, [history]);

    const getServiceCurrentHealth = useCallback((serviceId: string) => {
        const points = history[serviceId];
        return points && points.length > 0 ? points[points.length - 1] : null;
    }, [history]);

    const value = useMemo(() => ({
        getServiceHistory,
        getServiceCurrentHealth,
        latestTopology,
        refreshTopology: fetchTopology
    }), [getServiceHistory, getServiceCurrentHealth, latestTopology, fetchTopology]);

    // ⚡ Bolt Optimization: Split context for topology to avoid re-renders on metrics updates
    const topologyValue = useMemo(() => ({
        latestTopology,
        refreshTopology: fetchTopology
    }), [latestTopology, fetchTopology]);

    return (
        <ServiceHealthContext.Provider value={value}>
            <TopologyContext.Provider value={topologyValue}>
                {children}
            </TopologyContext.Provider>
        </ServiceHealthContext.Provider>
    );
}

/**
 * useServiceHealth is a hook to access service health history and current status.
 * @returns The service health context.
 * @throws Error if used outside of a ServiceHealthProvider.
 */
export function useServiceHealth() {
    const context = useContext(ServiceHealthContext);
    if (!context) {
        throw new Error("useServiceHealth must be used within a ServiceHealthProvider");
    }
    return context;
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
