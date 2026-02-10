/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState } from "react";
import { Dialog, DialogContent, DialogTrigger } from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Wand2 } from "lucide-react";
import { CatalogStep } from "./steps/catalog-step";
import { ServiceConfigStep } from "./steps/service-config-step";
import { AuthStep } from "./steps/auth-step";
import { ProfileStep } from "./steps/profile-step";
import { toast } from "sonner";
import { apiClient } from "@/lib/client";

/**
 * WizardService represents a service being configured in the wizard.
 */
export interface WizardService {
    templateId: string;
    instanceName: string; // e.g. "my-google-cal"
    config: any; // The upstream service config
    isAuthenticated: boolean;
    credentials?: any;
}

/**
 * WizardDialog is a dialog for creating a new profile using a step-by-step wizard.
 * @param props.onProfileCreated - Callback when profile is created.
 */
export function WizardDialog({ onProfileCreated }: { onProfileCreated: () => void }) {
    const [open, setOpen] = useState(false);
    const [step, setStep] = useState(1);
    const [selectedServices, setSelectedServices] = useState<WizardService[]>([]);

    // Reset state when opening
    const handleOpenChange = (newOpen: boolean) => {
        if (newOpen) {
            setStep(1);
            setSelectedServices([]);
        }
        setOpen(newOpen);
    };

    const nextStep = () => setStep(s => s + 1);
    const prevStep = () => setStep(s => s - 1);

    const handleServicesSelected = (services: WizardService[]) => {
        setSelectedServices(services);
        nextStep();
    };

    const handleConfigComplete = (services: WizardService[]) => {
        setSelectedServices(services);
        nextStep();
    };

    const handleAuthComplete = (services: WizardService[]) => {
        setSelectedServices(services);
        nextStep();
    };

    const handleProfileCreated = (profileName: string) => {
        toast.success(`Profile ${profileName} created successfully!`);
        setOpen(false);
        onProfileCreated();
    };

    return (
        <Dialog open={open} onOpenChange={handleOpenChange}>
            <DialogTrigger asChild>
                <Button variant="outline" className="gap-2">
                    <Wand2 className="h-4 w-4" />
                    Profile Wizard
                </Button>
            </DialogTrigger>
            <DialogContent className="max-w-4xl h-[80vh] flex flex-col p-0 gap-0">
                <div className="p-6 border-b">
                    <h2 className="text-2xl font-semibold tracking-tight">Profile Wizard</h2>
                    <p className="text-sm text-muted-foreground">
                        Step {step} of 4: {
                            step === 1 ? "Select Services" :
                            step === 2 ? "Configure Services" :
                            step === 3 ? "Authenticate" :
                            "Finalize Profile"
                        }
                    </p>
                </div>

                <div className="flex-1 overflow-y-auto p-6">
                    {step === 1 && (
                        <CatalogStep onNext={handleServicesSelected} />
                    )}
                    {step === 2 && (
                        <ServiceConfigStep
                            services={selectedServices}
                            onNext={handleConfigComplete}
                            onBack={prevStep}
                        />
                    )}
                    {step === 3 && (
                        <AuthStep
                            services={selectedServices}
                            onNext={handleAuthComplete}
                            onBack={prevStep}
                        />
                    )}
                    {step === 4 && (
                        <ProfileStep
                            services={selectedServices}
                            onBack={prevStep}
                            onComplete={handleProfileCreated}
                        />
                    )}
                </div>
            </DialogContent>
        </Dialog>
    );
}
