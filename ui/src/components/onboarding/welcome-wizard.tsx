/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useEffect } from "react";
import { ArrowRight, Check, CheckCircle2, Copy, Link as LinkIcon, Plus, Server, Terminal, Zap } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from "@/components/ui/card";
import { RegisterServiceDialog } from "@/components/register-service-dialog";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Input } from "@/components/ui/input";
import { JsonView } from "@/components/ui/json-view";
import { useToast } from "@/hooks/use-toast";
import Link from "next/link";
import { cn } from "@/lib/utils";

interface WelcomeWizardProps {
    /**
     * Callback when the wizard is completed.
     */
    onComplete: () => void;
}

/**
 * WelcomeWizard component.
 * Displays a multi-step onboarding wizard for new users.
 * Uses CSS animations instead of framer-motion to avoid extra dependencies.
 *
 * @param props - The component props.
 * @returns The rendered component.
 */
export function WelcomeWizard({ onComplete }: WelcomeWizardProps) {
    const [step, setStep] = useState<"intro" | "register" | "connect">("intro");
    const { toast } = useToast();

    const nextStep = () => {
        if (step === "intro") setStep("register");
        else if (step === "register") setStep("connect");
        else if (step === "connect") onComplete();
    };

    return (
        <div className="flex flex-col items-center justify-center min-h-[80vh] p-4 max-w-4xl mx-auto">
            {step === "intro" && (
                <div
                    key="intro"
                    className="text-center space-y-6 max-w-2xl animate-in fade-in slide-in-from-bottom-4 duration-500 fill-mode-forwards"
                >
                    <div className="flex justify-center mb-6">
                        <div className="bg-primary/10 p-4 rounded-full">
                            <Zap className="h-12 w-12 text-primary" />
                        </div>
                    </div>
                    <h1 className="text-4xl font-bold tracking-tight">Welcome to MCP Any</h1>
                    <p className="text-xl text-muted-foreground">
                        The definitive management console for the Model Context Protocol.
                        <br />
                        Turn your APIs into AI tools in seconds.
                    </p>
                    <div className="grid grid-cols-1 md:grid-cols-3 gap-4 text-left mt-8">
                        <FeatureCard icon={Server} title="Universal Gateway" description="Connect HTTP, gRPC, and CLI tools." />
                        <FeatureCard icon={Terminal} title="Zero-Config CLI" description="Standardized stdio for any client." />
                        <FeatureCard icon={CheckCircle2} title="Enterprise Grade" description="Auth, Policy, and Audit logging." />
                    </div>
                    <div className="pt-8">
                        <Button size="lg" onClick={nextStep} className="gap-2">
                            Get Started <ArrowRight className="h-4 w-4" />
                        </Button>
                    </div>
                </div>
            )}

            {step === "register" && (
                <div
                    key="register"
                    className="w-full max-w-lg animate-in fade-in slide-in-from-right-4 duration-500 fill-mode-forwards"
                >
                    <Card className="border-2 border-primary/20 shadow-2xl">
                        <CardHeader>
                            <CardTitle>Step 1: Register a Service</CardTitle>
                            <CardDescription>
                                Connect your first tool or API to MCP Any.
                            </CardDescription>
                        </CardHeader>
                        <CardContent className="space-y-4">
                            <p className="text-sm text-muted-foreground">
                                You can start with a simple local command, or connect to a remote API.
                            </p>
                            <div className="flex flex-col gap-2">
                                <RegisterServiceDialog
                                    onSuccess={() => {
                                        toast({ title: "Service Connected", description: "Great job! Now let's connect your AI client." });
                                        nextStep();
                                    }}
                                    trigger={
                                        <Button size="lg" className="w-full gap-2">
                                            <Plus className="h-4 w-4" /> Register New Service
                                        </Button>
                                    }
                                />
                                <div className="relative py-2">
                                    <div className="absolute inset-0 flex items-center">
                                        <span className="w-full border-t" />
                                    </div>
                                    <div className="relative flex justify-center text-xs uppercase">
                                        <span className="bg-background px-2 text-muted-foreground">Or</span>
                                    </div>
                                </div>
                                <Button variant="outline" className="w-full" asChild>
                                    <Link href="/marketplace?wizard=open">
                                        Browse Marketplace Templates
                                    </Link>
                                </Button>
                            </div>
                        </CardContent>
                        <CardFooter className="justify-between">
                            <Button variant="ghost" onClick={() => setStep("intro")}>Back</Button>
                            <Button variant="ghost" onClick={nextStep} className="text-muted-foreground">Skip for now</Button>
                        </CardFooter>
                    </Card>
                </div>
            )}

            {step === "connect" && (
                <div
                    key="connect"
                    className="w-full max-w-2xl animate-in fade-in slide-in-from-right-4 duration-500 fill-mode-forwards"
                >
                    <Card className="border-2 border-primary/20 shadow-2xl">
                        <CardHeader>
                            <CardTitle>Step 2: Connect AI Client</CardTitle>
                            <CardDescription>
                                Configure your favorite AI editor to use MCP Any.
                            </CardDescription>
                        </CardHeader>
                        <CardContent>
                            <ClientConnectTabs />
                        </CardContent>
                        <CardFooter className="justify-between">
                            <Button variant="ghost" onClick={() => setStep("register")}>Back</Button>
                            <Button onClick={onComplete} size="lg" className="gap-2">
                                Go to Dashboard <Check className="h-4 w-4" />
                            </Button>
                        </CardFooter>
                    </Card>
                </div>
            )}
        </div>
    );
}

