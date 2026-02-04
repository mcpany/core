/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import React, { createContext, useContext, useState, ReactNode } from "react";

interface DashboardContextType {
  serviceId: string | undefined;
  setServiceId: (id: string | undefined) => void;
  timeRange: string;
  setTimeRange: (range: string) => void;
}

const DashboardContext = createContext<DashboardContextType | undefined>(undefined);

/**
 * Provides dashboard context to children, allowing them to access and update shared state
 * like the currently selected service filter.
 *
 * @param props - The component props.
 * @param props.children - Child components.
 * @returns The context provider.
 */
export function DashboardProvider({ children }: { children: ReactNode }) {
  const [serviceId, setServiceId] = useState<string | undefined>(undefined);
  const [timeRange, setTimeRange] = useState<string>("24h");

  return (
    <DashboardContext.Provider value={{ serviceId, setServiceId, timeRange, setTimeRange }}>
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
