/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { RadioGroup, RadioGroupItem } from "@/components/ui/radio-group";
import { Check, ChevronRight, Terminal, Globe, ShoppingBag, Loader2 } from "lucide-react";
import { apiClient, UpstreamServiceConfig } from "@/lib/client";
import { useToast } from "@/hooks/use-toast";
import { useRouter } from "next/navigation";

// Steps
enum Step {
  WELCOME = 0,
  SELECT_TYPE = 1,
  CONFIGURE = 2,
  COMPLETE = 3
}

// Service Types
enum ServiceType {
  LOCAL = "local",
  REMOTE = "remote",
  MARKETPLACE = "marketplace"
}

export function SetupWizard() {
  const [step, setStep] = useState<Step>(Step.WELCOME);
  const [serviceType, setServiceType] = useState<ServiceType>(ServiceType.LOCAL);
  const [loading, setLoading] = useState(false);

  // Local Config
  const [command, setCommand] = useState("");

  // Remote Config
  const [url, setUrl] = useState("");

  // Common
  const [name, setName] = useState("");

  const { toast } = useToast();
  const router = useRouter();

  const handleNext = () => {
    if (step === Step.WELCOME) {
      setStep(Step.SELECT_TYPE);
    } else if (step === Step.SELECT_TYPE) {
      if (serviceType === ServiceType.MARKETPLACE) {
        router.push("/marketplace");
      } else {
        setStep(Step.CONFIGURE);
      }
    } else if (step === Step.CONFIGURE) {
      handleSubmit();
    }
  };

  const handleBack = () => {
    if (step > Step.WELCOME) {
      setStep(step - 1);
    }
  };

  const handleSubmit = async () => {
    if (!name) {
      toast({ variant: "destructive", title: "Name is required" });
      return;
    }

    setLoading(true);
    try {
      const config: UpstreamServiceConfig = {
        id: name.toLowerCase().replace(/[^a-z0-9-]/g, "-"),
        name: name,
        version: "1.0.0",
        disable: false,
        priority: 0,
        loadBalancingStrategy: 0,
        tags: ["setup-wizard"],
        toolCount: 0, // Generated field, ignore
      };

      if (serviceType === ServiceType.LOCAL) {
        if (!command) {
            toast({ variant: "destructive", title: "Command is required" });
            setLoading(false);
            return;
        }
        config.commandLineService = {
          command: command,
          workingDirectory: "", // Optional
          env: {},
          tools: [],
          resources: [],
          prompts: [],
          calls: {},
          communicationProtocol: 0, // Stdio
          local: true
        };
      } else if (serviceType === ServiceType.REMOTE) {
         if (!url) {
            toast({ variant: "destructive", title: "URL is required" });
            setLoading(false);
            return;
        }
        // Assuming SSE for remote
        config.httpService = {
            url: url,
            headers: {},
            toolUrl: "", // Optional specific endpoints
            resourceUrl: "",
            promptUrl: "",
            sse: true // Enable SSE by default for "Remote MCP"
        };
      }

      await apiClient.registerService(config);
      setStep(Step.COMPLETE);
      toast({ title: "Service Connected", description: `Successfully connected to ${name}.` });

      // Refresh or redirect
      // For now, show complete step
    } catch (e) {
      console.error(e);
      toast({ variant: "destructive", title: "Connection Failed", description: String(e) });
    } finally {
      setLoading(false);
    }
  };

  const renderWelcome = () => (
    <div className="flex flex-col items-center text-center space-y-6 animate-in fade-in slide-in-from-bottom-4 duration-500">
      <div className="space-y-2">
        <h1 className="text-3xl font-bold tracking-tight">Welcome to MCP Any</h1>
        <p className="text-muted-foreground max-w-md mx-auto">
          Your unified control plane for the Model Context Protocol. Let's get your first service connected.
        </p>
      </div>
      <Button size="lg" onClick={handleNext} className="w-full max-w-sm">
        Get Started <ChevronRight className="ml-2 h-4 w-4" />
      </Button>
    </div>
  );

  const renderSelectType = () => (
    <div className="space-y-6">
      <div className="text-center">
        <h2 className="text-2xl font-bold">Choose Connection Type</h2>
        <p className="text-muted-foreground">How do you want to connect your MCP server?</p>
      </div>

      <RadioGroup
        value={serviceType}
        onValueChange={(v) => setServiceType(v as ServiceType)}
        className="grid grid-cols-1 md:grid-cols-3 gap-4"
      >
        <div className={`
            cursor-pointer rounded-lg border-2 p-4 hover:border-primary transition-all
            ${serviceType === ServiceType.LOCAL ? "border-primary bg-primary/5" : "border-muted"}
        `} onClick={() => setServiceType(ServiceType.LOCAL)}>
            <RadioGroupItem value={ServiceType.LOCAL} id="local" className="sr-only" />
            <div className="flex flex-col items-center text-center gap-2">
                <Terminal className="h-8 w-8 text-primary" />
                <Label htmlFor="local" className="font-semibold cursor-pointer">Local Command</Label>
                <p className="text-xs text-muted-foreground">Run a local executable or script (stdio).</p>
            </div>
        </div>

        <div className={`
            cursor-pointer rounded-lg border-2 p-4 hover:border-primary transition-all
            ${serviceType === ServiceType.REMOTE ? "border-primary bg-primary/5" : "border-muted"}
        `} onClick={() => setServiceType(ServiceType.REMOTE)}>
            <RadioGroupItem value={ServiceType.REMOTE} id="remote" className="sr-only" />
             <div className="flex flex-col items-center text-center gap-2">
                <Globe className="h-8 w-8 text-blue-500" />
                <Label htmlFor="remote" className="font-semibold cursor-pointer">Remote Server</Label>
                <p className="text-xs text-muted-foreground">Connect to an MCP server via SSE (HTTP).</p>
            </div>
        </div>

        <div className={`
            cursor-pointer rounded-lg border-2 p-4 hover:border-primary transition-all
            ${serviceType === ServiceType.MARKETPLACE ? "border-primary bg-primary/5" : "border-muted"}
        `} onClick={() => setServiceType(ServiceType.MARKETPLACE)}>
             <RadioGroupItem value={ServiceType.MARKETPLACE} id="marketplace" className="sr-only" />
             <div className="flex flex-col items-center text-center gap-2">
                <ShoppingBag className="h-8 w-8 text-purple-500" />
                <Label htmlFor="marketplace" className="font-semibold cursor-pointer">Marketplace</Label>
                <p className="text-xs text-muted-foreground">Install pre-configured community servers.</p>
            </div>
        </div>
      </RadioGroup>

      <div className="flex justify-between pt-4">
        <Button variant="ghost" onClick={handleBack}>Back</Button>
        <Button onClick={handleNext}>Continue <ChevronRight className="ml-2 h-4 w-4" /></Button>
      </div>
    </div>
  );

  const renderConfigure = () => (
    <div className="space-y-6 max-w-md mx-auto">
      <div className="text-center">
        <h2 className="text-2xl font-bold">Configure {serviceType === ServiceType.LOCAL ? "Local" : "Remote"} Service</h2>
        <p className="text-muted-foreground">Enter the connection details below.</p>
      </div>

      <div className="space-y-4">
        <div className="space-y-2">
            <Label htmlFor="name">Service Name</Label>
            <Input
                id="name"
                placeholder="e.g. My Service"
                value={name}
                onChange={(e) => setName(e.target.value)}
            />
        </div>

        {serviceType === ServiceType.LOCAL ? (
             <div className="space-y-2">
                <Label htmlFor="command">Command</Label>
                <Input
                    id="command"
                    placeholder="e.g. npx -y @modelcontextprotocol/server-filesystem /path/to/files"
                    value={command}
                    onChange={(e) => setCommand(e.target.value)}
                />
                <p className="text-xs text-muted-foreground">The full command to execute the server.</p>
            </div>
        ) : (
             <div className="space-y-2">
                <Label htmlFor="url">Server URL (SSE)</Label>
                <Input
                    id="url"
                    placeholder="e.g. http://localhost:8080/sse"
                    value={url}
                    onChange={(e) => setUrl(e.target.value)}
                />
            </div>
        )}
      </div>

       <div className="flex justify-between pt-4">
        <Button variant="ghost" onClick={handleBack} disabled={loading}>Back</Button>
        <Button onClick={handleSubmit} disabled={loading}>
            {loading ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : <Check className="mr-2 h-4 w-4" />}
            Connect Service
        </Button>
      </div>
    </div>
  );

  const renderComplete = () => (
     <div className="flex flex-col items-center text-center space-y-6 animate-in zoom-in duration-300">
      <div className="rounded-full bg-green-100 p-4 dark:bg-green-900/20">
        <Check className="h-12 w-12 text-green-600 dark:text-green-400" />
      </div>
      <div className="space-y-2">
        <h2 className="text-2xl font-bold">Setup Complete!</h2>
        <p className="text-muted-foreground">
          Your service <strong>{name}</strong> is now connected and ready to use.
        </p>
      </div>
      <div className="flex gap-4">
        <Button variant="outline" onClick={() => window.location.reload()}>
            Go to Dashboard
        </Button>
        <Button onClick={() => router.push("/playground")}>
            Try in Playground
        </Button>
      </div>
    </div>
  );

  return (
    <div className="flex flex-col items-center justify-center min-h-[60vh] p-4">
      <Card className="w-full max-w-2xl border-none shadow-none bg-transparent">
        <CardContent className="p-6">
            {step === Step.WELCOME && renderWelcome()}
            {step === Step.SELECT_TYPE && renderSelectType()}
            {step === Step.CONFIGURE && renderConfigure()}
            {step === Step.COMPLETE && renderComplete()}
        </CardContent>
      </Card>
    </div>
  );
}
