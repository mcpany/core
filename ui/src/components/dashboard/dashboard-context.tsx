/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import React, { createContext, useContext, useState, useEffect, ReactNode } from "react";

export type Density = "comfortable" | "compact";

interface DashboardContextType {
  serviceId: string | undefined;
  setServiceId: (id: string | undefined) => void;
  density: Density;
  setDensity: (density: Density) => void;
}

const DashboardContext = createContext<DashboardContextType | undefined>(undefined);

/**
 * Provides dashboard context to children, allowing them to access and update shared state
 * like the currently selected service filter and density preference.
 *
 * @param props - The component props.
 * @param props.children - Child components.
 * @returns The context provider.
 */
export function DashboardProvider({ children }: { children: ReactNode }) {
  const [serviceId, setServiceId] = useState<string | undefined>(undefined);
  const [density, setDensity] = useState<Density>("comfortable");
  const [isMounted, setIsMounted] = useState(false);

  // Load persistence
  useEffect(() => {
    setIsMounted(true);
    const saved = localStorage.getItem("dashboard-density");
    if (saved === "compact" || saved === "comfortable") {
        setDensity(saved);
    }
  }, []);

  // Save persistence
  useEffect(() => {
    if (!isMounted) return;
    localStorage.setItem("dashboard-density", density);
  }, [density, isMounted]);

  return (
    <DashboardContext.Provider value={{ serviceId, setServiceId, density, setDensity }}>
      {children}
    </DashboardContext.Provider>
  );
}

/**
 * Hook to access the dashboard context.
 * Must be used within a DashboardProvider.
 *
 * @returns The dashboard context value.
 * @throws Error if used outside of a DashboardProvider.
 */
export function useDashboard() {
  const context = useContext(DashboardContext);
  if (!context) {
    throw new Error("useDashboard must be used within a DashboardProvider");
  }
  return context;
}
