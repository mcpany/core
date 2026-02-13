/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useEffect, useState } from "react";
import { apiClient } from "@/lib/client";
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Copy, Plus, ArrowRight, Check, Terminal, ExternalLink, Cloud, Server } from "lucide-react";
import Link from "next/link";
import { useToast } from "@/hooks/use-toast";
import { AddWidgetSheet } from "@/components/dashboard/add-widget-sheet";

/**
 * OnboardingCard component.
 * Guides the user through the first steps of setting up MCP Any.
 * @param props.onAddWidget - Callback to add a widget (to complete the onboarding visually).
 */
export function OnboardingCard({ onAddWidget }: { onAddWidget?: (type: string) => void }) {
    const [step, setStep] = useState<1 | 2>(1);
    const [loading, setLoading] = useState(true);
    const [baseUrl, setBaseUrl] = useState("");
    const { toast } = useToast();

    useEffect(() => {
        setBaseUrl(window.location.origin);
        checkServices();
    }, []);

    const checkServices = async () => {
        try {
            const services = await apiClient.listServices();
            if (services.length > 0) {
                setStep(2);
            }
        } catch (e) {
            console.error("Failed to check services", e);
        } finally {
            setLoading(false);
        }
    };

    const copyToClipboard = (text: string) => {
        navigator.clipboard.writeText(text);
        toast({ title: "Copied", description: "Configuration copied to clipboard." });
    };

    const claudeConfig = JSON.stringify({
        "mcpServers": {
            "mcp-any": {
                "command": "npx",
                "args": ["-y", "@modelcontextprotocol/server-sse-client", "--url", `${baseUrl}/sse`]
            }
        }
    }, null, 2);

    const cursorConfig = `// Add to your Cursor settings or project config
// Cursor supports MCP via command or SSE.
// For SSE:
${baseUrl}/sse`;

    const cliCommand = `npx @modelcontextprotocol/client-cli ${baseUrl}/sse`;

    if (loading) {
        return <div className="animate-pulse h-64 bg-muted/20 rounded-lg border-2 border-dashed"></div>;
    }

    return (
        <Card className="w-full max-w-4xl mx-auto border-2 border-primary/20 shadow-lg backdrop-blur-sm bg-background/80">
            <CardHeader>
                <div className="flex items-center justify-between">
                    <div>
                        <CardTitle className="text-2xl">Welcome to MCP Any</CardTitle>
                        <CardDescription>Let's get your AI agents connected in two simple steps.</CardDescription>
                    </div>
                    <div className="flex items-center gap-2 text-sm text-muted-foreground">
                        <span className={step >= 1 ? "text-primary font-bold" : ""}>1. Services</span>
                        <ArrowRight className="h-4 w-4" />
                        <span className={step >= 2 ? "text-primary font-bold" : ""}>2. Connect Client</span>
                    </div>
                </div>
            </CardHeader>
            <CardContent>
                {step === 1 ? (
                    <div className="space-y-6">
                        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                            <div className="bg-muted/30 p-6 rounded-lg border border-dashed flex flex-col items-center text-center space-y-4 hover:border-primary/50 transition-colors">
                                <div className="p-3 bg-primary/10 rounded-full">
                                    <Cloud className="h-8 w-8 text-primary" />
                                </div>
                                <div>
                                    <h3 className="font-semibold text-lg">Marketplace</h3>
                                    <p className="text-sm text-muted-foreground mt-1">Browse verified community servers.</p>
                                </div>
                                <Button asChild className="w-full">
                                    <Link href="/marketplace">Browse Marketplace</Link>
                                </Button>
                            </div>

                            <div className="bg-muted/30 p-6 rounded-lg border border-dashed flex flex-col items-center text-center space-y-4 hover:border-primary/50 transition-colors">
                                <div className="p-3 bg-secondary/10 rounded-full">
                                    <Server className="h-8 w-8 text-secondary-foreground" />
                                </div>
                                <div>
                                    <h3 className="font-semibold text-lg">Add Custom Service</h3>
                                    <p className="text-sm text-muted-foreground mt-1">Connect a local script or API.</p>
                                </div>
                                <Button asChild variant="secondary" className="w-full">
                                    <Link href="/upstream-services">Add Service</Link>
                                </Button>
                            </div>
                        </div>
                        <p className="text-center text-sm text-muted-foreground">
                            You need at least one running service to power your agents.
                        </p>
                    </div>
                ) : (
                    <div className="space-y-6">
                        <div className="bg-green-500/10 border border-green-500/20 p-4 rounded-md flex items-center gap-3">
                            <Check className="h-5 w-5 text-green-500" />
                            <p className="text-sm">Great! You have active services running. Now connect your client.</p>
                        </div>

                        <Tabs defaultValue="claude" className="w-full">
                            <TabsList className="grid w-full grid-cols-3">
                                <TabsTrigger value="claude">Claude Desktop</TabsTrigger>
                                <TabsTrigger value="cursor">Cursor</TabsTrigger>
                                <TabsTrigger value="cli">MCP CLI</TabsTrigger>
                            </TabsList>
                            <TabsContent value="claude" className="space-y-4 mt-4">
                                <div className="space-y-2">
                                    <p className="text-sm text-muted-foreground">
                                        Add this to your <code>claude_desktop_config.json</code>:
                                    </p>
                                    <div className="relative">
                                        <pre className="bg-muted/50 p-4 rounded-md overflow-x-auto text-xs font-mono border">
                                            {claudeConfig}
                                        </pre>
                                        <Button
                                            size="sm"
                                            variant="ghost"
                                            className="absolute top-2 right-2 h-8 w-8 p-0"
                                            onClick={() => copyToClipboard(claudeConfig)}
                                        >
                                            <Copy className="h-4 w-4" />
                                        </Button>
                                    </div>
                                    <p className="text-xs text-muted-foreground">
                                        Location: <code className="bg-muted px-1 rounded">~/Library/Application Support/Claude/</code> (macOS)
                                    </p>
                                </div>
                            </TabsContent>
                            <TabsContent value="cursor" className="space-y-4 mt-4">
                                <div className="space-y-2">
                                    <p className="text-sm text-muted-foreground">
                                        Use this SSE URL in Cursor MCP settings:
                                    </p>
                                    <div className="relative">
                                        <div className="bg-muted/50 p-4 rounded-md text-sm font-mono border flex items-center justify-between">
                                            <span>{baseUrl}/sse</span>
                                            <Button
                                                size="sm"
                                                variant="ghost"
                                                className="h-8 w-8 p-0"
                                                onClick={() => copyToClipboard(`${baseUrl}/sse`)}
                                            >
                                                <Copy className="h-4 w-4" />
                                            </Button>
                                        </div>
                                    </div>
                                </div>
                            </TabsContent>
                            <TabsContent value="cli" className="space-y-4 mt-4">
                                <div className="space-y-2">
                                    <p className="text-sm text-muted-foreground">
                                        Run this command to test via CLI:
                                    </p>
                                    <div className="relative">
                                        <div className="bg-muted/50 p-4 rounded-md text-xs font-mono border flex items-center justify-between">
                                            <span>{cliCommand}</span>
                                            <Button
                                                size="sm"
                                                variant="ghost"
                                                className="h-8 w-8 p-0"
                                                onClick={() => copyToClipboard(cliCommand)}
                                            >
                                                <Copy className="h-4 w-4" />
                                            </Button>
                                        </div>
                                    </div>
                                </div>
                            </TabsContent>
                        </Tabs>
                    </div>
                )}
            </CardContent>
            <CardFooter className="flex justify-between border-t p-6 bg-muted/5">
                <Button variant="ghost" asChild>
                    <Link href="https://modelcontextprotocol.io/introduction" target="_blank">
                        <ExternalLink className="mr-2 h-4 w-4" /> Docs
                    </Link>
                </Button>
                {step === 2 && onAddWidget && (
                    <div className="flex gap-2">
                         <AddWidgetSheet
                            onAdd={onAddWidget}
                            trigger={
                                <Button variant="outline">
                                    <Plus className="mr-2 h-4 w-4" /> Add Widgets
                                </Button>
                            }
                        />
                    </div>
                )}
                {step === 1 && (
                    <Button variant="ghost" onClick={() => setStep(2)}>
                        Skip to Connection
                    </Button>
                )}
            </CardFooter>
        </Card>
    );
}
