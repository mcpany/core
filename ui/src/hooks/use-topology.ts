/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useEffect, useCallback, useRef } from "react";
import { apiClient } from "@/lib/client";

/**
 * TopologyNode represents a node in the network topology graph.
 */
export interface TopologyNode {
    id: string;
    label: string;
    type: string;
    status: string;
    metadata?: Record<string, string>;
    children?: TopologyNode[];
    metrics?: {
        qps: number;
        latencyMs: number;
        errorRate: number;
    };
}

/**
 * TopologyGraph represents the full network topology.
 */
export interface TopologyGraph {
    clients: TopologyNode[];
    core: TopologyNode;
}

/**
 * useTopology hook fetches the network topology data.
 * @param isLive - Whether to poll for updates.
 * @returns The topology graph state.
 */
export function useTopology(isLive: boolean = false) {
    const [graph, setGraph] = useState<TopologyGraph | null>(null);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);
    const pollIntervalRef = useRef<NodeJS.Timeout | null>(null);

    const fetchTopology = useCallback(async () => {
        try {
            // Don't set loading true on poll to avoid flicker
            if (!graph) setLoading(true);
            const data = await apiClient.getTopology();
            setGraph(data);
            setError(null);
        } catch (err) {
            console.error("Failed to fetch topology", err);
            setError("Failed to load topology data.");
        } finally {
            setLoading(false);
        }
    }, [graph]);

    useEffect(() => {
        fetchTopology();
    }, []);

    useEffect(() => {
        if (isLive) {
            pollIntervalRef.current = setInterval(fetchTopology, 3000);
        } else {
            if (pollIntervalRef.current) {
                clearInterval(pollIntervalRef.current);
                pollIntervalRef.current = null;
            }
        }
        return () => {
            if (pollIntervalRef.current) {
                clearInterval(pollIntervalRef.current);
            }
        };
    }, [isLive, fetchTopology]);

    return { graph, loading, error, refresh: fetchTopology };
}
