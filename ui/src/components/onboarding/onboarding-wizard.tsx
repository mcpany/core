/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from "@/components/ui/card";
import { SERVICE_TEMPLATES } from "@/lib/templates";
import { apiClient } from "@/lib/client";
import { useToast } from "@/hooks/use-toast";
import { Loader2, Zap, Settings2, ArrowRight, CheckCircle2 } from "lucide-react";
import { RegisterServiceDialog } from "@/components/register-service-dialog";

interface OnboardingWizardProps {
    onComplete: () => void;
}

export function OnboardingWizard({ onComplete }: OnboardingWizardProps) {
    const [isLoading, setIsLoading] = useState(false);
    const { toast } = useToast();

    const handleQuickStart = async () => {
        setIsLoading(true);
        try {
            // Find templates
            const weatherTemplate = SERVICE_TEMPLATES.find(t => t.id === "wttrin");
            const timeTemplate = SERVICE_TEMPLATES.find(t => t.id === "time");

            if (!weatherTemplate || !timeTemplate) {
                throw new Error("Required templates not found.");
            }

            // Register services sequentially
            // We use generic cast or verify config type matches UpstreamServiceConfig
            await apiClient.registerService({
                ...weatherTemplate.config,
                id: "weather-service", // Explicit ID
            } as any);

            await apiClient.registerService({
                ...timeTemplate.config,
                id: "time-service", // Explicit ID
            } as any);

            toast({
                title: "System Ready",
                description: "Weather and Time services have been deployed successfully.",
            });

            // Small delay for effect
            setTimeout(() => {
                onComplete();
            }, 1000);

        } catch (error: any) {
            console.error(error);
            toast({
                variant: "destructive",
                title: "Setup Failed",
                description: error.message || "Failed to register services.",
            });
            setIsLoading(false);
        }
    };

    return (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-background/80 backdrop-blur-sm p-4">
            <Card className="w-full max-w-lg border-2 shadow-2xl animate-in fade-in zoom-in-95 duration-300">
                <CardHeader className="text-center pb-2">
                    <div className="mx-auto bg-primary/10 p-4 rounded-full w-16 h-16 flex items-center justify-center mb-4">
                        <Zap className="h-8 w-8 text-primary" />
                    </div>
                    <CardTitle className="text-2xl font-bold">Welcome to MCP Any</CardTitle>
                    <CardDescription className="text-base mt-2">
                        The Universal Adapter for Model Context Protocol.
                        <br />
                        Connect your first services to get started.
                    </CardDescription>
                </CardHeader>
                <CardContent className="space-y-4 pt-6">
                    <Button
                        size="lg"
                        className="w-full h-16 text-lg justify-between px-6 group"
                        onClick={handleQuickStart}
                        disabled={isLoading}
                    >
                        <div className="flex items-center gap-3">
                            <div className="bg-white/20 p-2 rounded-full">
                                {isLoading ? <Loader2 className="h-5 w-5 animate-spin" /> : <Zap className="h-5 w-5" />}
                            </div>
                            <div className="flex flex-col items-start">
                                <span className="font-semibold">Quick Start (Demo)</span>
                                <span className="text-xs font-normal opacity-80">Deploys Weather & Time services</span>
                            </div>
                        </div>
                        <ArrowRight className="h-5 w-5 opacity-0 group-hover:opacity-100 transition-opacity transform group-hover:translate-x-1" />
                    </Button>

                    <div className="relative">
                        <div className="absolute inset-0 flex items-center">
                            <span className="w-full border-t" />
                        </div>
                        <div className="relative flex justify-center text-xs uppercase">
                            <span className="bg-background px-2 text-muted-foreground">Or</span>
                        </div>
                    </div>

                    <RegisterServiceDialog
                        trigger={
                            <Button variant="outline" size="lg" className="w-full h-14 justify-between px-6 group">
                                <div className="flex items-center gap-3">
                                    <Settings2 className="h-5 w-5 text-muted-foreground" />
                                    <div className="flex flex-col items-start">
                                        <span className="font-semibold">Manual Configuration</span>
                                        <span className="text-xs font-normal text-muted-foreground">Connect existing tools or APIs</span>
                                    </div>
                                </div>
                                <ArrowRight className="h-5 w-5 opacity-0 group-hover:opacity-100 text-muted-foreground transition-opacity" />
                            </Button>
                        }
                        onSuccess={onComplete}
                    />
                </CardContent>
                <CardFooter className="justify-center pb-6">
                    <p className="text-xs text-muted-foreground">
                        You can always add more services later from the Dashboard.
                    </p>
                </CardFooter>
            </Card>
        </div>
    );
}
