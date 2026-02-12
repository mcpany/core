/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useEffect } from "react";
import { DashboardGrid } from "@/components/dashboard/dashboard-grid";
import { OnboardingView } from "@/components/onboarding/onboarding-view";
import { apiClient } from "@/lib/client";
import { Loader2 } from "lucide-react";

export function DashboardContent() {
    const [isLoading, setIsLoading] = useState(true);
    const [hasServices, setHasServices] = useState(false);

    const checkServices = async () => {
        try {
            const data = await apiClient.listServices();
            // data is UpstreamServiceConfig[] or { services: ... }
            const list = Array.isArray(data) ? data : (data.services || []);
            setHasServices(list.length > 0);
        } catch (e) {
            console.error("Failed to list services", e);
            // In case of error, default to dashboard grid (maybe shows error there or empty)
            // But for onboarding purposes, if we can't fetch, we assume empty or error state.
            // Let's assume empty so user might try to register?
            // Or better, show dashboard which will show connection error.
            setHasServices(true);
        } finally {
            setIsLoading(false);
        }
    };

    useEffect(() => {
        checkServices();
    }, []);

    if (isLoading) {
        return (
            <div className="flex h-[50vh] items-center justify-center">
                <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
            </div>
        );
    }

    if (!hasServices) {
        return <OnboardingView onComplete={() => setHasServices(true)} />;
    }

    return <DashboardGrid />;
}
