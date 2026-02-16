/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { useState, useEffect } from "react";
import { apiClient, ServiceHealth, HealthHistoryPoint } from "@/lib/client";

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

  useEffect(() => {
    async function fetchHealth() {
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
    }

    fetchHealth();

    // Poll every 10 seconds
    const interval = setInterval(() => {
      if (!document.hidden) {
        fetchHealth();
      }
    }, 10000);

    const onVisibilityChange = () => {
      if (!document.hidden) {
        fetchHealth();
      }
    };
    document.addEventListener("visibilitychange", onVisibilityChange);

    return () => {
      clearInterval(interval);
      document.removeEventListener("visibilitychange", onVisibilityChange);
    };
  }, []);

  return { services, history, isLoading };
}
