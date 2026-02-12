/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useEffect, useCallback } from "react";
import { apiClient } from "@/lib/client";
import { DashboardGrid } from "./dashboard-grid";
import { OnboardingView } from "./onboarding-view";
import { Loader2 } from "lucide-react";

/**
 * DashboardShell component.
 * Manages the state of the dashboard, showing either the main grid or the onboarding view
 * based on whether any services are registered.
 */
export function DashboardShell() {
    const [isLoading, setIsLoading] = useState(true);
    const [hasServices, setHasServices] = useState(false);

    const checkServices = useCallback(async () => {
        try {
            const data = await apiClient.listServices();
            // Data might be an array or { services: [] } depending on API version/mock
            const list = Array.isArray(data) ? data : (data.services || []);
            setHasServices(list.length > 0);
        } catch (error) {
            console.error("Failed to check services status", error);
            // Fallback to dashboard if error, so user can debug
            setHasServices(true);
        } finally {
            setIsLoading(false);
        }
    }, []);

    useEffect(() => {
        checkServices();
    }, [checkServices]);

    if (isLoading) {
        return (
            <div className="flex h-full items-center justify-center min-h-[50vh]">
                <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
            </div>
        );
    }

    if (!hasServices) {
        return <OnboardingView onServiceRegistered={checkServices} />;
    }

    return <DashboardGrid />;
}
