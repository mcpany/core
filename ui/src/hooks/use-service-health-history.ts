/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { useState, useEffect } from "react";

/**
 * ServiceStatus represents the possible health states of a service.
 */
export type ServiceStatus = "healthy" | "degraded" | "unhealthy" | "inactive" | "unknown";

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
}

/**
 * ServiceHistory maps service IDs to their list of historical health points.
 */
export interface ServiceHistory {
  [serviceId: string]: HealthHistoryPoint[];
}

interface HealthResponse {
  services: ServiceHealth[];
  history: ServiceHistory;
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
        const res = await fetch("/api/v1/dashboard/health");
        if (res.ok) {
          const data: HealthResponse = await res.json();

          // Backend returns history keyed by ID
          setServices(data.services || []);
          setHistory(data.history || {});
        }
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
