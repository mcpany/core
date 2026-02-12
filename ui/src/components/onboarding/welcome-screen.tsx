/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { Zap, Terminal, LayoutDashboard, ArrowRight } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card, CardDescription, CardFooter, CardHeader, CardTitle } from "@/components/ui/card";

interface WelcomeScreenProps {
  onRegister: () => void;
  onTemplate: () => void;
}

/**
 * WelcomeScreen component.
 * @param props - The component props.
 * @param props.onRegister - The onRegister callback.
 * @param props.onTemplate - The onTemplate callback.
 * @returns The rendered component.
 */
export function WelcomeScreen({ onRegister, onTemplate }: WelcomeScreenProps) {
  return (
    <div className="flex flex-col items-center justify-center min-h-[80vh] p-8 space-y-12 animate-in fade-in duration-700">
      <div className="text-center space-y-4 max-w-2xl">
        <div className="mx-auto w-16 h-16 bg-primary/10 rounded-2xl flex items-center justify-center mb-6">
            <LayoutDashboard className="h-8 w-8 text-primary" />
        </div>
        <h1 className="text-4xl md:text-5xl font-bold tracking-tight bg-gradient-to-br from-foreground to-muted-foreground bg-clip-text text-transparent">
          Welcome to MCP Any
        </h1>
        <p className="text-xl text-muted-foreground leading-relaxed">
          The definitive management console for Model Context Protocol.
          <br />
          Connect your tools, resources, and services in seconds.
        </p>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-6 w-full max-w-4xl">
        <Card className="group hover:border-primary/50 transition-all hover:shadow-lg cursor-pointer bg-card/50 backdrop-blur-sm" onClick={onTemplate}>
          <CardHeader>
            <div className="mb-2 w-10 h-10 rounded-lg bg-blue-500/10 flex items-center justify-center group-hover:bg-blue-500/20 transition-colors">
              <Zap className="h-5 w-5 text-blue-500" />
            </div>
            <CardTitle className="text-xl">Quick Start</CardTitle>
            <CardDescription>
              Use a pre-configured template for popular services like GitHub, Google Drive, or Slack.
            </CardDescription>
          </CardHeader>
          <CardFooter>
            <Button variant="ghost" className="p-0 hover:bg-transparent text-blue-500 group-hover:translate-x-1 transition-transform">
              Browse Templates <ArrowRight className="ml-2 h-4 w-4" />
            </Button>
          </CardFooter>
        </Card>

        <Card className="group hover:border-primary/50 transition-all hover:shadow-lg cursor-pointer bg-card/50 backdrop-blur-sm" onClick={onRegister}>
          <CardHeader>
            <div className="mb-2 w-10 h-10 rounded-lg bg-orange-500/10 flex items-center justify-center group-hover:bg-orange-500/20 transition-colors">
              <Terminal className="h-5 w-5 text-orange-500" />
            </div>
            <CardTitle className="text-xl">Connect Manually</CardTitle>
            <CardDescription>
              Connect an existing MCP server via Stdio (Command Line) or SSE (HTTP).
            </CardDescription>
          </CardHeader>
          <CardFooter>
            <Button variant="ghost" className="p-0 hover:bg-transparent text-orange-500 group-hover:translate-x-1 transition-transform">
              Configure Service <ArrowRight className="ml-2 h-4 w-4" />
            </Button>
          </CardFooter>
        </Card>
      </div>

      <div className="text-sm text-muted-foreground">
        Need help? Check out the <a href="#" className="underline hover:text-foreground">documentation</a> or <a href="#" className="underline hover:text-foreground">community examples</a>.
      </div>
    </div>
  );
}
