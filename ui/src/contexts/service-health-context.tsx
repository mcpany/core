/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { createContext, useContext, useEffect, useState, ReactNode, useRef } from 'react';
import { Graph, NodeType, NodeStatus } from '@/types/topology';

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
}

const ServiceHealthContext = createContext<ServiceHealthContextType | undefined>(undefined);

const MAX_HISTORY_POINTS = 30; // 30 points * 5s = 2.5 minutes history
const POLLING_INTERVAL = 5000;

/**
 * ServiceHealthProvider component.
 * @param props - The component props.
 * @param props.children - The child components.
 * @returns The rendered component.
 */
export function ServiceHealthProvider({ children }: { children: ReactNode }) {
    const [history, setHistory] = useState<Record<string, MetricPoint[]>>({});
    const [latestTopology, setLatestTopology] = useState<Graph | null>(null);

    useEffect(() => {
        const fetchTopology = async () => {
            if (document.hidden) return;
            try {
                // Handle relative URL for fetch in jsdom/test env
                const url = typeof window !== 'undefined' ? '/api/v1/topology' : 'http://localhost/api/v1/topology';
                const res = await fetch(url);
                if (!res.ok) return;
                const graph: Graph = await res.json();

                setLatestTopology(graph);

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

                    // Also initialize entries for services that didn't report (optional, or let them be empty)
                    // We only track what we see in topology.

                    return next;
                });

            } catch (e) {
                console.error("Failed to fetch topology for health history", e);
            }
        };

        void fetchTopology();
        const interval = setInterval(() => void fetchTopology(), POLLING_INTERVAL);

        const onVisibilityChange = () => {
            if (!document.hidden) void fetchTopology();
        };
        document.addEventListener("visibilitychange", onVisibilityChange);

        return () => {
            clearInterval(interval);
            document.removeEventListener("visibilitychange", onVisibilityChange);
        };
    }, []);

    const getServiceHistory = (serviceId: string) => {
        return history[serviceId] || [];
    };

    const getServiceCurrentHealth = (serviceId: string) => {
        const points = history[serviceId];
        return points && points.length > 0 ? points[points.length - 1] : null;
    };

    return (
        <ServiceHealthContext.Provider value={{ getServiceHistory, getServiceCurrentHealth, latestTopology }}>
            {children}
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
