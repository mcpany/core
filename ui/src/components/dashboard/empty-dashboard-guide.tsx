/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import Link from "next/link";
import { ArrowRight, Server, Zap, Activity, Plus } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from "@/components/ui/card";

export function EmptyDashboardGuide() {
  return (
    <div className="flex flex-col items-center justify-center min-h-[600px] p-8 animate-in fade-in zoom-in duration-500">
      <div className="max-w-3xl w-full space-y-8 text-center">
        <div className="space-y-4">
          <h1 className="text-4xl font-extrabold tracking-tight lg:text-5xl bg-clip-text text-transparent bg-gradient-to-r from-primary to-primary/60">
            Welcome to MCP Any
          </h1>
          <p className="text-xl text-muted-foreground max-w-2xl mx-auto">
            Your centralized control plane for AI context. Connect your APIs, databases, and tools to give your agents superpowers.
          </p>
        </div>

        <div className="grid gap-6 md:grid-cols-3 text-left">
          <Card className="relative overflow-hidden backdrop-blur-sm bg-background/50 border-primary/20 hover:border-primary/50 transition-colors group">
            <div className="absolute inset-0 bg-gradient-to-br from-blue-500/10 to-transparent opacity-0 group-hover:opacity-100 transition-opacity" />
            <CardHeader>
              <div className="w-12 h-12 rounded-lg bg-blue-500/10 flex items-center justify-center mb-4">
                <Server className="h-6 w-6 text-blue-500" />
              </div>
              <CardTitle>1. Connect a Service</CardTitle>
              <CardDescription>
                Register an upstream API, database, or command-line tool.
              </CardDescription>
            </CardHeader>
            <CardContent>
                <ul className="text-sm text-muted-foreground list-disc list-inside space-y-1">
                    <li>HTTP/REST APIs</li>
                    <li>PostgreSQL / MySQL</li>
                    <li>Local Scripts</li>
                </ul>
            </CardContent>
          </Card>

          <Card className="relative overflow-hidden backdrop-blur-sm bg-background/50 border-primary/20 hover:border-primary/50 transition-colors group">
            <div className="absolute inset-0 bg-gradient-to-br from-amber-500/10 to-transparent opacity-0 group-hover:opacity-100 transition-opacity" />
            <CardHeader>
              <div className="w-12 h-12 rounded-lg bg-amber-500/10 flex items-center justify-center mb-4">
                <Zap className="h-6 w-6 text-amber-500" />
              </div>
              <CardTitle>2. Explore Tools</CardTitle>
              <CardDescription>
                MCP Any automatically discovers and adapts endpoints into MCP Tools.
              </CardDescription>
            </CardHeader>
            <CardContent>
                <ul className="text-sm text-muted-foreground list-disc list-inside space-y-1">
                    <li>Auto-generated schemas</li>
                    <li>Security policies</li>
                    <li>Testing playground</li>
                </ul>
            </CardContent>
          </Card>

          <Card className="relative overflow-hidden backdrop-blur-sm bg-background/50 border-primary/20 hover:border-primary/50 transition-colors group">
            <div className="absolute inset-0 bg-gradient-to-br from-green-500/10 to-transparent opacity-0 group-hover:opacity-100 transition-opacity" />
            <CardHeader>
              <div className="w-12 h-12 rounded-lg bg-green-500/10 flex items-center justify-center mb-4">
                <Activity className="h-6 w-6 text-green-500" />
              </div>
              <CardTitle>3. Monitor Traffic</CardTitle>
              <CardDescription>
                Watch your agents interact with your infrastructure in real-time.
              </CardDescription>
            </CardHeader>
            <CardContent>
                <ul className="text-sm text-muted-foreground list-disc list-inside space-y-1">
                    <li>Request logging</li>
                    <li>Error tracking</li>
                    <li>Performance metrics</li>
                </ul>
            </CardContent>
          </Card>
        </div>

        <div className="pt-8">
          <Button size="lg" className="h-12 px-8 text-lg shadow-lg hover:shadow-xl transition-all hover:scale-105" asChild>
            <Link href="/upstream-services">
              <Plus className="mr-2 h-5 w-5" />
              Connect Your First Service
            </Link>
          </Button>
          <p className="mt-4 text-sm text-muted-foreground">
            Need help? <Link href="https://github.com/mcpany/core" target="_blank" className="underline hover:text-primary">Read the documentation</Link>.
          </p>
        </div>
      </div>
    </div>
  );
}
