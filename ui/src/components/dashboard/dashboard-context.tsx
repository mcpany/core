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

export function DashboardProvider({ children }: { children: ReactNode }) {
  const [serviceId, setServiceId] = useState<string | undefined>(undefined);

  return (
    <DashboardContext.Provider value={{ serviceId, setServiceId }}>
      {children}
    </DashboardContext.Provider>
  );
}

export function useDashboard() {
  const context = useContext(DashboardContext);
  if (!context) {
    throw new Error("useDashboard must be used within a DashboardProvider");
  }
  return context;
}
