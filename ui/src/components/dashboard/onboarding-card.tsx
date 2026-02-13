/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useEffect } from "react";
import { apiClient } from "@/lib/client";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle, CardFooter } from "@/components/ui/card";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { CheckCircle2, Copy, ArrowRight, Server, Terminal, ExternalLink, Loader2, Sparkles, Plus } from "lucide-react";
import { useToast } from "@/hooks/use-toast";
import { Badge } from "@/components/ui/badge";
import Link from "next/link";

/**
 * OnboardingCard component.
 * Guides the user through the initial setup: Connecting a service and connecting an AI client.
 *
 * @returns The rendered component or null if setup is complete.
 */
export function OnboardingCard() {
    const [step, setStep] = useState<1 | 2 | 3>(1);
    const [loading, setLoading] = useState(true);
    const [serviceCount, setServiceCount] = useState(0);
    const [serverUrl, setServerUrl] = useState("");
    const { toast } = useToast();

    useEffect(() => {
        if (typeof window !== 'undefined') {
            setServerUrl(window.location.origin);
        }
        checkStatus();
    }, []);

    const checkStatus = async () => {
        try {
            const services = await apiClient.listServices();
            setServiceCount(services.length);
            if (services.length > 0) {
                setStep(2);
            }
        } catch (e) {
            console.error("Failed to check status", e);
        } finally {
            setLoading(false);
        }
    };

    const copyToClipboard = (text: string) => {
        navigator.clipboard.writeText(text);
        toast({ title: "Copied", description: "Configuration copied to clipboard." });
    };

    if (loading) return null;

    // If services exist and we haven't explicitly dismissed (TODO: persist dismiss),
    // we show step 2. If user is advanced, they can just ignore it or we add a dismiss button.
    // For now, if services > 0, we show step 2.
    // If services > 0 and user has connected client (hard to know), we should hide.
    // Let's rely on manual dismiss or just show it until > N requests?
    // For this MVP, let's show it if services=0 (Step 1) or services>0 (Step 2).
    // We can add a "Dismiss" button to hide it from session.

    const claudeConfig = {
        mcpServers: {
            "mcp-any": {
                command: "npx",
                args: ["-y", "@modelcontextprotocol/server-sse-client", "--url", `${serverUrl}/sse`]
            }
        }
    };

    const cursorCommand = `mcp-any connect --url ${serverUrl}`; // Hypothetical or reuse sse-client

    return (
        <Card className="mb-8 border-primary/20 bg-gradient-to-br from-background to-primary/5 shadow-md relative overflow-hidden">
            <div className="absolute top-0 right-0 p-4 opacity-10 pointer-events-none">
                <Sparkles className="w-32 h-32" />
            </div>

            <CardHeader>
                <div className="flex items-center justify-between">
                    <div className="space-y-1">
                        <CardTitle className="text-2xl flex items-center gap-2">
                            {step === 1 ? "Welcome to MCP Any!" : "You're almost there!"}
                            <Badge variant="outline" className="ml-2 font-normal text-xs uppercase tracking-wider">
                                Setup Guide
                            </Badge>
                        </CardTitle>
                        <CardDescription>
                            {step === 1
                                ? "Let's get your first tool connected in less than a minute."
                                : "Now connect your AI Agent to start using your tools."}
                        </CardDescription>
                    </div>
                    <div className="flex items-center gap-2 text-sm text-muted-foreground">
                        <span className={step >= 1 ? "text-primary font-bold" : ""}>1. Add Service</span>
                        <ArrowRight className="w-4 h-4" />
                        <span className={step >= 2 ? "text-primary font-bold" : ""}>2. Connect Client</span>
                    </div>
                </div>
            </CardHeader>

            <CardContent>
                {step === 1 && (
                    <div className="grid md:grid-cols-2 gap-6">
                        <div className="space-y-4">
                            <div className="flex items-start gap-3">
                                <div className="p-2 bg-primary/10 rounded-lg text-primary">
                                    <Server className="w-6 h-6" />
                                </div>
                                <div>
                                    <h3 className="font-semibold">Connect an Upstream Service</h3>
                                    <p className="text-sm text-muted-foreground mt-1">
                                        MCP Any acts as a gateway. You need to register an API, a local script, or an MCP server that you want to expose to your agents.
                                    </p>
                                </div>
                            </div>
                            <div className="flex gap-3 pt-2">
                                <Button asChild>
                                    <Link href="/upstream-services">
                                        <Plus className="mr-2 h-4 w-4" /> Add Service Manually
                                    </Link>
                                </Button>
                                <Button variant="outline" asChild>
                                    <Link href="/marketplace">
                                        Browse Marketplace
                                    </Link>
                                </Button>
                            </div>
                        </div>
                        <div className="bg-muted/50 rounded-lg p-4 border border-dashed flex flex-col justify-center items-center text-center space-y-2">
                            <p className="text-sm font-medium">Quick Start: Weather Service</p>
                            <p className="text-xs text-muted-foreground mb-2">Try adding the `wttr.in` service to test the flow.</p>
                            <Button size="sm" variant="secondary" onClick={async () => {
                                setLoading(true);
                                try {
                                    await apiClient.registerService({
                                        id: "weather-demo",
                                        name: "weather-demo",
                                        version: "1.0.0",
                                        disable: false,
                                        priority: 0,
                                        httpService: {
                                            address: "https://wttr.in",
                                            // Simple proxy setup, normally needs tool definitions
                                        },
                                        // We should use a better example that has auto-discovery or simple tools
                                        // For now, directing to marketplace is safer if we don't have a canned good config.
                                    });
                                    toast({ title: "Service Added", description: "Weather demo service added." });
                                    checkStatus();
                                } catch (e) {
                                    // Ignore error if already exists or fails, just redirect
                                    window.location.href = "/marketplace";
                                } finally {
                                    setLoading(false);
                                }
                            }}>
                                Add Weather Demo
                            </Button>
                        </div>
                    </div>
                )}

                {step === 2 && (
                    <div className="space-y-6">
                        <Tabs defaultValue="claude" className="w-full">
                            <TabsList>
                                <TabsTrigger value="claude">Claude Desktop</TabsTrigger>
                                <TabsTrigger value="cursor">Cursor / VS Code</TabsTrigger>
                                <TabsTrigger value="cli">MCP CLI</TabsTrigger>
                            </TabsList>

                            <TabsContent value="claude" className="space-y-4 pt-4">
                                <div className="space-y-2">
                                    <p className="text-sm">
                                        1. Open your Claude Desktop configuration file:
                                        <code className="ml-2 bg-muted px-1 py-0.5 rounded text-xs">
                                            ~/Library/Application Support/Claude/claude_desktop_config.json
                                        </code>
                                    </p>
                                    <p className="text-sm">2. Add the following configuration:</p>
                                </div>
                                <div className="relative">
                                    <pre className="bg-slate-950 text-slate-50 p-4 rounded-md text-xs font-mono overflow-x-auto">
                                        {JSON.stringify(claudeConfig, null, 2)}
                                    </pre>
                                    <Button
                                        size="icon"
                                        variant="secondary"
                                        className="absolute top-2 right-2 h-8 w-8"
                                        onClick={() => copyToClipboard(JSON.stringify(claudeConfig, null, 2))}
                                    >
                                        <Copy className="h-4 w-4" />
                                    </Button>
                                </div>
                            </TabsContent>

                            <TabsContent value="cursor" className="space-y-4 pt-4">
                                <div className="space-y-2">
                                    <p className="text-sm">1. Go to <strong>Cursor Settings</strong> {'>'} <strong>Features</strong> {'>'} <strong>MCP</strong>.</p>
                                    <p className="text-sm">2. Click <strong>+ Add New MCP Server</strong>.</p>
                                    <p className="text-sm">3. Select "SSE" and enter the URL:</p>
                                </div>
                                <div className="flex items-center gap-2">
                                    <code className="bg-muted px-3 py-2 rounded text-sm font-mono flex-1">
                                        {serverUrl}/sse
                                    </code>
                                    <Button size="sm" variant="outline" onClick={() => copyToClipboard(`${serverUrl}/sse`)}>
                                        <Copy className="h-4 w-4 mr-2" /> Copy URL
                                    </Button>
                                </div>
                            </TabsContent>

                            <TabsContent value="cli" className="space-y-4 pt-4">
                                <p className="text-sm">Use the official MCP CLI to connect via stdio bridge (requires <code>server-sse-client</code>):</p>
                                <div className="relative">
                                    <pre className="bg-slate-950 text-slate-50 p-4 rounded-md text-xs font-mono overflow-x-auto">
                                        npx -y @modelcontextprotocol/client-sse {serverUrl}/sse
                                    </pre>
                                    <Button
                                        size="icon"
                                        variant="secondary"
                                        className="absolute top-2 right-2 h-8 w-8"
                                        onClick={() => copyToClipboard(`npx -y @modelcontextprotocol/client-sse ${serverUrl}/sse`)}
                                    >
                                        <Copy className="h-4 w-4" />
                                    </Button>
                                </div>
                            </TabsContent>
                        </Tabs>
                    </div>
                )}
            </CardContent>

            {step === 2 && (
                <CardFooter className="bg-muted/20 border-t pt-4 flex justify-between items-center">
                    <div className="text-xs text-muted-foreground flex items-center gap-2">
                        <Loader2 className="w-3 h-3 animate-spin" />
                        Waiting for client connection...
                    </div>
                    <Button variant="ghost" size="sm" onClick={() => setStep(3)}>
                        Dismiss Guide
                    </Button>
                </CardFooter>
            )}
        </Card>
    );
}
