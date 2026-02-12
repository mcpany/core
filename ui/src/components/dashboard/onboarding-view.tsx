/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import {
    Cpu,
    ArrowRight,
    Globe,
    Plus,
    Loader2,
    Terminal
} from "lucide-react";
import { Card, CardDescription, CardFooter, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { apiClient, UpstreamServiceConfig } from "@/lib/client";
import { useToast } from "@/hooks/use-toast";
import { RegisterServiceDialog } from "@/components/register-service-dialog";

interface OnboardingViewProps {
    onServiceRegistered?: () => void;
}

/**
 * OnboardingView component.
 * displayed when there are no services registered.
 */
export function OnboardingView({ onServiceRegistered }: OnboardingViewProps) {
    const router = useRouter();
    const { toast } = useToast();
    const [installing, setInstalling] = useState<string | null>(null);

    const handleQuickInstall = async (type: "memory" | "filesystem") => {
        setInstalling(type);
        try {
            let config: UpstreamServiceConfig;

            if (type === "memory") {
                config = {
                    name: "memory",
                    id: "memory",
                    version: "1.0.0",
                    disable: false,
                    priority: 0,
                    loadBalancingStrategy: 0,
                    sanitizedName: "memory",
                    readOnly: false,
                    callPolicies: [],
                    preCallHooks: [],
                    postCallHooks: [],
                    prompts: [],
                    autoDiscoverTool: true,
                    configError: "",
                    configurationSchema: "",
                    commandLineService: {
                        command: "npx -y @modelcontextprotocol/server-memory",
                        env: {},
                        workingDirectory: "",
                        local: false,
                        tools: [],
                        resources: [],
                        prompts: [],
                        calls: {},
                        communicationProtocol: 0
                    },
                    tags: ["quick-start", "memory"],
                };
            } else {
                // Filesystem logic would go here
                return;
            }

            await apiClient.registerService(config);

            toast({
                title: "Service Installed",
                description: "Memory server has been registered successfully.",
            });

            if (onServiceRegistered) {
                onServiceRegistered();
            } else {
                window.location.reload();
            }

        } catch (e: any) {
            console.error(e);
            toast({
                variant: "destructive",
                title: "Installation Failed",
                description: e.message || "Failed to register service.",
            });
        } finally {
            setInstalling(null);
        }
    };

    return (
        <div className="flex flex-col items-center justify-center min-h-[80vh] p-4 animate-in fade-in zoom-in duration-500">
            <div className="text-center mb-12 max-w-2xl">
                <div className="flex justify-center mb-6">
                    <div className="h-20 w-20 bg-primary/10 rounded-3xl flex items-center justify-center shadow-inner border border-primary/20">
                        <Terminal className="h-10 w-10 text-primary" />
                    </div>
                </div>
                <h1 className="text-4xl font-bold tracking-tight mb-4">Welcome to MCP Any</h1>
                <p className="text-xl text-muted-foreground">
                    Your enterprise gateway for the Model Context Protocol.
                    Connect your first server to get started.
                </p>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-3 gap-6 w-full max-w-5xl">
                {/* Option 1: Quick Start Memory */}
                <Card className="relative overflow-hidden group hover:border-primary/50 transition-all cursor-pointer border-dashed" onClick={() => handleQuickInstall("memory")}>
                    <div className="absolute inset-0 bg-gradient-to-br from-blue-500/5 to-transparent opacity-0 group-hover:opacity-100 transition-opacity" />
                    <CardHeader>
                        <div className="h-12 w-12 bg-blue-100 dark:bg-blue-900/30 rounded-lg flex items-center justify-center mb-4 text-blue-600 dark:text-blue-400">
                            <Cpu className="h-6 w-6" />
                        </div>
                        <CardTitle>Memory Server</CardTitle>
                        <CardDescription>
                            Ephemeral knowledge graph. Perfect for testing tools and prompts without persistent storage.
                        </CardDescription>
                    </CardHeader>
                    <CardFooter>
                        <Button variant="ghost" className="w-full group-hover:bg-blue-500/10 group-hover:text-blue-600 dark:group-hover:text-blue-400" disabled={!!installing}>
                            {installing === "memory" ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : "One-Click Install"}
                        </Button>
                    </CardFooter>
                </Card>

                {/* Option 2: Browse Marketplace */}
                <Card className="relative overflow-hidden group hover:border-primary/50 transition-all cursor-pointer border-dashed" onClick={() => router.push("/marketplace")}>
                    <div className="absolute inset-0 bg-gradient-to-br from-purple-500/5 to-transparent opacity-0 group-hover:opacity-100 transition-opacity" />
                    <CardHeader>
                        <div className="h-12 w-12 bg-purple-100 dark:bg-purple-900/30 rounded-lg flex items-center justify-center mb-4 text-purple-600 dark:text-purple-400">
                            <Globe className="h-6 w-6" />
                        </div>
                        <CardTitle>Marketplace</CardTitle>
                        <CardDescription>
                            Browse community servers like GitHub, Google Drive, Slack, and PostgreSQL.
                        </CardDescription>
                    </CardHeader>
                    <CardFooter>
                        <Button variant="ghost" className="w-full group-hover:bg-purple-500/10 group-hover:text-purple-600 dark:group-hover:text-purple-400">
                            Browse Catalog <ArrowRight className="ml-2 h-4 w-4" />
                        </Button>
                    </CardFooter>
                </Card>

                {/* Option 3: Connect Manually */}
                <RegisterServiceDialog
                    trigger={
                        <Card className="relative overflow-hidden group hover:border-primary/50 transition-all cursor-pointer h-full border-dashed">
                            <div className="absolute inset-0 bg-gradient-to-br from-green-500/5 to-transparent opacity-0 group-hover:opacity-100 transition-opacity" />
                            <CardHeader>
                                <div className="h-12 w-12 bg-green-100 dark:bg-green-900/30 rounded-lg flex items-center justify-center mb-4 text-green-600 dark:text-green-400">
                                    <Plus className="h-6 w-6" />
                                </div>
                                <CardTitle>Connect Manually</CardTitle>
                                <CardDescription>
                                    Have a local MCP server or remote endpoint? Connect it directly via Wizard.
                                </CardDescription>
                            </CardHeader>
                            <CardFooter className="mt-auto">
                                <Button variant="ghost" className="w-full group-hover:bg-green-500/10 group-hover:text-green-600 dark:group-hover:text-green-400">
                                    Open Wizard
                                </Button>
                            </CardFooter>
                        </Card>
                    }
                    onSuccess={onServiceRegistered}
                />
            </div>

            <div className="mt-12 text-center text-sm text-muted-foreground">
                <p>Not sure where to start? Read the <a href="#" className="underline hover:text-primary">documentation</a> or try the <a href="/playground" className="underline hover:text-primary">Playground</a>.</p>
            </div>
        </div>
    );
}
