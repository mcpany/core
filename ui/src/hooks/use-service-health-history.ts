/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { ServiceHealth, HealthHistoryPoint } from "@/lib/client";
import { useServiceHealth } from "@/contexts/service-health-context";

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
 * ⚡ Bolt Optimization: Now consumes data from ServiceHealthContext to avoid redundant polling.
 * Randomized Selection from Top 5 High-Impact Targets.
 *
 * @returns An object containing the current services list, their health history, and a loading state.
 */
export function useServiceHealthHistory() {
  const { services, serverHistory, latestTopology } = useServiceHealth();

  // If we have no services and no topology yet, we are probably loading
  const isLoading = services.length === 0 && !latestTopology;

  return {
    services,
    history: serverHistory,
    isLoading
  };
}
