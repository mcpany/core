/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { CloudSun, Terminal, Globe, ArrowRight, Loader2, Plus, Sparkles } from "lucide-react";
import { apiClient } from "@/lib/client";
import { useToast } from "@/hooks/use-toast";
import { RegisterServiceDialog } from "@/components/register-service-dialog";

/**
 * OnboardingView component.
 * displayed when no services are registered.
 * @returns The rendered component.
 */
export function OnboardingView() {
  const router = useRouter();
  const { toast } = useToast();
  const [installingDemo, setInstallingDemo] = useState(false);

  const handleInstallDemo = async () => {
    setInstallingDemo(true);
    try {
      await apiClient.registerService({
        name: "weather-demo",
        id: "weather-demo",
        version: "1.0.0",
        disable: false,
        priority: 0,
        loadBalancingStrategy: 0,
        tags: ["demo", "weather"],
        httpService: {
          address: "https://wttr.in",
          tools: [
            {
              name: "get_weather",
              description: "Get weather for a location",
              inputSchema: {
                type: "object",
                properties: {
                  location: {
                    type: "string",
                    description: "The location to get weather for",
                  },
                },
                required: ["location"],
              },
            },
          ],
          calls: {},
          resources: [],
          prompts: [],
        },
        readOnly: false,
        sanitizedName: "weather-demo",
        callPolicies: [],
        preCallHooks: [],
        postCallHooks: [],
        prompts: [],
        autoDiscoverTool: false,
        configError: "",
      });

      toast({
        title: "Demo Installed",
        description: "The Weather Demo service has been registered.",
      });

      // Reload to refresh the dashboard state
      window.location.reload();
    } catch (error: any) {
      console.error("Failed to install demo", error);
      toast({
        title: "Installation Failed",
        description: error.message || "Could not register the demo service.",
        variant: "destructive",
      });
    } finally {
      setInstallingDemo(false);
    }
  };

  return (
    <div className="flex flex-col items-center justify-center min-h-[60vh] animate-in fade-in zoom-in duration-500">
      <div className="text-center mb-10 max-w-2xl px-4">
        <div className="inline-flex items-center justify-center p-3 bg-primary/10 rounded-full mb-6">
          <Sparkles className="h-8 w-8 text-primary" />
        </div>
        <h1 className="text-4xl font-bold tracking-tight mb-4">Welcome to MCP Any</h1>
        <p className="text-lg text-muted-foreground">
          The definitive platform for managing Model Context Protocol servers.
          Connect your tools, agents, and data in one place.
        </p>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-3 gap-6 w-full max-w-5xl px-4">
        <Card className="flex flex-col hover:border-primary/50 transition-all cursor-pointer group hover:shadow-lg">
          <CardHeader>
            <div className="w-12 h-12 rounded-lg bg-blue-100 dark:bg-blue-900/30 flex items-center justify-center mb-4 text-blue-600 dark:text-blue-400">
              <CloudSun className="h-6 w-6" />
            </div>
            <CardTitle>Quick Start: Weather</CardTitle>
            <CardDescription>
              Install a demo HTTP adapter for wttr.in to see MCP in action instantly.
            </CardDescription>
          </CardHeader>
          <CardContent className="flex-1">
            <p className="text-sm text-muted-foreground">
              Perfect for verifying your setup. Adds a `get_weather` tool.
            </p>
          </CardContent>
          <CardFooter>
            <Button
                className="w-full group-hover:bg-primary/90"
                onClick={handleInstallDemo}
                disabled={installingDemo}
            >
              {installingDemo ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : <Plus className="mr-2 h-4 w-4" />}
              {installingDemo ? "Installing..." : "Install Demo"}
            </Button>
          </CardFooter>
        </Card>

        <Card className="flex flex-col hover:border-primary/50 transition-all cursor-pointer group hover:shadow-lg">
          <CardHeader>
            <div className="w-12 h-12 rounded-lg bg-orange-100 dark:bg-orange-900/30 flex items-center justify-center mb-4 text-orange-600 dark:text-orange-400">
              <Terminal className="h-6 w-6" />
            </div>
            <CardTitle>Connect Local Service</CardTitle>
            <CardDescription>
              Register an existing MCP server running locally or via command line.
            </CardDescription>
          </CardHeader>
          <CardContent className="flex-1">
            <p className="text-sm text-muted-foreground">
              Use the wizard to connect stdio, HTTP (SSE), or OpenAPI services.
            </p>
          </CardContent>
          <CardFooter>
            <RegisterServiceDialog
                trigger={
                    <Button variant="outline" className="w-full">
                        Open Wizard <ArrowRight className="ml-2 h-4 w-4 opacity-0 group-hover:opacity-100 transition-opacity" />
                    </Button>
                }
                onSuccess={() => window.location.reload()}
            />
          </CardFooter>
        </Card>

        <Card className="flex flex-col hover:border-primary/50 transition-all cursor-pointer group hover:shadow-lg" onClick={() => router.push("/marketplace")}>
          <CardHeader>
            <div className="w-12 h-12 rounded-lg bg-purple-100 dark:bg-purple-900/30 flex items-center justify-center mb-4 text-purple-600 dark:text-purple-400">
              <Globe className="h-6 w-6" />
            </div>
            <CardTitle>Browse Marketplace</CardTitle>
            <CardDescription>
              Discover and install community servers and official integrations.
            </CardDescription>
          </CardHeader>
          <CardContent className="flex-1">
            <p className="text-sm text-muted-foreground">
              Explore hundreds of ready-to-use MCP servers for GitHub, Google, and more.
            </p>
          </CardContent>
          <CardFooter>
            <Button variant="outline" className="w-full">
              Explore <ArrowRight className="ml-2 h-4 w-4 opacity-0 group-hover:opacity-100 transition-opacity" />
            </Button>
          </CardFooter>
        </Card>
      </div>
    </div>
  );
}
