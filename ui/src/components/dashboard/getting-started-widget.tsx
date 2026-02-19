/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useEffect } from "react";
import { useRouter } from "next/navigation";
import {
    Card,
    CardContent,
    CardDescription,
    CardHeader,
    CardTitle,
    CardFooter
} from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Progress } from "@/components/ui/progress";
import { apiClient, UpstreamServiceConfig } from "@/lib/client";
import { useToast } from "@/hooks/use-toast";
import { Rocket, Plus, ShoppingBag, CheckCircle2, Play, ArrowRight, Loader2 } from "lucide-react";
import { cn } from "@/lib/utils";

/**
 * GettingStartedWidget component.
 * Guides new users through the initial setup process.
 */
export function GettingStartedWidget() {
    const [services, setServices] = useState<UpstreamServiceConfig[]>([]);
    const [loading, setLoading] = useState(true);
    const [installingDemo, setInstallingDemo] = useState(false);
    const router = useRouter();
    const { toast } = useToast();

    // Check system state
    useEffect(() => {
        fetchServices();
    }, []);

    const fetchServices = async () => {
        try {
            const list = await apiClient.listServices();
            setServices(list);
        } catch (e) {
            console.error("Failed to check services", e);
        } finally {
            setLoading(false);
        }
    };

    const hasServices = services.length > 0;
    // We can add more checks later (e.g. hasProfiles, hasRunTool)
    const steps = [
        {
            id: "connect_service",
            title: "Connect a Service",
            description: "Add your first upstream service.",
            completed: hasServices,
            icon: Plus,
            action: () => router.push("/upstream-services")
        },
        {
            id: "browse_marketplace",
            title: "Browse Marketplace",
            description: "Discover community MCP servers.",
            completed: false, // Hard to track if they "browsed", maybe check if they installed from market?
            icon: ShoppingBag,
            action: () => router.push("/marketplace")
        }
    ];

    const progress = (steps.filter(s => s.completed).length / steps.length) * 100;

    const handleInstallDemo = async () => {
        setInstallingDemo(true);
        try {
            const demoConfig: UpstreamServiceConfig = {
                id: "demo-echo",
                name: "demo-echo",
                version: "1.0.0",
                commandLineService: {
                    command: "echo 'Hello World! This is a demo MCP service.'",
                    env: {},
                    workingDirectory: "",
                },
                disable: false,
                priority: 0,
                tags: ["demo", "example"],
            };

            await apiClient.registerService(demoConfig);
            toast({
                title: "Demo Installed",
                description: "The 'demo-echo' service has been created successfully."
            });
            fetchServices();
        } catch (e) {
            console.error("Failed to install demo", e);
            toast({
                variant: "destructive",
                title: "Installation Failed",
                description: "Could not create the demo service."
            });
        } finally {
            setInstallingDemo(false);
        }
    };

    if (loading) {
        return (
            <Card className="h-full flex items-center justify-center min-h-[200px]">
                <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
            </Card>
        );
    }

    // If fully setup, we might want to hide this widget or show "Advanced Tips".
    // For now, let's keep it visible but in a "Success" state if services exist?
    // Or just let the user remove it.

    return (
        <Card className="h-full border-l-4 border-l-primary overflow-hidden relative group">
            <div className="absolute top-0 right-0 p-4 opacity-10 group-hover:opacity-20 transition-opacity">
                <Rocket className="h-32 w-32 -mr-8 -mt-8 rotate-12" />
            </div>

            <CardHeader>
                <div className="flex items-center justify-between">
                    <div>
                        <CardTitle className="text-xl">Welcome to MCP Any</CardTitle>
                        <CardDescription>Let's get your environment set up.</CardDescription>
                    </div>
                    {hasServices && (
                        <div className="flex items-center gap-2 text-green-600 bg-green-50 dark:bg-green-900/20 px-3 py-1 rounded-full text-xs font-medium">
                            <CheckCircle2 className="h-3 w-3" />
                            Active
                        </div>
                    )}
                </div>
            </CardHeader>

            <CardContent className="space-y-6">
                {!hasServices ? (
                    <div className="space-y-4">
                        <p className="text-sm text-muted-foreground">
                            You currently have no services connected. To start using MCP tools, you need to register an upstream service.
                        </p>
                        <div className="grid grid-cols-1 sm:grid-cols-3 gap-3">
                            <Button
                                variant="default"
                                className="h-auto py-3 flex flex-col items-center gap-1"
                                onClick={handleInstallDemo}
                                disabled={installingDemo}
                            >
                                {installingDemo ? <Loader2 className="h-5 w-5 animate-spin" /> : <Play className="h-5 w-5" />}
                                <span className="font-semibold">Quick Start</span>
                                <span className="text-[10px] font-normal opacity-80">Install Demo Service</span>
                            </Button>

                            <Button
                                variant="outline"
                                className="h-auto py-3 flex flex-col items-center gap-1"
                                onClick={() => router.push("/marketplace")}
                            >
                                <ShoppingBag className="h-5 w-5" />
                                <span className="font-semibold">Marketplace</span>
                                <span className="text-[10px] font-normal text-muted-foreground">Browse Community</span>
                            </Button>

                            <Button
                                variant="outline"
                                className="h-auto py-3 flex flex-col items-center gap-1"
                                onClick={() => router.push("/upstream-services")}
                            >
                                <Plus className="h-5 w-5" />
                                <span className="font-semibold">Custom</span>
                                <span className="text-[10px] font-normal text-muted-foreground">Configure Manually</span>
                            </Button>
                        </div>
                    </div>
                ) : (
                    <div className="space-y-4">
                        <div className="flex items-center justify-between text-sm">
                            <span className="font-medium">Setup Progress</span>
                            <span className="text-muted-foreground">Great start!</span>
                        </div>
                        <Progress value={50} className="h-2" />

                        <div className="grid gap-2">
                            <div className="flex items-center justify-between p-3 bg-muted/50 rounded-lg border">
                                <div className="flex items-center gap-3">
                                    <div className="bg-green-100 dark:bg-green-900/30 p-2 rounded-full">
                                        <CheckCircle2 className="h-4 w-4 text-green-600" />
                                    </div>
                                    <div className="flex flex-col">
                                        <span className="text-sm font-medium">Connect Service</span>
                                        <span className="text-xs text-muted-foreground">
                                            {services.length} services active
                                        </span>
                                    </div>
                                </div>
                                <Button variant="ghost" size="sm" onClick={() => router.push("/upstream-services")}>
                                    Manage
                                </Button>
                            </div>

                            <div className="flex items-center justify-between p-3 bg-muted/50 rounded-lg border">
                                <div className="flex items-center gap-3">
                                    <div className="bg-primary/10 p-2 rounded-full">
                                        <Play className="h-4 w-4 text-primary" />
                                    </div>
                                    <div className="flex flex-col">
                                        <span className="text-sm font-medium">Try Playground</span>
                                        <span className="text-xs text-muted-foreground">Test your tools</span>
                                    </div>
                                </div>
                                <Button size="sm" onClick={() => router.push("/playground")}>
                                    Go <ArrowRight className="ml-1 h-3 w-3" />
                                </Button>
                            </div>
                        </div>
                    </div>
                )}
            </CardContent>

            {!hasServices && (
                <CardFooter className="bg-muted/20 border-t p-4">
                    <p className="text-xs text-muted-foreground text-center w-full">
                        Need help? Check the <a href="https://github.com/mcpany/core" target="_blank" rel="noreferrer" className="underline hover:text-primary">documentation</a> or join our community.
                    </p>
                </CardFooter>
            )}
        </Card>
    );
}
