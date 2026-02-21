/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import Link from "next/link";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle, CardFooter } from "@/components/ui/card";
import { Plus, ShoppingBag, BookOpen, CheckCircle2, Circle } from "lucide-react";

/**
 * OnboardingHero component.
 * Displays a welcoming hero section for new users with no services connected.
 * @returns The rendered component.
 */
export function OnboardingHero() {
  return (
    <div className="flex items-center justify-center min-h-[60vh] p-4">
      <Card className="w-full max-w-3xl backdrop-blur-xl bg-background/60 border-white/20 shadow-2xl relative overflow-hidden">
        {/* Background Decoration */}
        <div className="absolute -top-24 -right-24 w-64 h-64 bg-primary/10 rounded-full blur-3xl pointer-events-none" />
        <div className="absolute -bottom-24 -left-24 w-64 h-64 bg-blue-500/10 rounded-full blur-3xl pointer-events-none" />

        <CardHeader className="text-center pb-8 pt-10 relative z-10">
          <div className="mx-auto bg-primary/10 p-4 rounded-full w-16 h-16 flex items-center justify-center mb-4">
             <div className="w-8 h-8 bg-primary rounded-md" />
          </div>
          <CardTitle className="text-4xl font-bold tracking-tight bg-gradient-to-br from-foreground to-muted-foreground bg-clip-text text-transparent">
            Welcome to MCP Any
          </CardTitle>
          <CardDescription className="text-lg mt-2 max-w-lg mx-auto">
            Your unified control plane for Model Context Protocol. Connect your first service to get started.
          </CardDescription>
        </CardHeader>

        <CardContent className="grid md:grid-cols-2 gap-8 px-12 relative z-10">
            <div className="space-y-4">
                <h3 className="font-semibold text-sm uppercase tracking-wider text-muted-foreground">Quick Start</h3>
                <div className="space-y-3">
                    <Link href="/upstream-services" className="block group">
                        <div className="p-4 rounded-lg border bg-card hover:border-primary/50 hover:shadow-md transition-all flex items-center gap-4">
                            <div className="p-2 bg-primary/10 rounded-md text-primary">
                                <Plus className="w-6 h-6" />
                            </div>
                            <div>
                                <div className="font-semibold group-hover:text-primary transition-colors">Connect Service</div>
                                <div className="text-xs text-muted-foreground">Add a local or remote MCP server</div>
                            </div>
                        </div>
                    </Link>
                    <Link href="/marketplace" className="block group">
                        <div className="p-4 rounded-lg border bg-card hover:border-primary/50 hover:shadow-md transition-all flex items-center gap-4">
                            <div className="p-2 bg-secondary/80 rounded-md text-secondary-foreground">
                                <ShoppingBag className="w-6 h-6" />
                            </div>
                            <div>
                                <div className="font-semibold group-hover:text-primary transition-colors">Browse Marketplace</div>
                                <div className="text-xs text-muted-foreground">Discover community servers</div>
                            </div>
                        </div>
                    </Link>
                </div>
            </div>

            <div className="space-y-4">
                <h3 className="font-semibold text-sm uppercase tracking-wider text-muted-foreground">Your Journey</h3>
                <div className="space-y-4 pt-2">
                    <div className="flex items-start gap-3">
                        <Circle className="w-5 h-5 text-muted-foreground mt-0.5" />
                        <div>
                            <div className="text-sm font-medium">1. Connect a Service</div>
                            <div className="text-xs text-muted-foreground">Register an upstream MCP server via HTTP, gRPC, or Command.</div>
                        </div>
                    </div>
                    <div className="flex items-start gap-3 opacity-50">
                        <Circle className="w-5 h-5 text-muted-foreground mt-0.5" />
                        <div>
                            <div className="text-sm font-medium">2. Explore Tools</div>
                            <div className="text-xs text-muted-foreground">View and test exposed capabilities in the Playground.</div>
                        </div>
                    </div>
                    <div className="flex items-start gap-3 opacity-50">
                        <Circle className="w-5 h-5 text-muted-foreground mt-0.5" />
                        <div>
                            <div className="text-sm font-medium">3. Monitor Traffic</div>
                            <div className="text-xs text-muted-foreground">See live metrics and traces on this dashboard.</div>
                        </div>
                    </div>
                </div>
            </div>
        </CardContent>

        <CardFooter className="flex justify-center pb-8 relative z-10">
            <Button variant="ghost" size="sm" asChild>
                <a href="https://github.com/mcpany/core" target="_blank" rel="noopener noreferrer" className="text-muted-foreground hover:text-foreground">
                    <BookOpen className="mr-2 h-4 w-4" /> Read Documentation
                </a>
            </Button>
        </CardFooter>
      </Card>
    </div>
  );
}
