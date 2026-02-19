/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { useState, useEffect, useCallback } from "react";
import { apiClient } from "@/lib/client";

/**
 * TopologyGraph represents the full topology graph structure.
 */
export interface TopologyGraph {
    clients: TopologyNode[];
    core: TopologyNode;
}

/**
 * TopologyNode represents a single node in the topology graph.
 */
export interface TopologyNode {
    id: string;
    label: string;
    type: string; // CLIENT, CORE, SERVICE, TOOL, etc.
    status: string; // ACTIVE, INACTIVE, ERROR
    metadata?: Record<string, string>;
    children?: TopologyNode[];
    metrics?: {
        qps: number;
        latencyMs: number;
        errorRate: number;
    };
}

/**
 * useTopology is a hook to fetch and manage the topology graph data.
 * @param pollIntervalMs - Interval in milliseconds to poll for updates. Set to null to disable polling.
 * @returns An object containing the graph data, loading state, error state, and a refresh function.
 */
export function useTopology(pollIntervalMs: number | null = 5000) {
    const [graph, setGraph] = useState<TopologyGraph | null>(null);
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState<Error | null>(null);

    const fetchTopology = useCallback(async (silent = false) => {
        if (!silent) setLoading(true);
        try {
            const data = await apiClient.getTopology();
            setGraph(data);
            setError(null);
        } catch (err) {
            console.error("Failed to fetch topology", err);
            setError(err instanceof Error ? err : new Error(String(err)));
        } finally {
            if (!silent) setLoading(false);
        }
    }, []);

    useEffect(() => {
        fetchTopology();
    }, [fetchTopology]);

    useEffect(() => {
        if (!pollIntervalMs) return;
        const interval = setInterval(() => {
            fetchTopology(true);
        }, pollIntervalMs);
        return () => clearInterval(interval);
    }, [pollIntervalMs, fetchTopology]);

    return { graph, loading, error, refresh: fetchTopology };
}
