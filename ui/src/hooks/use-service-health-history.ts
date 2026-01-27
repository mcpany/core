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
}

/**
 * ServiceHistory maps service IDs to their list of historical health points.
 */
export interface ServiceHistory {
  [serviceId: string]: HealthHistoryPoint[];
}

const HISTORY_KEY = "mcp_service_health_history";
const MAX_HISTORY_POINTS = 8640; // Keep last 24 hours of checks (polling every 10s: 6 * 60 * 24 = 8640)

/**
 * useServiceHealthHistory is a hook that fetches and maintains the health history of services.
 * It polls the backend API for health data and persists the history in local storage.
 *
 * @returns An object containing the current services list, their health history, and a loading state.
 */
export function useServiceHealthHistory() {
  const [history, setHistory] = useState<ServiceHistory>({});
  const [services, setServices] = useState<ServiceHealth[]>([]);
  const [isLoading, setIsLoading] = useState(true);

  // Load initial history
  useEffect(() => {
    if (typeof window !== "undefined") {
      try {
        const stored = window.localStorage.getItem(HISTORY_KEY);
        if (stored) {
          setHistory(JSON.parse(stored));
        }
      } catch (e) {
        console.error("Failed to load health history", e);
      }
    }
  }, []);

  useEffect(() => {
    async function fetchHealth() {
      try {
        const res = await fetch("/api/dashboard/health");
        if (res.ok) {
          const data = await res.json();
          const currentServices: ServiceHealth[] = Array.isArray(data) ? data : [];

          setServices(currentServices);

          // Update history
          setHistory(prev => {
            const next = { ...prev };
            const now = Date.now();

            currentServices.forEach(svc => {
              const point: HealthHistoryPoint = {
                timestamp: now,
                status: svc.status
              };

              const svcHistory = next[svc.id] || [];
              // Add new point
              const newHistory = [...svcHistory, point];

              // Sort by time just in case
              newHistory.sort((a, b) => a.timestamp - b.timestamp);

              // Limit size
              if (newHistory.length > MAX_HISTORY_POINTS) {
                 newHistory.splice(0, newHistory.length - MAX_HISTORY_POINTS);
              }

              next[svc.id] = newHistory;
            });

            // Persist
            if (typeof window !== "undefined") {
               window.localStorage.setItem(HISTORY_KEY, JSON.stringify(next));
            }

            return next;
          });
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
