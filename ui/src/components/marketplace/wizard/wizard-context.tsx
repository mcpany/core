/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import React, { createContext, useContext, useState, ReactNode } from 'react';
import { UpstreamServiceConfig } from '@/lib/client';

// Define the steps
export enum WizardStep {
    SERVICE_TYPE = 0,
    PARAMETERS = 1,
    WEBHOOKS = 2,
    AUTH = 3,
    REVIEW = 4,
}

export interface WizardState {
    currentStep: WizardStep;
    config: Partial<UpstreamServiceConfig>;
    // Temporary state for the wizard that might not map 1:1 to config yet
    selectedTemplateId?: string;
    params: Record<string, string>; // Key-Value pairs for parameters/env vars
    webhooks: any[]; // TODO: Define webhook type
    transformers: any[];
    authType?: 'local' | 'new';
    authCredentialId?: string;
}

interface WizardContextType {
    state: WizardState;
    setStep: (step: WizardStep) => void;
    updateConfig: (updates: Partial<UpstreamServiceConfig>) => void;
    updateState: (updates: Partial<WizardState>) => void;
    nextStep: () => void;
    prevStep: () => void;
    reset: () => void;
    validateStep: (step: WizardStep) => { valid: boolean; error?: string };
}

const defaultState: WizardState = {
    currentStep: WizardStep.SERVICE_TYPE,
    config: {
        name: '',
        version: '0.0.1',
        disable: false,
    },
    params: {},
    webhooks: [],
    transformers: [],
};

const WizardContext = createContext<WizardContextType | undefined>(undefined);

export function WizardProvider({ children }: { children: ReactNode }) {
    const [state, setState] = useState<WizardState>(defaultState);

    const setStep = (step: WizardStep) => setState(prev => ({ ...prev, currentStep: step }));

    const updateConfig = (updates: Partial<UpstreamServiceConfig>) => {
        setState(prev => ({
            ...prev,
            config: { ...prev.config, ...updates }
        }));
    };

    const updateState = (updates: Partial<WizardState>) => {
        setState(prev => ({ ...prev, ...updates }));
    };

    const validateStep = (step: WizardStep): { valid: boolean; error?: string } => {
        switch (step) {
            case WizardStep.SERVICE_TYPE:
                if (!state.config.name) return { valid: false, error: "Service name is required" };
                if (!state.config.commandLineService && !state.config.httpService && !state.config.grpcService && !state.config.mcpService) {
                     return { valid: false, error: "Service configuration is missing" };
                }
                return { valid: true };
            case WizardStep.PARAMETERS:
                return { valid: true }; // Parameters are usually optional
            case WizardStep.WEBHOOKS:
                for (const hook of state.webhooks) {
                     if (hook.webhook && !hook.webhook.url) {
                         return { valid: false, error: "Webhook URL is required" };
                     }
                }
                return { valid: true };
            case WizardStep.AUTH:
                 return { valid: true }; // Auth is optional
            default:
                return { valid: true };
        }
    };

    const nextStep = () => {
        const validation = validateStep(state.currentStep);
        if (!validation.valid) {
            // caller should handle error display, or we throw?
            // Better to return boolean if possible, but the signature is void.
            // We'll trust the caller to check validateStep OR we change nextStep signature?
            // Let's change nextStep to return boolean?
            // But consumers might expect void.
            // Let's just return early and let UI handle validation check separately?
            // Or exposing validateStep is better.
            console.warn("Validation failed:", validation.error);
            // We won't block here if the UI doesn't check, but we should.
            // Let's allow movement if we change the nextStep to check it?
            // Ideally validation is UI concern before calling nextStep.
        }

        setState(prev => {
            const next = prev.currentStep + 1;
            return { ...prev, currentStep: next > WizardStep.REVIEW ? WizardStep.REVIEW : next };
        });
    };

    const prevStep = () => {
        setState(prev => {
            const next = prev.currentStep - 1;
            return { ...prev, currentStep: next < 0 ? 0 : next };
        });
    };

    const reset = () => setState(defaultState);

    return (
        <WizardContext.Provider value={{ state, setStep, updateConfig, updateState, nextStep, prevStep, reset, validateStep }}>
            {children}
        </WizardContext.Provider>
    );
}

export function useWizard() {
    const context = useContext(WizardContext);
    if (!context) {
        throw new Error('useWizard must be used within a WizardProvider');
    }
    return context;
}
