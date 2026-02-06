/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useEffect, useState } from "react";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { apiClient } from "@/lib/client";
import { Plus, BookOpen, ExternalLink, X } from "lucide-react";
import Link from "next/link";
import { useDashboard } from "@/components/dashboard/dashboard-context";

/**
 * GettingStartedWidget.
 * Shows onboarding steps if no services are connected.
 */
export function GettingStartedWidget() {
    const [hasServices, setHasServices] = useState<boolean | null>(null);
    const [dismissed, setDismissed] = useState(false);
    const { serviceId } = useDashboard();

    useEffect(() => {
        async function checkServices() {
            try {
                const services = await apiClient.listServices();
                setHasServices(services && services.length > 0);
            } catch (e) {
                console.error("Failed to check services", e);
                // Assume no services on error to show help
                setHasServices(false);
            }
        }
        checkServices();
    }, [serviceId]);

    // Don't render if loading (null), has services (true), or user dismissed
    if (dismissed || hasServices === true || hasServices === null) {
        return null;
    }

    return (
        <Card className="col-span-12 lg:col-span-12 bg-gradient-to-r from-primary/10 to-background border-primary/20 relative overflow-hidden">
            <div className="absolute top-0 right-0 p-4">
                 <Button variant="ghost" size="icon" className="h-6 w-6 opacity-50 hover:opacity-100" onClick={() => setDismissed(true)}>
                    <X className="h-4 w-4" />
                 </Button>
            </div>
            <CardHeader>
                <CardTitle className="text-2xl">Welcome to MCP Any</CardTitle>
                <CardDescription className="text-base max-w-2xl">
                    Your universal adapter for the Model Context Protocol. Connect your APIs, databases, and services to AI agents instantly.
                </CardDescription>
            </CardHeader>
            <CardContent className="flex flex-wrap gap-4">
                <Button asChild size="lg" className="shadow-md">
                    <Link href="/upstream-services">
                        <Plus className="mr-2 h-4 w-4" />
                        Connect First Service
                    </Link>
                </Button>
                <Button variant="outline" size="lg" asChild>
                    <a href="https://github.com/mcpany/core" target="_blank" rel="noopener noreferrer">
                        <BookOpen className="mr-2 h-4 w-4" />
                        Read Documentation
                    </a>
                </Button>
                <Button variant="ghost" size="lg" asChild>
                    <a href="https://modelcontextprotocol.io" target="_blank" rel="noopener noreferrer">
                        <ExternalLink className="mr-2 h-4 w-4" />
                        Learn about MCP
                    </a>
                </Button>
            </CardContent>
        </Card>
    );
}
