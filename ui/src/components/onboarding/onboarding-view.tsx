/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState } from "react";
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Badge } from "@/components/ui/badge";
import { Terminal, Server, ArrowRight, Zap, CheckCircle2, Loader2, AlertCircle } from "lucide-react";
import { motion, AnimatePresence } from "framer-motion";
import { apiClient, UpstreamServiceConfig } from "@/lib/client";
import { useToast } from "@/hooks/use-toast";

export type OnboardingStep = "selection" | "configuration" | "validation" | "success";

interface OnboardingViewProps {
    onComplete: () => void;
}

export function OnboardingView({ onComplete }: OnboardingViewProps) {
    const [step, setStep] = useState<OnboardingStep>("selection");
    const [selectedType, setSelectedType] = useState<"quick" | "local" | "remote" | null>(null);
    const [loading, setLoading] = useState(false);
    const [config, setConfig] = useState<{ command?: string; url?: string; name: string }>({ name: "my-first-service" });
    const { toast } = useToast();

    const handleSelect = (type: "quick" | "local" | "remote") => {
        setSelectedType(type);
        if (type === "quick") {
            setConfig({ name: "weather-service", url: "https://wttr.in" });
        } else if (type === "local") {
            setConfig({ name: "local-cli", command: "python3 server.py" });
        } else {
            setConfig({ name: "remote-mcp", url: "http://localhost:8080/sse" });
        }
        setStep("configuration");
    };

    const handleConnect = async () => {
        setLoading(true);
        setStep("validation");

        try {
            let serviceConfig: UpstreamServiceConfig;

            if (selectedType === "quick") {
                // Register wttr.in (HTTP Service)
                serviceConfig = {
                    id: "weather-service",
                    name: "weather-service",
                    version: "1.0.0",
                    disable: false,
                    priority: 0,
                    loadBalancingStrategy: 0,
                    sanitizedName: "weather-service",
                    readOnly: false,
                    callPolicies: [],
                    preCallHooks: [],
                    postCallHooks: [],
                    prompts: [],
                    autoDiscoverTool: true, // Auto discover for HTTP? Usually needs OpenAPI.
                    // But wttr.in is simple HTTP. Let's assume we treat it as generic HTTP service.
                    httpService: {
                        address: "https://wttr.in",
                        tools: [], // We rely on manual or auto-discovery
                        calls: {},
                        resources: [],
                        prompts: [],
                    },
                    tags: ["demo", "weather"]
                };
            } else if (selectedType === "local") {
                serviceConfig = {
                    id: config.name,
                    name: config.name,
                    version: "1.0.0",
                    disable: false,
                    priority: 0,
                    loadBalancingStrategy: 0,
                    sanitizedName: config.name,
                    readOnly: false,
                    callPolicies: [],
                    preCallHooks: [],
                    postCallHooks: [],
                    prompts: [],
                    autoDiscoverTool: true,
                    commandLineService: {
                        command: config.command || "",
                        workingDirectory: ".",
                        env: {},
                        local: true,
                        tools: [],
                        resources: [],
                        prompts: [],
                        calls: {},
                        communicationProtocol: 0
                    },
                    tags: ["local"]
                };
            } else {
                // Remote SSE
                // We use MCP Service type or HTTP Service?
                // MCP Service is for "MCP over HTTP (SSE)"
                serviceConfig = {
                    id: config.name,
                    name: config.name,
                    version: "1.0.0",
                    disable: false,
                    priority: 0,
                    loadBalancingStrategy: 0,
                    sanitizedName: config.name,
                    readOnly: false,
                    callPolicies: [],
                    preCallHooks: [],
                    postCallHooks: [],
                    prompts: [],
                    autoDiscoverTool: true,
                    mcpService: {
                        toolAutoDiscovery: true,
                        httpConnection: {
                            httpAddress: config.url || "",
                        },
                        tools: [],
                        resources: [],
                        prompts: [],
                        calls: {}
                    },
                    tags: ["remote"]
                };
            }

            // 1. Validate
            const validation = await apiClient.validateService(serviceConfig);
            if (!validation.valid) {
                throw new Error(validation.message || "Validation failed");
            }

            // 2. Register
            await apiClient.registerService(serviceConfig);

            // Success
            setStep("success");

        } catch (e: any) {
            toast({
                variant: "destructive",
                title: "Connection Failed",
                description: e.message
            });
            setStep("configuration"); // Go back
        } finally {
            setLoading(false);
        }
    };

    return (
        <div className="flex flex-col items-center justify-center min-h-[60vh] p-4 max-w-5xl mx-auto">
            <div className="text-center mb-10 space-y-2">
                <h1 className="text-4xl font-bold tracking-tight bg-gradient-to-r from-primary to-primary/60 bg-clip-text text-transparent">
                    Welcome to MCP Any
                </h1>
                <p className="text-lg text-muted-foreground max-w-lg mx-auto">
                    Your universal gateway for Model Context Protocol tools. Connect your first server to get started.
                </p>
            </div>

            <AnimatePresence mode="wait">
                {step === "selection" && (
                    <motion.div
                        key="selection"
                        initial={{ opacity: 0, y: 20 }}
                        animate={{ opacity: 1, y: 0 }}
                        exit={{ opacity: 0, y: -20 }}
                        className="grid grid-cols-1 md:grid-cols-3 gap-6 w-full"
                    >
                        <SelectionCard
                            icon={Zap}
                            title="Quick Start"
                            description="Connect a demo Weather Service (wttr.in) instantly."
                            badge="Recommended"
                            onClick={() => handleSelect("quick")}
                        />
                        <SelectionCard
                            icon={Terminal}
                            title="Local Command"
                            description="Connect a CLI tool running on your machine via stdio."
                            onClick={() => handleSelect("local")}
                        />
                        <SelectionCard
                            icon={Server}
                            title="Remote Server"
                            description="Connect to an existing MCP server via HTTP (SSE)."
                            onClick={() => handleSelect("remote")}
                        />
                    </motion.div>
                )}

                {step === "configuration" && (
                    <motion.div
                        key="configuration"
                        initial={{ opacity: 0, x: 20 }}
                        animate={{ opacity: 1, x: 0 }}
                        className="w-full max-w-md"
                    >
                        <Card className="border-2 border-primary/10 shadow-xl backdrop-blur-xl bg-background/80">
                            <CardHeader>
                                <CardTitle>Configure {selectedType === "quick" ? "Demo Server" : selectedType === "local" ? "Local Command" : "Remote Server"}</CardTitle>
                                <CardDescription>
                                    {selectedType === "quick"
                                        ? "We will connect to the public wttr.in API."
                                        : "Enter connection details below."}
                                </CardDescription>
                            </CardHeader>
                            <CardContent className="space-y-4">
                                <div className="space-y-2">
                                    <Label>Service Name</Label>
                                    <Input
                                        value={config.name}
                                        onChange={(e) => setConfig({...config, name: e.target.value})}
                                        placeholder="my-service"
                                    />
                                </div>

                                {selectedType === "local" && (
                                    <div className="space-y-2">
                                        <Label>Command</Label>
                                        <Input
                                            value={config.command}
                                            onChange={(e) => setConfig({...config, command: e.target.value})}
                                            placeholder="npx -y @modelcontextprotocol/server-memory"
                                        />
                                        <p className="text-xs text-muted-foreground">The command to start the MCP server.</p>
                                    </div>
                                )}

                                {(selectedType === "remote" || selectedType === "quick") && (
                                    <div className="space-y-2">
                                        <Label>URL</Label>
                                        <Input
                                            value={config.url}
                                            onChange={(e) => setConfig({...config, url: e.target.value})}
                                            placeholder="http://localhost:8080/sse"
                                            disabled={selectedType === "quick"}
                                        />
                                    </div>
                                )}
                            </CardContent>
                            <CardFooter className="flex justify-between">
                                <Button variant="ghost" onClick={() => setStep("selection")}>Back</Button>
                                <Button onClick={handleConnect} disabled={loading}>
                                    {loading ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : null}
                                    Connect <ArrowRight className="ml-2 h-4 w-4" />
                                </Button>
                            </CardFooter>
                        </Card>
                    </motion.div>
                )}

                {step === "validation" && (
                    <motion.div
                        key="validation"
                        initial={{ opacity: 0, scale: 0.95 }}
                        animate={{ opacity: 1, scale: 1 }}
                        className="w-full max-w-md text-center"
                    >
                         <Card className="border-none shadow-none bg-transparent">
                            <CardContent className="flex flex-col items-center gap-4 py-10">
                                <div className="h-12 w-12 rounded-full border-4 border-primary/30 border-t-primary animate-spin" />
                                <p className="text-muted-foreground">Verifying connection & registering...</p>
                            </CardContent>
                         </Card>
                    </motion.div>
                )}

                {step === "success" && (
                    <motion.div
                        key="success"
                        initial={{ opacity: 0, scale: 0.95 }}
                        animate={{ opacity: 1, scale: 1 }}
                        className="w-full max-w-md text-center"
                    >
                         <Card className="border-green-500/20 bg-green-500/5 shadow-2xl">
                            <CardContent className="flex flex-col items-center gap-4 py-10">
                                <div className="h-16 w-16 rounded-full bg-green-500/10 flex items-center justify-center text-green-500 mb-2">
                                    <CheckCircle2 className="h-8 w-8" />
                                </div>
                                <h2 className="text-2xl font-bold">Connected Successfully!</h2>
                                <p className="text-muted-foreground mb-4">Your MCP server is now active.</p>
                                <Button size="lg" className="w-full" onClick={onComplete}>
                                    Go to Dashboard
                                </Button>
                            </CardContent>
                         </Card>
                    </motion.div>
                )}
            </AnimatePresence>
        </div>
    );
}

