/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { createContext, useContext, useEffect, useState, ReactNode, useRef } from 'react';
import { Graph, NodeType, NodeStatus } from '@/types/topology';

export interface MetricPoint {
    timestamp: number;
    latencyMs: number;
    errorRate: number;
    qps: number;
    status: NodeStatus;
}

interface ServiceHealthContextType {
    getServiceHistory: (serviceId: string) => MetricPoint[];
    getServiceCurrentHealth: (serviceId: string) => MetricPoint | null;
}

const ServiceHealthContext = createContext<ServiceHealthContextType | undefined>(undefined);

const MAX_HISTORY_POINTS = 30; // 30 points * 5s = 2.5 minutes history
const POLLING_INTERVAL = 5000;

export function ServiceHealthProvider({ children }: { children: ReactNode }) {
    const [history, setHistory] = useState<Record<string, MetricPoint[]>>({});
    const historyRef = useRef<Record<string, MetricPoint[]>>({});

    useEffect(() => {
        const fetchTopology = async () => {
            if (document.hidden) return;
            try {
                const res = await fetch('/api/v1/topology');
                if (!res.ok) return;
                const graph: Graph = await res.json();

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

                    historyRef.current = next;
                    return next;
                });

            } catch (e) {
                console.error("Failed to fetch topology for health history", e);
            }
        };

        fetchTopology();
        const interval = setInterval(fetchTopology, POLLING_INTERVAL);

        const onVisibilityChange = () => {
            if (!document.hidden) fetchTopology();
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
        <ServiceHealthContext.Provider value={{ getServiceHistory, getServiceCurrentHealth }}>
            {children}
        </ServiceHealthContext.Provider>
    );
}

export function useServiceHealth() {
    const context = useContext(ServiceHealthContext);
    if (!context) {
        throw new Error("useServiceHealth must be used within a ServiceHealthProvider");
    }
    return context;
}
