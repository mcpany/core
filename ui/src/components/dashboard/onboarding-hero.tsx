/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import Link from "next/link";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Rocket, Zap, BookOpen, ExternalLink, CheckCircle2, Server, Package } from "lucide-react";
import { Badge } from "@/components/ui/badge";

/**
 * OnboardingHero component.
 * Displays a welcoming "Get Started" hero section for users with no connected services.
 * @returns The rendered component.
 */
export function OnboardingHero() {
  return (
    <div className="flex flex-col items-center justify-center min-h-[60vh] text-center space-y-8 animate-in fade-in duration-700 slide-in-from-bottom-4">
      <div className="space-y-4 max-w-2xl">
        <div className="inline-flex items-center justify-center p-3 mb-4 rounded-full bg-primary/10 text-primary">
          <Rocket className="h-8 w-8" />
        </div>
        <h1 className="text-4xl font-extrabold tracking-tight lg:text-5xl bg-clip-text text-transparent bg-gradient-to-r from-primary to-primary/60">
          Welcome to MCP Any
        </h1>
        <p className="text-xl text-muted-foreground leading-relaxed">
          Your unified control plane for the Model Context Protocol.
          <br />
          Connect your APIs, manage tools, and empower your AI agents in minutes.
        </p>
      </div>

      <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3 max-w-5xl w-full px-4">
        {/* Step 1: Connect Service */}
        <Card className="relative overflow-hidden border-primary/20 bg-gradient-to-b from-background to-muted/20 hover:shadow-lg transition-all duration-300 group">
          <div className="absolute top-0 left-0 w-1 h-full bg-primary" />
          <CardHeader>
            <div className="flex items-center justify-between">
                <Badge variant="secondary" className="mb-2">Step 1</Badge>
                <Server className="h-5 w-5 text-muted-foreground group-hover:text-primary transition-colors" />
            </div>
            <CardTitle>Connect a Service</CardTitle>
            <CardDescription>
              Register your first upstream API or MCP server.
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
             <div className="text-sm text-muted-foreground text-left">
                Start by adding an existing API (OpenAPI, REST) or a local command.
             </div>
             <Link href="/upstream-services" className="block">
                <Button className="w-full group-hover:bg-primary group-hover:text-primary-foreground transition-colors">
                    Connect Service <Zap className="ml-2 h-4 w-4" />
                </Button>
             </Link>
          </CardContent>
        </Card>

        {/* Step 2: Browse Marketplace */}
        <Card className="relative overflow-hidden hover:shadow-lg transition-all duration-300 group">
           <CardHeader>
            <div className="flex items-center justify-between">
                <Badge variant="outline" className="mb-2">Step 2</Badge>
                <Package className="h-5 w-5 text-muted-foreground group-hover:text-primary transition-colors" />
            </div>
            <CardTitle>Explore Marketplace</CardTitle>
            <CardDescription>
              Discover pre-built integrations and community servers.
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
             <div className="text-sm text-muted-foreground text-left">
                Install popular integrations like GitHub, Google Drive, or Slack.
             </div>
             <Link href="/marketplace" className="block">
                <Button variant="outline" className="w-full">
                    Browse Marketplace
                </Button>
             </Link>
          </CardContent>
        </Card>

        {/* Step 3: Learn More */}
        <Card className="relative overflow-hidden hover:shadow-lg transition-all duration-300 group">
           <CardHeader>
            <div className="flex items-center justify-between">
                <Badge variant="outline" className="mb-2">Resources</Badge>
                <BookOpen className="h-5 w-5 text-muted-foreground group-hover:text-primary transition-colors" />
            </div>
            <CardTitle>Documentation</CardTitle>
            <CardDescription>
              Learn how to build and deploy MCP servers.
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
             <div className="text-sm text-muted-foreground text-left">
                Read the guides to master MCP Any configuration and deployment.
             </div>
             <a href="https://github.com/mcpany/core" target="_blank" rel="noopener noreferrer" className="block">
                <Button variant="ghost" className="w-full border-dashed border hover:border-solid">
                    View Documentation <ExternalLink className="ml-2 h-4 w-4" />
                </Button>
             </a>
          </CardContent>
        </Card>
      </div>

      <div className="flex items-center gap-2 text-sm text-muted-foreground opacity-60 mt-8">
        <CheckCircle2 className="h-4 w-4" /> Secure by Default
        <span className="mx-2">•</span>
        <CheckCircle2 className="h-4 w-4" /> Open Source
        <span className="mx-2">•</span>
        <CheckCircle2 className="h-4 w-4" /> Enterprise Grade
      </div>
    </div>
  );
}
