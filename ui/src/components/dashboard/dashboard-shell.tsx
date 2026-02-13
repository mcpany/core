/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useEffect, useCallback } from "react";
import { apiClient } from "@/lib/client";
import { DashboardGrid } from "./dashboard-grid";
import { OnboardingView } from "./onboarding-view";
import { Loader2, AlertCircle, RefreshCw } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";

/**
 * DashboardShell component.
 * Manages the state of the dashboard, showing either the main grid or the onboarding view
 * based on whether any services are registered.
 */
export function DashboardShell() {
    const [isLoading, setIsLoading] = useState(true);
    const [hasServices, setHasServices] = useState(false);
    const [error, setError] = useState<string | null>(null);

    const checkServices = useCallback(async () => {
        setIsLoading(true);
        setError(null);
        try {
            const data = await apiClient.listServices();
            // Data might be an array or { services: [] } depending on API version/mock
            const list = Array.isArray(data) ? data : (data.services || []);
            setHasServices(list.length > 0);
        } catch (error: any) {
            console.error("Failed to check services status", error);
            setError(error.message || "Failed to load services.");
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

    if (error) {
        return (
            <div className="flex h-full items-center justify-center min-h-[50vh] p-8">
                <Alert variant="destructive" className="max-w-md">
                    <AlertCircle className="h-4 w-4" />
                    <AlertTitle>Connection Error</AlertTitle>
                    <AlertDescription className="mt-2 flex flex-col gap-4">
                        <p>{error}</p>
                        <Button onClick={checkServices} variant="outline" className="w-full">
                            <RefreshCw className="mr-2 h-4 w-4" /> Retry
                        </Button>
                    </AlertDescription>
                </Alert>
            </div>
        );
    }

    if (!hasServices) {
        return <OnboardingView onServiceRegistered={checkServices} />;
    }

    return <DashboardGrid />;
}
