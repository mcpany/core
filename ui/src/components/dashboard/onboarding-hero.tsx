/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import Link from "next/link";
import { ArrowRight, Server, Terminal, Wrench, Sparkles, LayoutGrid } from "lucide-react";
import { ConnectClientButton } from "@/components/connect-client-button";

/**
 * OnboardingHero component.
 * displayed when the dashboard is empty (no services configured).
 * Guides the user through the initial setup steps.
 */
export function OnboardingHero() {
  return (
    <div className="flex flex-col items-center justify-center min-h-[60vh] space-y-8 p-4 animate-in fade-in zoom-in duration-500">
      <div className="text-center space-y-4 max-w-2xl">
        <div className="flex justify-center mb-6">
            <div className="p-4 bg-primary/10 rounded-2xl ring-1 ring-primary/20 shadow-xl backdrop-blur-sm">
                <Sparkles className="h-12 w-12 text-primary" />
            </div>
        </div>
        <h1 className="text-4xl font-bold tracking-tight bg-gradient-to-br from-foreground to-muted-foreground bg-clip-text text-transparent">
          Welcome to MCP Any
        </h1>
        <p className="text-xl text-muted-foreground leading-relaxed">
          Your universal gateway for Model Context Protocol. <br/>
          Connect your existing APIs and tools to AI agents in minutes.
        </p>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-3 gap-6 w-full max-w-5xl px-4">
        {/* Step 1: Connect Service */}
        <Card className="relative overflow-hidden group hover:shadow-lg transition-all border-dashed hover:border-solid hover:border-primary/50">
          <div className="absolute top-0 right-0 p-4 opacity-10 group-hover:opacity-20 transition-opacity">
            <Server className="h-24 w-24 -mr-8 -mt-8" />
          </div>
          <CardHeader>
            <div className="h-10 w-10 rounded-lg bg-blue-500/10 flex items-center justify-center mb-2 text-blue-500">
                <span className="font-bold font-mono">1</span>
            </div>
            <CardTitle>Connect Service</CardTitle>
            <CardDescription>
              Add your first upstream service or use a template.
            </CardDescription>
          </CardHeader>
          <CardContent>
            <Link href="/marketplace?wizard=open">
                <Button className="w-full gap-2 group/btn">
                    Open Wizard <ArrowRight className="h-4 w-4 group-hover/btn:translate-x-1 transition-transform" />
                </Button>
            </Link>
          </CardContent>
        </Card>

        {/* Step 2: Configure Client */}
        <Card className="relative overflow-hidden group hover:shadow-lg transition-all border-dashed hover:border-solid hover:border-primary/50">
          <div className="absolute top-0 right-0 p-4 opacity-10 group-hover:opacity-20 transition-opacity">
            <Terminal className="h-24 w-24 -mr-8 -mt-8" />
          </div>
          <CardHeader>
            <div className="h-10 w-10 rounded-lg bg-purple-500/10 flex items-center justify-center mb-2 text-purple-500">
                <span className="font-bold font-mono">2</span>
            </div>
            <CardTitle>Configure Client</CardTitle>
            <CardDescription>
              Connect Claude, Cursor, or VS Code to this server.
            </CardDescription>
          </CardHeader>
          <CardContent className="flex items-end">
             {/* ConnectClientButton usually renders a trigger button. We can wrap it or style it.
                 The component renders a button by default. We can rely on that.
             */}
             <div className="w-full [&>button]:w-full">
                <ConnectClientButton />
             </div>
          </CardContent>
        </Card>

        {/* Step 3: Verify */}
        <Card className="relative overflow-hidden group hover:shadow-lg transition-all border-dashed hover:border-solid hover:border-primary/50">
          <div className="absolute top-0 right-0 p-4 opacity-10 group-hover:opacity-20 transition-opacity">
            <Wrench className="h-24 w-24 -mr-8 -mt-8" />
          </div>
          <CardHeader>
            <div className="h-10 w-10 rounded-lg bg-green-500/10 flex items-center justify-center mb-2 text-green-500">
                <span className="font-bold font-mono">3</span>
            </div>
            <CardTitle>Verify Tools</CardTitle>
            <CardDescription>
              Test your configuration in the Playground.
            </CardDescription>
          </CardHeader>
          <CardContent>
            <Link href="/playground">
                <Button variant="secondary" className="w-full gap-2">
                    Open Playground <LayoutGrid className="h-4 w-4 opacity-50" />
                </Button>
            </Link>
          </CardContent>
        </Card>
      </div>

      <div className="text-sm text-muted-foreground pt-8 flex gap-4">
          <Link href="/docs" className="hover:underline hover:text-foreground transition-colors">Documentation</Link>
          <span>•</span>
          <Link href="https://github.com/modelcontextprotocol" target="_blank" className="hover:underline hover:text-foreground transition-colors">MCP Protocol</Link>
      </div>
    </div>
  );
}
