/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import React from 'react';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription, DialogFooter } from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { WizardProvider, useWizard, WizardStep } from './wizard-context';
import { StepServiceType } from './steps/step-service-type';
import { StepParameters } from './steps/step-parameters';
import { StepWebhooks } from './steps/step-webhooks';
import { StepAuth } from './steps/step-auth';
import { StepReview } from './steps/step-review';
import { Separator } from '@/components/ui/separator';

interface CreateConfigWizardProps {
    open: boolean;
    onOpenChange: (open: boolean) => void;
    onComplete: (config: any) => void;
}


import { useToast } from "@/hooks/use-toast";

function WizardContent({ onComplete, onCancel }: { onComplete: (config: any) => void, onCancel: () => void }) {
    const { state, nextStep, prevStep, validateStep } = useWizard();
    const { currentStep: step } = state;
    const { toast } = useToast();


    const renderStep = () => {
        switch (step) {
            case WizardStep.SERVICE_TYPE: return <StepServiceType />;
            case WizardStep.PARAMETERS: return <StepParameters />;
            case WizardStep.WEBHOOKS: return <StepWebhooks />;
            case WizardStep.AUTH: return <StepAuth />;
            case WizardStep.REVIEW: return <StepReview onComplete={onComplete} />;
            default: return null;
        }
    };

    const getStepTitle = () => {
        switch (step) {
            case WizardStep.SERVICE_TYPE: return "1. Select Service Type";
            case WizardStep.PARAMETERS: return "2. Configure Parameters";
            case WizardStep.WEBHOOKS: return "3. Webhooks & Transformers";
            case WizardStep.AUTH: return "4. Authentication";
            case WizardStep.REVIEW: return "5. Review & Finish";
            default: return "";
        }
    };

    const handleNext = () => {
        const validation = validateStep(step);
        if (!validation.valid) {
            // Ideally useToast here, but WizardContent needs access to it.
            // We can just alert for now or expect parent to pass toast?
            // Actually useToast is available in components.
             toast({ title: "Validation Error", description: validation.error, variant: "destructive" });
             return;
        }
        nextStep();
    };

    return (
        <>
            <DialogHeader>
                <DialogTitle>Create Upstream Service Config</DialogTitle>
                <DialogDescription>
                    {getStepTitle()}
                </DialogDescription>
            </DialogHeader>

            <Separator className="my-4" />

            <div className="flex-1 overflow-y-auto py-2 px-1 min-h-[300px] max-h-[60vh]">
                {renderStep()}
            </div>

            <DialogFooter className="mt-4 gap-2">
                 <Button variant="outline" onClick={step === 0 ? onCancel : prevStep}>
                    {step === 0 ? "Cancel" : "Back"}
                </Button>
                {step < WizardStep.REVIEW && (
                    <Button onClick={handleNext}>
                        Next
                    </Button>
                )}
            </DialogFooter>
        </>
    );
}

export function CreateConfigWizard({ open, onOpenChange, onComplete }: CreateConfigWizardProps) {
    return (
        <Dialog open={open} onOpenChange={onOpenChange}>
            <DialogContent className="sm:max-w-[700px] flex flex-col h-[80vh]">
                <WizardProvider>
                     {/* We need a wrapper to use useWizard hook inside */}
                     <WizardContent onComplete={onComplete} onCancel={() => onOpenChange(false)} />
                </WizardProvider>
            </DialogContent>
        </Dialog>
    );
}
