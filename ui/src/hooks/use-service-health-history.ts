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
  const { serverHistory, serverServices } = useServiceHealth();

  // Heuristic: If we have no data, we might be loading, or we might just have no services.
  // Since the context initializes with empty arrays, we can't distinguish easily without extra state.
  // However, for the widget, returning empty arrays with isLoading=false renders "No services", which is fine.
  // If we want to show a skeleton, we could check if latestTopology is null in the context, but we didn't expose that directly here.
  // For now, we assume if the provider is mounted, we have data or are fetching it.

  return {
    services: serverServices,
    history: serverHistory,
    isLoading: false
  };
}
