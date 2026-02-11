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
import { Cloud, CheckCircle2, AlertCircle, ArrowRight, Loader2, Download, Plus } from "lucide-react";
import { motion } from "framer-motion";
import { useToast } from "@/hooks/use-toast";
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle, DialogTrigger } from "@/components/ui/dialog";
import { BulkServiceImport } from "@/components/services/bulk-service-import";
import Link from "next/link";

interface OnboardingWizardProps {
    onComplete: () => void;
}

/**
 * OnboardingWizard component.
 * Guides the user through the initial setup process, offering a one-click demo deployment.
 *
 * @param props - Component props.
 * @param props.onComplete - Callback fired when the onboarding is successfully completed.
 */
export function OnboardingWizard({ onComplete }: OnboardingWizardProps) {
    const [step, setStep] = useState<"welcome" | "deploying" | "success" | "error">("welcome");
    const [error, setError] = useState<string | null>(null);
    const { toast } = useToast();

    const handleDeployDemo = async () => {
        setStep("deploying");
        try {
            const template = SERVICE_TEMPLATES.find(t => t.id === "wttrin");
            if (!template) throw new Error("Demo template not found.");

            // Deep clone and ensure defaults
            const config = JSON.parse(JSON.stringify(template.config));

            // Check if service already exists to avoid error
            try {
                const existing = await apiClient.getService(config.name);
                if (existing) {
                    // Already exists, just proceed to success
                    setStep("success");
                    return;
                }
            } catch {
                // Ignore, proceed to register
            }

            console.log("Registering service...", config);
            await apiClient.registerService(config);
            console.log("Service registered!");
            setStep("success");
            toast({
                title: "Demo Deployed",
                description: "Weather service (wttr.in) is now active.",
            });
        } catch (e: any) {
            console.error("Deployment failed", e);
            setError(e.message || "Failed to deploy demo service.");
            setStep("error");
        }
    };

    const containerVariants = {
        hidden: { opacity: 0, y: 20 },
        visible: { opacity: 1, y: 0, transition: { duration: 0.5 } }
    };

    if (step === "success") {
        return (
            <motion.div
                initial="hidden"
                animate="visible"
                variants={containerVariants}
                className="max-w-md mx-auto mt-20"
            >
                <Card className="border-green-500/20 bg-green-500/5">
                    <CardHeader className="text-center">
                        <div className="mx-auto bg-green-100 dark:bg-green-900/30 p-3 rounded-full mb-4 w-fit">
                            <CheckCircle2 className="h-8 w-8 text-green-600 dark:text-green-400" />
                        </div>
                        <CardTitle className="text-2xl">You're All Set!</CardTitle>
                        <CardDescription>
                            Your first MCP service is live and ready to use.
                        </CardDescription>
                    </CardHeader>
                    <CardFooter className="flex flex-col gap-2">
                        <Button className="w-full" onClick={onComplete}>
                            Go to Dashboard <ArrowRight className="ml-2 h-4 w-4" />
                        </Button>
                        <Button variant="ghost" className="w-full" asChild>
                            <Link href="/inspector">
                                Open Inspector
                            </Link>
                        </Button>
                    </CardFooter>
                </Card>
            </motion.div>
        );
    }

    return (
        <motion.div
            initial="hidden"
            animate="visible"
            variants={containerVariants}
            className="max-w-2xl mx-auto mt-12 px-4"
        >
            <div className="text-center mb-8 space-y-2">
                <h1 className="text-4xl font-bold tracking-tight bg-gradient-to-r from-primary to-primary/60 bg-clip-text text-transparent">
                    Welcome to MCP Any
                </h1>
                <p className="text-xl text-muted-foreground">
                    Your enterprise gateway for Model Context Protocol tools.
                </p>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                {/* Primary Path: One-Click Demo */}
                <Card className="border-primary/20 shadow-lg relative overflow-hidden group hover:border-primary/40 transition-colors cursor-pointer md:col-span-2 bg-gradient-to-br from-background to-muted/20" onClick={handleDeployDemo}>
                    <div className="absolute inset-0 bg-primary/5 opacity-0 group-hover:opacity-100 transition-opacity pointer-events-none" />
                    <CardHeader>
                        <div className="flex items-center justify-between">
                            <div className="bg-primary/10 p-2 rounded-lg text-primary w-fit mb-2">
                                <Cloud className="h-6 w-6" />
                            </div>
                            {step === "deploying" && <Loader2 className="h-5 w-5 animate-spin text-primary" />}
                        </div>
                        <CardTitle>Deploy Demo Service</CardTitle>
                        <CardDescription>
                            Instantly deploy <strong>wttr.in</strong> (Weather API) as an MCP tool to see how it works. Zero configuration required.
                        </CardDescription>
                    </CardHeader>
                    <CardFooter>
                         <Button className="w-full group-hover:translate-x-1 transition-transform" disabled={step === "deploying"}>
                            {step === "deploying" ? "Deploying..." : "Start with One Click"}
                            {!step && <ArrowRight className="ml-2 h-4 w-4" />}
                        </Button>
                    </CardFooter>
                </Card>

                {/* Secondary Path: Import */}
                <Dialog>
                    <DialogTrigger asChild>
                        <Card className="hover:bg-muted/50 transition-colors cursor-pointer border-dashed">
                            <CardHeader>
                                <div className="bg-muted p-2 rounded-lg text-muted-foreground w-fit mb-2">
                                    <Download className="h-5 w-5" />
                                </div>
                                <CardTitle className="text-lg">Import Configuration</CardTitle>
                                <CardDescription>
                                    Already have a config or OpenAPI spec?
                                </CardDescription>
                            </CardHeader>
                        </Card>
                    </DialogTrigger>
                    <DialogContent className="sm:max-w-xl">
                        <DialogHeader>
                            <DialogTitle>Import Services</DialogTitle>
                            <DialogDescription>
                                Upload a JSON configuration or import from URL.
                            </DialogDescription>
                        </DialogHeader>
                        <BulkServiceImport
                            onImportSuccess={() => {
                                onComplete();
                                toast({ title: "Import Successful", description: "Services added to dashboard." });
                            }}
                            onCancel={() => {}}
                        />
                    </DialogContent>
                </Dialog>

                {/* Secondary Path: Manual */}
                <Link href="/upstream-services" className="block">
                    <Card className="hover:bg-muted/50 transition-colors cursor-pointer border-dashed h-full">
                        <CardHeader>
                             <div className="bg-muted p-2 rounded-lg text-muted-foreground w-fit mb-2">
                                <Plus className="h-5 w-5" />
                            </div>
                            <CardTitle className="text-lg">Add Manually</CardTitle>
                            <CardDescription>
                                Configure a service from scratch using the editor.
                            </CardDescription>
                        </CardHeader>
                    </Card>
                </Link>
            </div>

            {error && (
                <div className="mt-6 space-y-2">
                    <Alert variant="destructive">
                        <AlertCircle className="h-4 w-4" />
                        <AlertTitle>Error</AlertTitle>
                        <AlertDescription>{error}</AlertDescription>
                    </Alert>
                    <div className="text-center">
                         <Button variant="outline" size="sm" onClick={() => { setError(null); setStep("welcome"); }}>Try Again</Button>
                    </div>
                </div>
            )}
        </motion.div>
    );
}
