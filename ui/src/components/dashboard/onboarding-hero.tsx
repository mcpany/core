/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import Link from "next/link";
import { Server, ShoppingBag, Book, Activity } from "lucide-react";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";

interface OnboardingHeroProps {
    onAddWidget: (type: string) => void;
}

/**
 * OnboardingHero component.
 * Displays a welcome message and quick actions for new users.
 *
 * @param props - The component props.
 * @param props.onAddWidget - Callback to add a widget.
 * @returns The rendered component.
 */
export function OnboardingHero({ onAddWidget }: OnboardingHeroProps) {
    return (
        <div className="flex flex-col gap-8 py-10 animate-in fade-in slide-in-from-bottom-4 duration-700">
            <div className="text-center space-y-4">
                <h1 className="text-4xl font-bold tracking-tight bg-gradient-to-r from-foreground to-foreground/70 bg-clip-text text-transparent">
                    Welcome to MCP Any
                </h1>
                <p className="text-lg text-muted-foreground max-w-2xl mx-auto">
                    Your enterprise-grade management console for the Model Context Protocol.
                    Connect services, manage tools, and monitor your agent infrastructure.
                </p>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-3 gap-6 max-w-5xl mx-auto w-full px-4">
                <Link href="/upstream-services" className="group">
                    <Card className="h-full border-muted/60 bg-card/50 backdrop-blur-sm transition-all hover:bg-card hover:shadow-lg hover:-translate-y-1">
                        <CardHeader>
                            <div className="p-3 w-fit rounded-lg bg-blue-500/10 text-blue-500 mb-2 group-hover:bg-blue-500 group-hover:text-white transition-colors">
                                <Server className="h-6 w-6" />
                            </div>
                            <CardTitle>Connect a Service</CardTitle>
                            <CardDescription>
                                Register a new upstream service via HTTP, gRPC, or Command line.
                            </CardDescription>
                        </CardHeader>
                    </Card>
                </Link>

                <Link href="/marketplace" className="group">
                    <Card className="h-full border-muted/60 bg-card/50 backdrop-blur-sm transition-all hover:bg-card hover:shadow-lg hover:-translate-y-1">
                        <CardHeader>
                            <div className="p-3 w-fit rounded-lg bg-purple-500/10 text-purple-500 mb-2 group-hover:bg-purple-500 group-hover:text-white transition-colors">
                                <ShoppingBag className="h-6 w-6" />
                            </div>
                            <CardTitle>Browse Marketplace</CardTitle>
                            <CardDescription>
                                Discover and install pre-configured server templates and community tools.
                            </CardDescription>
                        </CardHeader>
                    </Card>
                </Link>

                <a href="https://modelcontextprotocol.io/introduction" target="_blank" rel="noopener noreferrer" className="group">
                    <Card className="h-full border-muted/60 bg-card/50 backdrop-blur-sm transition-all hover:bg-card hover:shadow-lg hover:-translate-y-1">
                        <CardHeader>
                            <div className="p-3 w-fit rounded-lg bg-emerald-500/10 text-emerald-500 mb-2 group-hover:bg-emerald-500 group-hover:text-white transition-colors">
                                <Book className="h-6 w-6" />
                            </div>
                            <CardTitle>Documentation</CardTitle>
                            <CardDescription>
                                Learn how to build MCP servers and integrate them with your agents.
                            </CardDescription>
                        </CardHeader>
                    </Card>
                </a>
            </div>

            <div className="flex flex-col items-center gap-4 mt-8">
                <div className="flex items-center gap-4 text-sm text-muted-foreground">
                    <div className="h-px w-12 bg-border" />
                    <span>Or start building your dashboard</span>
                    <div className="h-px w-12 bg-border" />
                </div>
                <Button
                    variant="outline"
                    size="lg"
                    className="gap-2 border-dashed"
                    onClick={() => onAddWidget("metrics")}
                >
                    <Activity className="h-4 w-4" />
                    Add Metrics Overview
                </Button>
            </div>
        </div>
    );
}
