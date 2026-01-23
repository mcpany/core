/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { useState, useEffect, useRef } from "react";

export type ServiceStatus = "healthy" | "degraded" | "unhealthy" | "inactive";

export interface ServiceHealth {
  id: string;
  name: string;
  status: ServiceStatus;
  latency: string;
  uptime: string;
  message?: string;
}

export interface HealthHistoryPoint {
  timestamp: number;
  status: ServiceStatus;
}

export interface ServiceHistory {
  [serviceId: string]: HealthHistoryPoint[];
}

const HISTORY_KEY = "mcp_service_health_history";
const MAX_HISTORY_POINTS = 60; // Keep last 60 checks (e.g. 1 hour if polling every minute, or 10 mins if polling every 10s)
// Actually, if we poll every 10s, 60 points is 10 minutes.
// If we want 24h, we need 24 * 60 = 1440 points (if 1 min interval).
// Storing 1440 * N services might be fine in localStorage (small ints).
// Let's stick to 100 points for now (visual density).

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
