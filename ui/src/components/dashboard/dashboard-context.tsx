/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import React, { createContext, useContext, useState, ReactNode } from "react";

interface DashboardContextType {
  serviceId: string | undefined;
  setServiceId: (id: string | undefined) => void;
}

const DashboardContext = createContext<DashboardContextType | undefined>(undefined);

/**
 * DashboardProvider component.
 * @param props - The component props.
 * @param props.children - The child components.
 * @returns The rendered component.
 */
export function DashboardProvider({ children }: { children: ReactNode }) {
  const [serviceId, setServiceId] = useState<string | undefined>(undefined);

  return (
    <DashboardContext.Provider value={{ serviceId, setServiceId }}>
      {children}
    </DashboardContext.Provider>
  );
}

/**
 * Hook to access the dashboard context.
 *
 * @returns The dashboard context value, including the current service ID filter.
 * @throws Error if used outside of a DashboardProvider.
 */
export function useDashboard() {
  const context = useContext(DashboardContext);
  if (!context) {
    throw new Error("useDashboard must be used within a DashboardProvider");
  }
  return context;
}
