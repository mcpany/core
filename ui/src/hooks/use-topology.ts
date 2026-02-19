/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useEffect, useCallback } from "react";
import { apiClient } from "@/lib/client";

/**
 * TopologyNode represents a node in the topology graph.
 */
export interface TopologyNode {
    id: string;
    label: string;
    type: string;
    status: string;
    metadata: Record<string, string>;
    children?: TopologyNode[];
    metrics?: {
        qps: number;
        latencyMs: number;
        errorRate: number;
    };
}

/**
 * TopologyGraph represents the full topology graph.
 */
export interface TopologyGraph {
    clients: TopologyNode[];
    core: TopologyNode;
}

/**
 * useTopology hook fetches the topology graph and handles polling.
 * @param pollInterval - The interval in milliseconds to poll the API.
 * @returns The topology graph and loading/error state.
 */
export function useTopology(pollInterval = 5000) {
    const [graph, setGraph] = useState<TopologyGraph | null>(null);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);
    const [isPolling, setIsPolling] = useState(false);

    const fetchTopology = useCallback(async () => {
        try {
            const data = await apiClient.getTopology();
            setGraph(data);
            setError(null);
        } catch (err: any) {
            console.error("Failed to fetch topology", err);
            setError(err.message || "Failed to fetch topology");
        } finally {
            setLoading(false);
        }
    }, []);

    useEffect(() => {
        fetchTopology();
    }, [fetchTopology]);

    useEffect(() => {
        let interval: NodeJS.Timeout;
        if (isPolling) {
            interval = setInterval(fetchTopology, pollInterval);
        }
        return () => clearInterval(interval);
    }, [isPolling, pollInterval, fetchTopology]);

    return { graph, loading, error, isPolling, setIsPolling, refresh: fetchTopology };
}
