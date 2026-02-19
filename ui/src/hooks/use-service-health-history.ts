/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { useServiceHealth } from "@/contexts/service-health-context";
import { ServiceHealth, HealthHistoryPoint } from "@/lib/client";

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
 *
 * ⚡ BOLT: Optimized to consume data from ServiceHealthContext instead of redundant polling.
 * Randomized Selection from Top 5 High-Impact Targets (Redundant Polling)
 *
 * @returns An object containing the current services list, their health history, and a loading state.
 */
export function useServiceHealthHistory() {
  const { serverHistory, serverServices, isInitialized } = useServiceHealth();

  // If not initialized, we are loading.
  // Once initialized, even if data is empty (no services), we are done loading.

  return {
    services: serverServices,
    history: serverHistory,
    isLoading: !isInitialized
  };
}
