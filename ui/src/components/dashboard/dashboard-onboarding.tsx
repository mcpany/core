/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import Link from "next/link";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { BookOpen, Package, Plus, Sparkles, Download } from "lucide-react";
import { AddWidgetSheet } from "./add-widget-sheet";

interface DashboardOnboardingProps {
    onAddWidget: (type: string) => void;
}

/**
 * DashboardOnboarding component.
 * displayed when the dashboard is empty to guide the user.
 *
 * @param props - The component props.
 * @returns The rendered component.
 */
export function DashboardOnboarding({ onAddWidget }: DashboardOnboardingProps) {
    return (
        <div className="flex flex-col items-center justify-center min-h-[60vh] space-y-8 animate-in fade-in zoom-in duration-500">
            <div className="text-center space-y-4 max-w-2xl">
                <div className="inline-flex items-center justify-center p-4 bg-primary/10 rounded-full mb-4">
                    <Sparkles className="h-12 w-12 text-primary" />
                </div>
                <h1 className="text-4xl font-bold tracking-tight bg-clip-text text-transparent bg-gradient-to-r from-foreground to-foreground/70">
                    Welcome to MCP Any
                </h1>
                <p className="text-xl text-muted-foreground">
                    Your enterprise-grade gateway for the Model Context Protocol.
                    <br />
                    Connect your first service or customize your dashboard to get started.
                </p>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 w-full max-w-5xl px-4">
                {/* Action 1: Connect Popular Service */}
                <Link href="/marketplace" className="group">
                    <Card className="h-full hover:shadow-lg transition-all duration-300 hover:border-primary/50 cursor-pointer bg-card/50 backdrop-blur-sm">
                        <CardHeader>
                            <CardTitle className="flex items-center gap-2">
                                <Package className="h-5 w-5 text-blue-500 group-hover:scale-110 transition-transform" />
                                Marketplace
                            </CardTitle>
                            <CardDescription>Connect popular services like OpenAI, SQLite, or Filesystem.</CardDescription>
                        </CardHeader>
                        <CardContent>
                            <Button variant="secondary" className="w-full group-hover:bg-primary group-hover:text-primary-foreground">
                                Browse Services
                            </Button>
                        </CardContent>
                    </Card>
                </Link>

                {/* Action 2: Import OpenAPI */}
                <Link href="/marketplace?wizard=open" className="group">
                    <Card className="h-full hover:shadow-lg transition-all duration-300 hover:border-primary/50 cursor-pointer bg-card/50 backdrop-blur-sm">
                        <CardHeader>
                            <CardTitle className="flex items-center gap-2">
                                <Download className="h-5 w-5 text-green-500 group-hover:scale-110 transition-transform" />
                                Import OpenAPI
                            </CardTitle>
                            <CardDescription>Turn any REST API with a Swagger spec into an MCP server.</CardDescription>
                        </CardHeader>
                        <CardContent>
                            <Button variant="secondary" className="w-full group-hover:bg-primary group-hover:text-primary-foreground">
                                Start Wizard
                            </Button>
                        </CardContent>
                    </Card>
                </Link>

                {/* Action 3: Add Widget (Custom Trigger) */}
                <AddWidgetSheet onAdd={onAddWidget}>
                    <Card className="h-full hover:shadow-lg transition-all duration-300 hover:border-primary/50 cursor-pointer bg-card/50 backdrop-blur-sm group text-left">
                        <CardHeader>
                            <CardTitle className="flex items-center gap-2">
                                <Plus className="h-5 w-5 text-amber-500 group-hover:scale-110 transition-transform" />
                                Customize View
                            </CardTitle>
                            <CardDescription>Add widgets to your dashboard to monitor health and metrics.</CardDescription>
                        </CardHeader>
                        <CardContent>
                            <Button variant="secondary" className="w-full group-hover:bg-primary group-hover:text-primary-foreground">
                                Add Widget
                            </Button>
                        </CardContent>
                    </Card>
                </AddWidgetSheet>

                {/* Action 4: Documentation */}
                <a href="https://github.com/mcpany/core" target="_blank" rel="noopener noreferrer" className="group">
                    <Card className="h-full hover:shadow-lg transition-all duration-300 hover:border-primary/50 cursor-pointer bg-card/50 backdrop-blur-sm">
                        <CardHeader>
                            <CardTitle className="flex items-center gap-2">
                                <BookOpen className="h-5 w-5 text-purple-500 group-hover:scale-110 transition-transform" />
                                Documentation
                            </CardTitle>
                            <CardDescription>Learn how to configure, secure, and extend your MCP gateway.</CardDescription>
                        </CardHeader>
                        <CardContent>
                            <Button variant="secondary" className="w-full group-hover:bg-primary group-hover:text-primary-foreground">
                                Read Docs
                            </Button>
                        </CardContent>
                    </Card>
                </a>
            </div>
        </div>
    );
}
