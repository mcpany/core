/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { useState, useEffect, useCallback } from "react";
import { apiClient } from "@/lib/client";

/**
 * Represents the full topology graph.
 */
export interface TopologyGraph {
    clients: Node[];
    core: Node;
}

/**
 * Represents a node in the topology graph.
 */
export interface Node {
    id: string;
    label: string;
    type: NodeType;
    status: NodeStatus;
    metadata: Record<string, string>;
    children: Node[];
    metrics?: NodeMetrics;
}

/**
 * Enum for node types.
 */
export enum NodeType {
    NODE_TYPE_UNSPECIFIED = 0,
    NODE_TYPE_CLIENT = 1,
    NODE_TYPE_CORE = 2,
    NODE_TYPE_SERVICE = 3,
    NODE_TYPE_TOOL = 4,
    NODE_TYPE_RESOURCE = 5,
    NODE_TYPE_PROMPT = 6,
    NODE_TYPE_API_CALL = 7,
    NODE_TYPE_MIDDLEWARE = 8,
    NODE_TYPE_WEBHOOK = 9,
}

/**
 * Enum for node status.
 */
export enum NodeStatus {
    NODE_STATUS_UNSPECIFIED = 0,
    NODE_STATUS_ACTIVE = 1,
    NODE_STATUS_INACTIVE = 2,
    NODE_STATUS_ERROR = 3,
}

/**
 * Metrics associated with a node.
 */
export interface NodeMetrics {
    qps: number;
    latencyMs: number;
    errorRate: number;
}

/**
 * Hook to fetch and manage topology data.
 * @param pollInterval - Optional interval in ms to poll for updates.
 * @returns The topology graph state and control functions.
 */
export function useTopology(pollInterval: number | null = null) {
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
        if (pollInterval === null || pollInterval <= 0) return;

        const interval = setInterval(() => {
            fetchTopology(true);
        }, pollInterval);

        return () => clearInterval(interval);
    }, [pollInterval, fetchTopology]);

    return { graph, loading, error, refresh: () => fetchTopology(false) };
}