/**
 * FeatureCard component.
 * Displays a feature highlight in the wizard.
 *
 * @param props - The component props.
 * @returns The rendered component.
 */
// eslint-disable-next-line @typescript-eslint/no-explicit-any
function FeatureCard({ icon: Icon, title, description }: { icon: any, title: string, description: string }) {
    return (
        <div className="p-4 rounded-lg bg-muted/50 border flex flex-col gap-2 hover:bg-muted/70 transition-colors">
            <Icon className="h-6 w-6 text-primary" />
            <h3 className="font-semibold">{title}</h3>
            <p className="text-sm text-muted-foreground">{description}</p>
        </div>
    );
}

/**
 * ClientConnectTabs component.
 * Displays connection instructions for different clients.
 *
 * @returns The rendered component.
 */
function ClientConnectTabs() {
    const [origin, setOrigin] = useState("");
    const { toast } = useToast();

    // Use window location on mount
    useEffect(() => {
        if (typeof window !== "undefined") {
            setOrigin(window.location.origin);
        }
    }, []);

    const displayUrl = origin || "http://localhost:50050";
    const sseUrl = `${displayUrl}/sse`;

    const copyToClipboard = (text: string) => {
        if (navigator.clipboard && navigator.clipboard.writeText) {
            navigator.clipboard.writeText(text).catch(() => {});
            toast({ title: "Copied", description: "Config copied to clipboard" });
        }
    };

    const claudeConfig = {
        "mcpServers": {
            "mcp-any": {
                "command": "npx",
                "args": ["-y", "@modelcontextprotocol/server-sse-client", "--url", sseUrl]
            }
        }
    };

    return (
        <Tabs defaultValue="claude" className="w-full">
            <TabsList className="grid w-full grid-cols-3">
                <TabsTrigger value="claude">Claude Desktop</TabsTrigger>
                <TabsTrigger value="cursor">Cursor</TabsTrigger>
                <TabsTrigger value="cli">Gemini / CLI</TabsTrigger>
            </TabsList>

            <TabsContent value="claude" className="space-y-4 pt-4">
                <div className="space-y-2">
                    <p className="text-sm text-muted-foreground">
                        Add this to your <code>claude_desktop_config.json</code>:
                    </p>
                    <div className="relative">
                        <JsonView data={claudeConfig} />
                        <Button size="icon" variant="outline" className="absolute top-2 right-2 h-6 w-6" onClick={() => copyToClipboard(JSON.stringify(claudeConfig, null, 2))}>
                            <Copy className="h-3 w-3" />
                        </Button>
                    </div>
                </div>
            </TabsContent>

            <TabsContent value="cursor" className="space-y-4 pt-4">
                <div className="space-y-2 text-sm">
                    <ol className="list-decimal pl-4 space-y-2 text-muted-foreground">
                        <li>Open Cursor Settings <span className="kbd bg-muted px-1 rounded text-xs">Cmd+,</span></li>
                        <li>Go to <strong>Features &gt; MCP</strong></li>
                        <li>Click <strong>+ Add New MCP Server</strong></li>
                        <li>Type: <strong>SSE</strong></li>
                        <li>URL: <code className="bg-muted px-1 rounded">{sseUrl}</code></li>
                    </ol>
                    <div className="flex items-center gap-2 mt-4">
                        <Input readOnly value={sseUrl} className="font-mono text-xs" />
                        <Button size="icon" variant="outline" onClick={() => copyToClipboard(sseUrl)}>
                            <Copy className="h-4 w-4" />
                        </Button>
                    </div>
                </div>
            </TabsContent>

            <TabsContent value="cli" className="space-y-4 pt-4">
                <div className="space-y-2">
                    <p className="text-sm text-muted-foreground">Run this command:</p>
                    <div className="relative group">
                        <pre className="text-xs font-mono bg-muted p-4 rounded-md overflow-x-auto border">
                            gemini mcp add --transport http --trust mcpany {displayUrl}
                        </pre>
                        <Button
                            variant="ghost"
                            size="icon"
                            className="absolute right-2 top-2 h-6 w-6"
                            onClick={() => copyToClipboard(`gemini mcp add --transport http --trust mcpany ${displayUrl}`)}
                        >
                            <Copy className="h-3 w-3" />
                        </Button>
                    </div>
                </div>
            </TabsContent>
        </Tabs>
    );
}
