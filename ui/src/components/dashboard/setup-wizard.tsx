/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { apiClient, UpstreamServiceConfig } from "@/lib/client";
import { useToast } from "@/hooks/use-toast";
import { Loader2, Terminal, Globe, ShoppingBag, CheckCircle, ArrowRight, ArrowLeft } from "lucide-react";
import Link from "next/link";
import { Tabs, TabsList, TabsTrigger, TabsContent } from "@/components/ui/tabs";

enum WizardStep {
  WELCOME = 0,
  SELECT_TYPE = 1,
  CONFIGURE = 2,
  SUCCESS = 3,
}

type ServiceType = "local" | "remote" | "marketplace";

/**
 * SetupWizard component.
 * Guides the user through the initial setup of connecting their first MCP service.
 * @returns The rendered component.
 */
export function SetupWizard() {
  const [step, setStep] = useState<WizardStep>(WizardStep.WELCOME);
  const [serviceType, setServiceType] = useState<ServiceType | null>(null);
  const [loading, setLoading] = useState(false);
  const { toast } = useToast();

  // Form State
  const [name, setName] = useState("");
  const [command, setCommand] = useState(""); // For local
  const [url, setUrl] = useState(""); // For remote
  const [apiKey, setApiKey] = useState(""); // For remote (optional)

  const handleSelectType = (type: ServiceType) => {
    if (type === "marketplace") {
      // Redirect immediately or show instructions?
      // Redirect seems best for now as Marketplace is separate.
      // But let's keep it consistent.
      window.location.href = "/marketplace";
      return;
    }
    setServiceType(type);
    setStep(WizardStep.CONFIGURE);
  };

  const handleSubmit = async () => {
    setLoading(true);
    try {
      const config: Partial<UpstreamServiceConfig> = {
        name: name || (serviceType === "local" ? "local-mcp" : "remote-mcp"),
        disable: false,
        priority: 0,
        version: "1.0.0",
        id: "", // Let server generate ID
      };

      if (serviceType === "local") {
        if (!command) throw new Error("Command is required");
        config.commandLineService = {
          command: command,
          env: {},
          workingDirectory: "",
          tools: [],
          resources: [],
          prompts: [],
          calls: {},
          communicationProtocol: 0,
          local: true,
        };
      } else if (serviceType === "remote") {
        if (!url) throw new Error("URL is required");
        config.httpService = {
          url: url,
          headers: apiKey ? { "Authorization": `Bearer ${apiKey}` } : {},
          toolUrl: `${url}/tools`, // Simple assumption for now, user can edit later
          resourceUrl: `${url}/resources`,
          promptUrl: `${url}/prompts`,
          sseUrl: "",
        };
      }

      await apiClient.registerService(config as UpstreamServiceConfig);
      toast({
        title: "Service Created",
        description: `Service ${config.name} has been successfully registered.`,
      });
      setStep(WizardStep.SUCCESS);
    } catch (e: any) {
      console.error(e);
      toast({
        variant: "destructive",
        title: "Error",
        description: e.message || "Failed to create service.",
      });
    } finally {
      setLoading(false);
    }
  };

  const renderWelcome = () => (
    <div className="flex flex-col items-center justify-center space-y-8 animate-in fade-in slide-in-from-bottom-4 duration-500">
      <div className="text-center space-y-4 max-w-2xl">
        <h1 className="text-4xl font-bold tracking-tight bg-gradient-to-r from-primary to-primary/60 bg-clip-text text-transparent">
          Welcome to MCP Any
        </h1>
        <p className="text-xl text-muted-foreground">
          Let's get you set up. Connect your first Model Context Protocol service to start using tools with your AI agents.
        </p>
      </div>
      <Button size="lg" onClick={() => setStep(WizardStep.SELECT_TYPE)} className="group">
        Get Started <ArrowRight className="ml-2 h-4 w-4 group-hover:translate-x-1 transition-transform" />
      </Button>
    </div>
  );

  const renderSelectType = () => (
    <div className="w-full max-w-4xl space-y-6 animate-in fade-in slide-in-from-right-4 duration-300">
      <div className="flex items-center gap-4 mb-8">
          <Button variant="ghost" onClick={() => setStep(WizardStep.WELCOME)}>
              <ArrowLeft className="mr-2 h-4 w-4" /> Back
          </Button>
          <h2 className="text-2xl font-bold">Choose a Connection Type</h2>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
        <Card
            className="hover:border-primary cursor-pointer transition-all hover:shadow-lg group relative overflow-hidden"
            onClick={() => handleSelectType("local")}
        >
            <div className="absolute inset-0 bg-gradient-to-br from-primary/5 to-transparent opacity-0 group-hover:opacity-100 transition-opacity pointer-events-none" />
            <CardHeader>
                <CardTitle className="flex items-center gap-2">
                    <Terminal className="h-5 w-5 text-primary" />
                    Local MCP Server
                </CardTitle>
                <CardDescription>Run a local command (stdio)</CardDescription>
            </CardHeader>
            <CardContent>
                <p className="text-sm text-muted-foreground">
                    Perfect for running Python or Node.js MCP servers directly on this machine.
                </p>
                <div className="mt-4 p-2 bg-muted rounded text-xs font-mono opacity-70">
                    uvx mcp-server-git ...
                </div>
            </CardContent>
        </Card>

        <Card
            className="hover:border-primary cursor-pointer transition-all hover:shadow-lg group relative overflow-hidden"
            onClick={() => handleSelectType("remote")}
        >
            <div className="absolute inset-0 bg-gradient-to-br from-blue-500/5 to-transparent opacity-0 group-hover:opacity-100 transition-opacity pointer-events-none" />
            <CardHeader>
                <CardTitle className="flex items-center gap-2">
                    <Globe className="h-5 w-5 text-blue-500" />
                    Remote API (HTTP)
                </CardTitle>
                <CardDescription>Connect to a remote MCP endpoint</CardDescription>
            </CardHeader>
            <CardContent>
                <p className="text-sm text-muted-foreground">
                    Connect to an existing MCP server running over HTTP/SSE.
                </p>
                <div className="mt-4 p-2 bg-muted rounded text-xs font-mono opacity-70">
                    https://api.example.com/mcp
                </div>
            </CardContent>
        </Card>

        <Card
            className="hover:border-primary cursor-pointer transition-all hover:shadow-lg group relative overflow-hidden"
            onClick={() => handleSelectType("marketplace")}
        >
            <div className="absolute inset-0 bg-gradient-to-br from-purple-500/5 to-transparent opacity-0 group-hover:opacity-100 transition-opacity pointer-events-none" />
            <CardHeader>
                <CardTitle className="flex items-center gap-2">
                    <ShoppingBag className="h-5 w-5 text-purple-500" />
                    Marketplace
                </CardTitle>
                <CardDescription>Install pre-configured services</CardDescription>
            </CardHeader>
            <CardContent>
                <p className="text-sm text-muted-foreground">
                    Browse the official and community catalog for popular services like GitHub, Slack, and Postgres.
                </p>
            </CardContent>
        </Card>
      </div>
    </div>
  );

  const renderConfigure = () => (
    <div className="w-full max-w-lg space-y-6 animate-in fade-in slide-in-from-right-4 duration-300 mx-auto">
        <div className="flex items-center gap-4 mb-4">
          <Button variant="ghost" onClick={() => setStep(WizardStep.SELECT_TYPE)}>
              <ArrowLeft className="mr-2 h-4 w-4" /> Back
          </Button>
          <h2 className="text-2xl font-bold">
              Configure {serviceType === "local" ? "Local Server" : "Remote Service"}
          </h2>
        </div>

        <Card>
            <CardContent className="space-y-4 pt-6">
                <div className="space-y-2">
                    <Label htmlFor="name">Service Name</Label>
                    <Input
                        id="name"
                        placeholder="e.g. my-git-server"
                        value={name}
                        onChange={(e) => setName(e.target.value)}
                    />
                    <p className="text-xs text-muted-foreground">A unique identifier for this service.</p>
                </div>

                {serviceType === "local" ? (
                    <div className="space-y-2">
                        <Label htmlFor="command">Command</Label>
                        <Input
                            id="command"
                            placeholder="e.g. npx -y @modelcontextprotocol/server-filesystem /path/to/files"
                            value={command}
                            onChange={(e) => setCommand(e.target.value)}
                        />
                        <p className="text-xs text-muted-foreground">
                            The full command to execute the MCP server. Ensure it outputs JSON-RPC to stdio.
                        </p>
                    </div>
                ) : (
                    <>
                        <div className="space-y-2">
                            <Label htmlFor="url">Base URL</Label>
                            <Input
                                id="url"
                                placeholder="https://api.myservice.com/mcp"
                                value={url}
                                onChange={(e) => setUrl(e.target.value)}
                            />
                        </div>
                        <div className="space-y-2">
                            <Label htmlFor="apiKey">API Key (Optional)</Label>
                            <Input
                                id="apiKey"
                                type="password"
                                placeholder="sk-..."
                                value={apiKey}
                                onChange={(e) => setApiKey(e.target.value)}
                            />
                            <p className="text-xs text-muted-foreground">
                                Will be sent as a Bearer token in the Authorization header.
                            </p>
                        </div>
                    </>
                )}
            </CardContent>
            <CardFooter>
                <Button className="w-full" onClick={handleSubmit} disabled={loading}>
                    {loading && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
                    {loading ? "Registering..." : "Connect Service"}
                </Button>
            </CardFooter>
        </Card>
    </div>
  );

  const renderSuccess = () => (
      <div className="flex flex-col items-center justify-center space-y-6 animate-in zoom-in duration-300">
          <div className="rounded-full bg-green-100 dark:bg-green-900/20 p-6">
              <CheckCircle className="h-12 w-12 text-green-500" />
          </div>
          <div className="text-center space-y-2">
              <h2 className="text-2xl font-bold">Service Connected!</h2>
              <p className="text-muted-foreground">
                  Your service is now registered and ready to accept tool calls.
              </p>
          </div>
          <div className="flex gap-4">
              <Link href="/">
                  <Button variant="outline">
                      Go to Dashboard
                  </Button>
              </Link>
              <Link href="/playground">
                  <Button>
                      Try in Playground <ArrowRight className="ml-2 h-4 w-4" />
                  </Button>
              </Link>
          </div>
      </div>
  );

  return (
    <div className="flex flex-col items-center justify-center min-h-[60vh] p-4 w-full">
      {step === WizardStep.WELCOME && renderWelcome()}
      {step === WizardStep.SELECT_TYPE && renderSelectType()}
      {step === WizardStep.CONFIGURE && renderConfigure()}
      {step === WizardStep.SUCCESS && renderSuccess()}
    </div>
  );
}