function SelectionCard({ icon: Icon, title, description, badge, onClick }: { icon: any, title: string, description: string, badge?: string, onClick: () => void }) {
    return (
        <Card
            className="cursor-pointer hover:border-primary/50 hover:shadow-lg transition-all duration-300 group relative overflow-hidden"
            onClick={onClick}
        >
            <div className="absolute inset-0 bg-gradient-to-br from-primary/5 to-transparent opacity-0 group-hover:opacity-100 transition-opacity" />
            <CardHeader>
                <div className="flex justify-between items-start mb-2">
                    <div className="p-3 rounded-xl bg-primary/10 text-primary group-hover:scale-110 transition-transform">
                        <Icon className="h-6 w-6" />
                    </div>
                    {badge && <Badge variant="secondary" className="bg-primary/10 text-primary hover:bg-primary/20">{badge}</Badge>}
                </div>
                <CardTitle className="text-xl">{title}</CardTitle>
                <CardDescription className="line-clamp-2 mt-2">{description}</CardDescription>
            </CardHeader>
            <CardFooter>
                <div className="text-sm font-medium text-primary flex items-center opacity-0 -translate-x-2 group-hover:opacity-100 group-hover:translate-x-0 transition-all">
                    Select <ArrowRight className="ml-1 h-3 w-3" />
                </div>
            </CardFooter>
        </Card>
    );
}
