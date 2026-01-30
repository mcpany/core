/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useEffect } from "react";
import { DashboardGrid } from "@/components/dashboard/dashboard-grid";
import { OnboardingWizard } from "@/components/onboarding/onboarding-wizard";
import { apiClient } from "@/lib/client";
import { Loader2 } from "lucide-react";

/**
 * DashboardView component.
 * Orchestrates the dashboard view, showing the Onboarding Wizard for new users
 * (no services registered) or the standard Dashboard Grid for existing users.
 */
export function DashboardView() {
    const [isLoading, setIsLoading] = useState(true);
    const [hasServices, setHasServices] = useState(false);

    const checkServices = async () => {
        try {
            const services = await apiClient.listServices();
            if (services && services.length > 0) {
                setHasServices(true);
            } else {
                setHasServices(false);
            }
        } catch (e) {
            console.error("Failed to check services", e);
            // Default to dashboard grid on error to avoid blocking user
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
            <div className="flex h-full items-center justify-center">
                <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
            </div>
        );
    }

    if (!hasServices) {
        return <OnboardingWizard onComplete={() => setHasServices(true)} />;
    }

    return <DashboardGrid />;
}
