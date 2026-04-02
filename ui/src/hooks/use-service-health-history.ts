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
 * Summary: ServiceHistory maps service IDs to their list of historical health points.
 *
 * Params:
 *   - N/A: Type definition.
 *
 * Returns:
 *   - N/A: Type definition.
 *
 * Errors:
 *   - N/A: Type definition.
 *
 * Side Effects:
 *   - N/A: Type definition.
 */
export interface ServiceHistory {
  [serviceId: string]: HealthHistoryPoint[];
}

/**
 * Summary: Fetches and maintains the health status and historical data of services.
 *
 * Params:
 *   - None.
 *
 * Returns:
 *   - Object: Contains current services list, their health history, and an `isLoading` flag.
 *
 * Errors:
 *   - N/A: Network errors are suppressed in polling.
 *
 * Side Effects:
 *   - Polls the backend `apiClient.getServiceHealth()` periodically.
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
