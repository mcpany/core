/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { useState, useEffect, useCallback } from "react";
import { apiClient, ServiceHealth, HealthHistoryPoint } from "@/lib/client";
import { usePolling } from "@/hooks/use-polling";

// Re-export types for consumers
export type { ServiceHealth, HealthHistoryPoint };

/**
 * ServiceHistory represents the public ServiceHistory entity.
 *
 * Summary: Defines the structured data model representing a history.
 *
 * Parameters:
 *   - None.
 *
 * Returns:
 *   - None.
 *
 * Throws/Errors:
 *   - None.
 *
 * Side Effects:
 *   - None.
 */
export interface ServiceHistory {
  [serviceId: string]: HealthHistoryPoint[];
}

/**
 * useServiceHealthHistory serves as a public interface for interacting with useServiceHealthHistory.
 *
 * Summary: Use the service health history appropriately based on current system conditions.
 *
 * Parameters:
 *   - Refer to the function signature for strongly-typed input arguments.
 *
 * Returns:
 *   - Returns the expected domain model or execution state.
 *
 * Throws/Errors:
 *   - Propagates exceptions from underlying validation layers.
 *
 * Side Effects:
 *   - May mutate state or perform network I/O depending on implementation.
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

  return { services, history, isLoading };
}
