/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { useState, useEffect, useCallback, useMemo } from "react";
import { apiClient, ServiceHealth, HealthHistoryPoint } from "@/lib/client";
import { usePolling } from "@/hooks/use-polling";

// Re-export types for consumers
export type { ServiceHealth, HealthHistoryPoint };

/**
 * ServiceHistory maps service IDs to their list of historical health points.
 */
export interface ServiceHistory {
  [serviceId: string]: HealthHistoryPoint[];
}

/**
 * useServiceHealthHistory is a hook that fetches and maintains the health history of services.
 * It polls the backend API for health data (which now includes server-side history).
 *
 * @returns An object containing the current services list, their health history, and a loading state.
 */
export function useServiceHealthHistory() {
  const [history, setHistory] = useState<ServiceHistory>({});
  const [services, setServices] = useState<ServiceHealth[]>([]);
  const [isLoading, setIsLoading] = useState(true);

  const fetchHealth = useCallback(async () => {
    try {
      const data = await apiClient.getDashboardHealth();
      // Backend returns history keyed by ID
      setServices(data.services || []);
      setHistory(data.history || {});
    } catch (error) {
      console.warn("Failed to fetch health data", error);
    } finally {
      setIsLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchHealth();
  }, [fetchHealth]);

  // ⚡ BOLT: Refactored to use centralized usePolling hook for consistent lifecycle management and memory leak prevention.
  // Randomized Selection from Top 5 High-Impact Targets (Network Category)
  usePolling(fetchHealth, 10000);

  // ⚡ BOLT: Memoized return values to prevent unnecessary re-renders.
  // Randomized Selection from Top 5 High-Impact Targets
  return useMemo(
    () => ({ services, history, isLoading }),
    [services, history, isLoading]
  );
}
