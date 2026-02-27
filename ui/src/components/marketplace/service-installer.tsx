/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import React, { useState } from "react";
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription, DialogFooter } from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { UpstreamServiceConfig, apiClient } from "@/lib/client";
import { UniversalSchemaForm as SchemaForm } from "@/components/shared/universal-schema-form";
import { useToast } from "@/hooks/use-toast";
import { Package, CheckCircle2, ArrowRight, Loader2, ArrowLeft } from "lucide-react";
import { useRouter } from "next/navigation";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";

interface ServiceInstallerProps {
    open: boolean;
    onOpenChange: (open: boolean) => void;
    templateConfig?: UpstreamServiceConfig;
    onComplete: () => void;
}

enum InstallStep {
    OVERVIEW = 0,
    CONFIGURE = 1,
    INSTALLING = 2,
    SUCCESS = 3
}

/**
 * ServiceInstaller component.
 * A wizard for installing services from templates.
 *
 * @param props - The component props.
 * @returns The rendered component.
 */
export function ServiceInstaller({ open, onOpenChange, templateConfig, onComplete }: ServiceInstallerProps) {
    const [step, setStep] = useState<InstallStep>(InstallStep.OVERVIEW);
    const [configValues, setConfigValues] = useState<any>({});
    const [installError, setInstallError] = useState<string | null>(null);
    const { toast } = useToast();
    const router = useRouter();

    const handleNext = () => {
        if (step === InstallStep.OVERVIEW) {
            setStep(InstallStep.CONFIGURE);
        } else if (step === InstallStep.CONFIGURE) {
            handleInstall();
        }
    };

    const handleBack = () => {
        if (step > InstallStep.OVERVIEW) {
            setStep(step - 1);
        }
    };

    const handleInstall = async () => {
        if (!templateConfig) return;
        setStep(InstallStep.INSTALLING);
        setInstallError(null);

        try {
            const newConfig = { ...templateConfig };
            // Generate a unique ID and Name
            const id = (newConfig.name + "-" + Math.random().toString(36).substring(7)).toLowerCase();
            newConfig.id = id;
            newConfig.name = id;

            // Apply configuration values if it's a CLI service
            if (newConfig.commandLineService) {
                // Initialize env if missing
                const env = { ...(newConfig.commandLineService.env || {}) };
                Object.entries(configValues).forEach(([key, value]) => {
                    // Assuming values map directly to env vars for now as per schema convention
                    // In a more complex setup, we might need a mapping in the template
                    env[key] = { plainText: String(value), validationRegex: "" };
                });
                newConfig.commandLineService.env = env;
            }

            await apiClient.registerService(newConfig);
            setStep(InstallStep.SUCCESS);
            onComplete();
        } catch (e: any) {
            setInstallError(e.message || "Failed to install service.");
            setStep(InstallStep.CONFIGURE); // Go back to config on error
            toast({
                title: "Installation Failed",
                description: e.message,
                variant: "destructive"
            });
        }
    };

    const renderContent = () => {
        if (!templateConfig) return null;

        switch (step) {
            case InstallStep.OVERVIEW:
                return (
                    <div className="space-y-4 py-4">
                        <div className="flex items-center gap-4">
                            <div className="p-3 bg-primary/10 rounded-lg">
                                <Package className="h-8 w-8 text-primary" />
                            </div>
                            <div>
                                <h3 className="text-lg font-semibold">{templateConfig.name}</h3>
                                <p className="text-sm text-muted-foreground">{templateConfig.description}</p>
                            </div>
                        </div>

                        <div className="bg-muted/50 p-4 rounded-md border text-sm space-y-2">
                            <h4 className="font-medium">Capabilities</h4>
                            <ul className="list-disc list-inside text-muted-foreground">
                                {templateConfig.toolCount ? (
                                    <li>Pre-configured tools available</li>
                                ) : (
                                    <li>Dynamic tool discovery enabled</li>
                                )}
                                <li>Secure environment variable management</li>
                                <li>Automatic health checks</li>
                            </ul>
                        </div>
                    </div>
                );

            case InstallStep.CONFIGURE:
                return (
                    <div className="space-y-4 py-4">
                        <div className="space-y-2">
                            <h3 className="font-medium">Configuration</h3>
                            <p className="text-sm text-muted-foreground">
                                Configure the service settings.
                            </p>
                        </div>
                        {installError && (
                            <Alert variant="destructive">
                                <AlertTitle>Error</AlertTitle>
                                <AlertDescription>{installError}</AlertDescription>
                            </Alert>
                        )}
                        <div className="border rounded-md p-4">
                            {templateConfig.configurationSchema ? (
                                <SchemaForm
                                    schema={JSON.parse(templateConfig.configurationSchema)}
                                    value={configValues}
                                    onChange={setConfigValues}
                                />
                            ) : (
                                <div className="text-center py-8 text-muted-foreground">
                                    No configuration required.
                                </div>
                            )}
                        </div>
                    </div>
                );

            case InstallStep.INSTALLING:
                return (
                    <div className="flex flex-col items-center justify-center py-12 space-y-4">
                        <Loader2 className="h-12 w-12 animate-spin text-primary" />
                        <div className="text-center">
                            <h3 className="font-medium">Installing Service...</h3>
                            <p className="text-sm text-muted-foreground">Registering service and verifying connection.</p>
                        </div>
                    </div>
                );

            case InstallStep.SUCCESS:
                return (
                    <div className="flex flex-col items-center justify-center py-8 space-y-6">
                        <div className="h-16 w-16 bg-green-100 dark:bg-green-900/30 rounded-full flex items-center justify-center">
                            <CheckCircle2 className="h-8 w-8 text-green-600 dark:text-green-400" />
                        </div>
                        <div className="text-center space-y-2">
                            <h3 className="text-xl font-semibold">Installation Complete!</h3>
                            <p className="text-muted-foreground">
                                <b>{templateConfig.name}</b> has been successfully installed and is ready to use.
                            </p>
                        </div>
                        <Button onClick={() => {
                            onOpenChange(false);
                            // Navigate to the newly created service (using the ID we generated or just list)
                            // Since we don't have the exact ID here easily without refactor, just go to list
                            router.push('/upstream-services');
                        }} className="w-full">
                            Go to Service
                            <ArrowRight className="ml-2 h-4 w-4" />
                        </Button>
                    </div>
                );
        }
    };

    return (
        <Dialog open={open} onOpenChange={onOpenChange}>
            <DialogContent className="sm:max-w-[600px]">
                <DialogHeader>
                    <DialogTitle>Install Service</DialogTitle>
                    <DialogDescription>
                        Set up a new instance of {templateConfig?.name}.
                    </DialogDescription>
                </DialogHeader>

                <div className="min-h-[300px]">
                    {renderContent()}
                </div>

                {step !== InstallStep.INSTALLING && step !== InstallStep.SUCCESS && (
                    <DialogFooter className="flex justify-between sm:justify-between">
                        <Button variant="ghost" onClick={step === InstallStep.OVERVIEW ? () => onOpenChange(false) : handleBack}>
                            {step === InstallStep.OVERVIEW ? "Cancel" : <><ArrowLeft className="mr-2 h-4 w-4" /> Back</>}
                        </Button>
                        <Button onClick={handleNext}>
                            {step === InstallStep.CONFIGURE ? "Install" : "Next"} <ArrowRight className="ml-2 h-4 w-4" />
                        </Button>
                    </DialogFooter>
                )}
            </DialogContent>
        </Dialog>
    );
}
