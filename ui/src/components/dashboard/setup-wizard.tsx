/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle, CardFooter } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Tabs, TabsList, TabsTrigger, TabsContent } from "@/components/ui/tabs";
import { Check, ChevronRight, ChevronLeft, Terminal, Globe, ShoppingBag, Loader2 } from "lucide-react";
import { apiClient, UpstreamServiceConfig } from "@/lib/client";
import { useToast } from "@/hooks/use-toast";

type WizardStep = "welcome" | "path" | "details" | "finishing";

/**
 * SetupWizard component provides a guided experience for configuring the first service.
 * It replaces the static onboarding hero with an interactive flow.
 * @returns The rendered component.
 */
export function SetupWizard() {
  const router = useRouter();
  const { toast } = useToast();
  const [step, setStep] = useState<WizardStep>("welcome");
  const [selectedPath, setSelectedPath] = useState<"local" | "remote" | "marketplace" | null>(null);
  const [loading, setLoading] = useState(false);

  // Form State
  const [name, setName] = useState("");
  const [command, setCommand] = useState("");
  const [args, setArgs] = useState("");
  const [env, setEnv] = useState("");
  const [url, setUrl] = useState("");

  const handleNext = () => {
    if (step === "welcome") setStep("path");
    else if (step === "path") {
      if (selectedPath === "marketplace") {
        router.push("/marketplace");
      } else {
        setStep("details");
      }
    }
    else if (step === "details") {
      handleSubmit();
    }
  };

  const handleBack = () => {
    if (step === "path") setStep("welcome");
    else if (step === "details") setStep("path");
  };

  const handleSubmit = async () => {
    setLoading(true);
    setStep("finishing");

    try {
      const config: UpstreamServiceConfig = {
        id: "", // Server generates
        name: name || (selectedPath === "local" ? "local-server" : "remote-server"),
        disable: false,
        priority: 0,
        loadBalancingStrategy: 0,
        tags: ["setup-wizard"],
      };

      if (selectedPath === "local") {
        // Parse args properly (simple split for now, robust parsing ideally needed)
        const argsList = args.split(" ").filter(a => a.length > 0);
        // Parse env (KEY=VALUE)
        const envMap: Record<string, string> = {};
        env.split("\n").forEach(line => {
          const [k, v] = line.split("=");
          if (k && v) envMap[k.trim()] = v.trim();
        });

        config.commandLineService = {
          command: command,
          env: envMap,
          workingDirectory: "", // Default
        };
      } else if (selectedPath === "remote") {
        config.httpService = {
          address: url,
          headers: {},
        };
        // Basic detection if it's SSE or HTTP
        if (url.includes("/sse")) {
            // It might be an MCP SSE endpoint
            config.mcpService = {
                sseUrl: url
            };
            delete config.httpService;
        }
      }

      await apiClient.registerService(config);

      toast({
        title: "Service Connected",
        description: "Your first service has been successfully configured.",
      });

      // Slight delay for visual confirmation
      setTimeout(() => {
        window.location.reload(); // Reload to show dashboard state
      }, 1000);

    } catch (e) {
      console.error(e);
      toast({
        variant: "destructive",
        title: "Setup Failed",
        description: e instanceof Error ? e.message : "An error occurred.",
      });
      setStep("details"); // Go back to details
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="flex flex-col items-center justify-center min-h-[60vh] p-4 animate-in fade-in zoom-in duration-500">
      <Card className="w-full max-w-2xl shadow-xl border-primary/20">
        <div className="bg-muted/50 p-1 h-2 w-full rounded-t-lg overflow-hidden flex">
             <div className={`h-full bg-primary transition-all duration-500 ${
                 step === "welcome" ? "w-1/4" :
                 step === "path" ? "w-2/4" :
                 step === "details" ? "w-3/4" : "w-full"
             }`} />
        </div>

        {step === "welcome" && (
            <>
                <CardHeader className="text-center pb-2">
                    <div className="mx-auto w-16 h-16 bg-primary/10 rounded-full flex items-center justify-center mb-4">
                        <Terminal className="w-8 h-8 text-primary" />
                    </div>
                    <CardTitle className="text-3xl">Welcome to MCP Any</CardTitle>
                    <CardDescription className="text-lg">
                        Your unified control plane for AI tools and context.
                    </CardDescription>
                </CardHeader>
                <CardContent className="text-center space-y-4 pt-4">
                    <p className="text-muted-foreground">
                        MCP Any connects your AI agents to real-world data and actions.
                        Let's get your first service connected in less than a minute.
                    </p>
                </CardContent>
                <CardFooter className="justify-center pt-4">
                    <Button size="lg" onClick={handleNext} className="w-full sm:w-auto px-8">
                        Get Started <ChevronRight className="ml-2 h-4 w-4" />
                    </Button>
                </CardFooter>
            </>
        )}

        {step === "path" && (
            <>
                <CardHeader>
                    <CardTitle>How do you want to connect?</CardTitle>
                    <CardDescription>Choose the type of service you are adding.</CardDescription>
                </CardHeader>
                <CardContent className="grid gap-4 md:grid-cols-3">
                    <PathCard
                        icon={<Terminal className="h-8 w-8 text-blue-500" />}
                        title="Local Service"
                        description="Run a local command (stdio)"
                        selected={selectedPath === "local"}
                        onClick={() => setSelectedPath("local")}
                    />
                    <PathCard
                        icon={<Globe className="h-8 w-8 text-green-500" />}
                        title="Remote API"
                        description="Connect via HTTP/SSE"
                        selected={selectedPath === "remote"}
                        onClick={() => setSelectedPath("remote")}
                    />
                    <PathCard
                        icon={<ShoppingBag className="h-8 w-8 text-purple-500" />}
                        title="Marketplace"
                        description="Browse community tools"
                        selected={selectedPath === "marketplace"}
                        onClick={() => setSelectedPath("marketplace")}
                    />
                </CardContent>
                <CardFooter className="justify-between">
                    <Button variant="ghost" onClick={handleBack}>Back</Button>
                    <Button onClick={handleNext} disabled={!selectedPath}>Next <ChevronRight className="ml-2 h-4 w-4" /></Button>
                </CardFooter>
            </>
        )}

        {step === "details" && selectedPath === "local" && (
            <>
                <CardHeader>
                    <CardTitle>Configure Local Service</CardTitle>
                    <CardDescription>Enter the command to run your MCP server.</CardDescription>
                </CardHeader>
                <CardContent className="space-y-4">
                    <div className="space-y-2">
                        <Label htmlFor="name">Service Name</Label>
                        <Input id="name" placeholder="e.g. my-local-server" value={name} onChange={e => setName(e.target.value)} />
                    </div>
                    <div className="space-y-2">
                        <Label htmlFor="command">Command</Label>
                        <Input id="command" placeholder="e.g. npx" value={command} onChange={e => setCommand(e.target.value)} />
                    </div>
                    <div className="space-y-2">
                        <Label htmlFor="args">Arguments</Label>
                        <Input id="args" placeholder="e.g. -y @modelcontextprotocol/server-filesystem /path/to/files" value={args} onChange={e => setArgs(e.target.value)} />
                    </div>
                    <div className="space-y-2">
                        <Label htmlFor="env">Environment Variables (Optional)</Label>
                        <Input id="env" placeholder="KEY=VALUE (one per line, comma separated... actually just one line for now)" value={env} onChange={e => setEnv(e.target.value)} />
                        <p className="text-xs text-muted-foreground">For multiple vars, functionality is limited in this wizard. Use the advanced editor later.</p>
                    </div>
                </CardContent>
                <CardFooter className="justify-between">
                    <Button variant="ghost" onClick={handleBack}>Back</Button>
                    <Button onClick={handleNext} disabled={!command || !name}>Connect Service</Button>
                </CardFooter>
            </>
        )}

        {step === "details" && selectedPath === "remote" && (
            <>
                <CardHeader>
                    <CardTitle>Configure Remote Service</CardTitle>
                    <CardDescription>Enter the URL of the remote MCP server or API.</CardDescription>
                </CardHeader>
                <CardContent className="space-y-4">
                    <div className="space-y-2">
                        <Label htmlFor="name">Service Name</Label>
                        <Input id="name" placeholder="e.g. weather-api" value={name} onChange={e => setName(e.target.value)} />
                    </div>
                    <div className="space-y-2">
                        <Label htmlFor="url">URL</Label>
                        <Input id="url" placeholder="e.g. https://api.weather.gov or http://localhost:8000/sse" value={url} onChange={e => setUrl(e.target.value)} />
                        <p className="text-xs text-muted-foreground">If the URL ends in `/sse`, we'll treat it as an MCP SSE connection.</p>
                    </div>
                </CardContent>
                <CardFooter className="justify-between">
                    <Button variant="ghost" onClick={handleBack}>Back</Button>
                    <Button onClick={handleNext} disabled={!url || !name}>Connect Service</Button>
                </CardFooter>
            </>
        )}

        {step === "finishing" && (
            <div className="py-12 flex flex-col items-center justify-center space-y-4">
                <Loader2 className="h-12 w-12 animate-spin text-primary" />
                <h3 className="text-xl font-medium">Connecting...</h3>
                <p className="text-muted-foreground">We are verifying the connection to your service.</p>
            </div>
        )}
      </Card>
    </div>
  );
}

function PathCard({ icon, title, description, selected, onClick }: { icon: React.ReactNode, title: string, description: string, selected: boolean, onClick: () => void }) {
    return (
        <div
            onClick={onClick}
            className={`cursor-pointer rounded-lg border-2 p-4 hover:bg-muted/50 transition-all flex flex-col items-center text-center space-y-2 ${
                selected ? "border-primary bg-primary/5" : "border-transparent bg-muted/20"
            }`}
        >
            <div className="p-2 rounded-full bg-background shadow-sm">
                {icon}
            </div>
            <h3 className="font-semibold">{title}</h3>
            <p className="text-xs text-muted-foreground">{description}</p>
            {selected && <div className="absolute top-2 right-2"><Check className="h-4 w-4 text-primary" /></div>}
        </div>
    )
}
