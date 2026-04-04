/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */



import React, { createContext, useContext, useState, ReactNode } from "react";

interface DashboardContextType {
  serviceId: string | undefined;
  setServiceId: (id: string | undefined) => void;
  timeRange: string;
  setTimeRange: (range: string) => void;
}

const DashboardContext = createContext<DashboardContextType | undefined>(undefined);

/**
 * DashboardProvider executes the DashboardProvider logic.
 *
 * Summary: Executes the DashboardProvider logic.
 *
 * @param { children } - The { children } parameter.
 * @returns The result of the operation.
 * @throws An error if the operation fails.
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
 * Intent: Document useDashboard
 *
 * Params:
 *   - None
 *
 * Returns:
 *   - Documented below.
 *
 * Errors:
 *   - Documented below.
 *
 * Side Effects:
 *   - None
 *
 * Hook to access the dashboard context.
 * Must be used within a DashboardProvider.
 *
 * @returns The dashboard context value.
 * @throws Error if used outside of a DashboardProvider.
 */
/**
 * useDashboard executes the useDashboard logic.
 *
 * Summary: Executes the useDashboard logic.
 *
 * @param None.
 * @returns The result of the operation.
 * @throws An error if the operation fails.
 */
export function useDashboard() {
  const context = useContext(DashboardContext);
  if (!context) {
    throw new Error("useDashboard must be used within a DashboardProvider");
  }
  return context;
}
