/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { useState, useEffect } from "react";

/**
 * ServiceStatus represents the possible health states of a service.
 */
export type ServiceStatus = "healthy" | "degraded" | "unhealthy" | "inactive";

/**
 * ServiceHealth describes the current health information of a service.
 */
export interface ServiceHealth {
  /** The unique identifier of the service. */
  id: string;
  /** The display name of the service. */
  name: string;
  /** The current status of the service. */
  status: ServiceStatus;
  /** The latency of the service check. */
  latency: string;
  /** The uptime duration of the service. */
  uptime: string;
  /** An optional message providing more details about the status. */
  message?: string;
}

/**
 * HealthHistoryPoint represents a single data point in the health history of a service.
 */
export interface HealthHistoryPoint {
  /** The timestamp of the health check in milliseconds. */
  timestamp: number;
  /** The status of the service at that time. */
  status: ServiceStatus;
  /** The latency of the check in ms. */
  latency?: number;
  /** Optional message. */
  message?: string;
}

/**
 * ServiceHistory maps service IDs to their list of historical health points.
 */
export interface ServiceHistory {
  [serviceId: string]: HealthHistoryPoint[];
}

/**
 * useServiceHealthHistory is a hook that fetches and maintains the health history of services.
 * It polls the backend API for health data history.
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
        // Parallel fetch for current status and history
        const [statusRes, historyRes] = await Promise.all([
             fetch("/api/dashboard/health"),
             fetch("/api/v1/services/health/history")
        ]);

        if (statusRes.ok) {
          const data = await statusRes.json();
          setServices(Array.isArray(data) ? data : []);
        }

        if (historyRes.ok) {
           const historyData = await historyRes.json();
           setHistory(historyData);
        }

      } catch (error) {
        console.warn("Failed to fetch health data", error);
      } finally {
        setIsLoading(false);
      }
    }

    fetchHealth();

    // Poll every 30 seconds (backend updates every 30s)
    const interval = setInterval(() => {
      if (!document.hidden) {
        fetchHealth();
      }
    }, 30000);

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
