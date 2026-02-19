/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { useState, useEffect, useRef } from "react";
import { apiClient } from "@/lib/client";
import { Graph } from "@proto/topology/v1/topology";

/**
 * Hook to fetch and manage topology data.
 * @param pollingInterval Interval in milliseconds for polling updates. Set to 0 to disable polling.
 * @returns Topology data and status.
 */
export function useTopology(pollingInterval = 5000) {
  const [graph, setGraph] = useState<Graph | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);
  const [isPaused, setIsPaused] = useState(false);

  const isMounted = useRef(true);

  useEffect(() => {
    isMounted.current = true;
    return () => {
      isMounted.current = false;
    };
  }, []);

  const fetchTopology = async () => {
    try {
      const data = await apiClient.getTopology();
      if (isMounted.current) {
        setGraph(data);
        setError(null);
      }
    } catch (err) {
      if (isMounted.current) {
        setError(err instanceof Error ? err : new Error("Failed to fetch topology"));
      }
    } finally {
      if (isMounted.current) {
        setLoading(false);
      }
    }
  };

  useEffect(() => {
    fetchTopology();

    if (pollingInterval > 0) {
      const interval = setInterval(() => {
        if (!isPaused) {
          fetchTopology();
        }
      }, pollingInterval);
      return () => clearInterval(interval);
    }
  }, [pollingInterval, isPaused]);

  return {
    graph,
    loading,
    error,
    isPaused,
    setIsPaused,
    refresh: fetchTopology
  };
}
