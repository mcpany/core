/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import React, { createContext, useContext, useState, useEffect, ReactNode } from "react";

type Density = "comfortable" | "compact";

interface ViewPreferencesContextType {
  density: Density;
  setDensity: (density: Density) => void;
}

const ViewPreferencesContext = createContext<ViewPreferencesContextType | undefined>(undefined);

const STORAGE_KEY = "view-preferences";

/**
 * ViewPreferencesProvider.
 *
 * @param children - The children.
 */
export function ViewPreferencesProvider({ children }: { children: ReactNode }) {
  const [density, setDensityState] = useState<Density>("comfortable");

  useEffect(() => {
    const saved = localStorage.getItem(STORAGE_KEY);
    if (saved) {
      try {
        const parsed = JSON.parse(saved);
        if (parsed.density === "comfortable" || parsed.density === "compact") {
            setDensityState(parsed.density);
        }
      } catch (e) {
        console.error("Failed to parse view preferences", e);
      }
    }
  }, []);

  const setDensity = (newDensity: Density) => {
    setDensityState(newDensity);
    localStorage.setItem(STORAGE_KEY, JSON.stringify({ density: newDensity }));
  };

  // Prevent hydration mismatch by returning null or a consistent state until mounted
  // However, for a provider that just provides state, it's often better to render children
  // with default state and let it update client-side to avoid layout shift,
  // OR if the layout depends heavily on it, wait.
  // Given this affects density, a small layout shift is acceptable vs blocking render.

  return (
    <ViewPreferencesContext.Provider value={{ density, setDensity }}>
      {children}
    </ViewPreferencesContext.Provider>
  );
}

/**
 * Hook to access view preferences.
 *
 * @returns The view preferences context.
 */
export function useViewPreferences() {
  const context = useContext(ViewPreferencesContext);
  if (!context) {
    throw new Error("useViewPreferences must be used within a ViewPreferencesProvider");
  }
  return context;
}
