/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import Link from "next/link";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Plus, ShoppingBag, BookOpen, CheckCircle2, Circle } from "lucide-react";

/**
 * OnboardingHero component.
 * Displays a welcome message and call-to-action for new users.
 * @returns The rendered component.
 */
export function OnboardingHero() {
  return (
    <div className="flex flex-col items-center justify-center min-h-[60vh] text-center space-y-8 p-4 animate-in fade-in slide-in-from-bottom-4 duration-500">
      <div className="space-y-4 max-w-2xl">
        <h1 className="text-4xl font-bold tracking-tight bg-gradient-to-r from-primary to-primary/60 bg-clip-text text-transparent">
          Welcome to MCP Any
        </h1>
        <p className="text-xl text-muted-foreground">
          Your unified control plane for the Model Context Protocol.
          <br />
          Connect your first service to unlock the power of AI tools.
        </p>
      </div>

      <div className="grid gap-6 md:grid-cols-2 max-w-4xl w-full">
        <Card className="hover:shadow-lg transition-all duration-300 border-primary/20 bg-gradient-to-br from-background to-muted/30">
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Plus className="h-5 w-5 text-primary" />
              Connect Service
            </CardTitle>
            <CardDescription>
              Manually configure an upstream service like a local MCP server or remote API.
            </CardDescription>
          </CardHeader>
          <CardContent>
            <Link href="/upstream-services">
              <Button size="lg" className="w-full">
                Connect Your First Service
              </Button>
            </Link>
          </CardContent>
        </Card>

        <Card className="hover:shadow-lg transition-all duration-300 border-muted bg-gradient-to-br from-background to-muted/10">
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <ShoppingBag className="h-5 w-5 text-purple-500" />
              Browse Marketplace
            </CardTitle>
            <CardDescription>
              Discover and install pre-configured services from the community or official catalog.
            </CardDescription>
          </CardHeader>
          <CardContent>
            <Link href="/marketplace">
              <Button variant="outline" size="lg" className="w-full">
                Explore Marketplace
              </Button>
            </Link>
          </CardContent>
        </Card>
      </div>

      <div className="pt-8 w-full max-w-md">
        <h3 className="text-sm font-medium text-muted-foreground mb-4 uppercase tracking-wider">Getting Started Checklist</h3>
        <div className="space-y-3 text-left bg-muted/20 p-6 rounded-lg border">
          <div className="flex items-center gap-3">
             <Circle className="h-5 w-5 text-muted-foreground" />
             <span className="font-medium">1. Connect an MCP Service</span>
          </div>
          <div className="flex items-center gap-3 text-muted-foreground opacity-60">
             <Circle className="h-5 w-5" />
             <span>2. Verify connection health</span>
          </div>
          <div className="flex items-center gap-3 text-muted-foreground opacity-60">
             <Circle className="h-5 w-5" />
             <span>3. Test tools in the Playground</span>
          </div>
        </div>
      </div>
    </div>
  );
}
