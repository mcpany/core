/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import React, { createContext, useContext, useEffect, useState } from "react";

export type DashboardDensity = "comfortable" | "compact";

interface DashboardDensityContextType {
    density: DashboardDensity;
    setDensity: (density: DashboardDensity) => void;
}

const DashboardDensityContext = createContext<DashboardDensityContextType | undefined>(undefined);

export function DashboardDensityProvider({ children }: { children: React.ReactNode }) {
    const [density, setDensity] = useState<DashboardDensity>("comfortable");
    const [isMounted, setIsMounted] = useState(false);

    useEffect(() => {
        setIsMounted(true);
        const saved = localStorage.getItem("dashboard-density");
        if (saved === "compact" || saved === "comfortable") {
            setDensity(saved);
        }
    }, []);

    const updateDensity = (newDensity: DashboardDensity) => {
        setDensity(newDensity);
        localStorage.setItem("dashboard-density", newDensity);
    };

    if (!isMounted) {
        return null; // or a loading spinner
    }

    return (
        <DashboardDensityContext.Provider value={{ density, setDensity: updateDensity }}>
            {children}
        </DashboardDensityContext.Provider>
    );
}

export function useDashboardDensity() {
    const context = useContext(DashboardDensityContext);
    if (context === undefined) {
        throw new Error("useDashboardDensity must be used within a DashboardDensityProvider");
    }
    return context;
}
