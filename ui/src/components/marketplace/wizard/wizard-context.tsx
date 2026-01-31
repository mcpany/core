/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import React, { createContext, useContext, useState, ReactNode, useEffect } from 'react';
import { UpstreamServiceConfig } from '@/lib/client';

/**
 * WizardStep defines the sequence of steps in the configuration wizard.
 */
export enum WizardStep {
    SERVICE_TYPE = 0,
    PARAMETERS = 1,
    WEBHOOKS = 2,
    AUTH = 3,
    REVIEW = 4,
}

export interface ParamValue {
    type: 'plainText' | 'environmentVariable';
    value: string;
}

/**
 * WizardState type definition.
 */
export interface WizardState {
    currentStep: WizardStep;
    config: Partial<UpstreamServiceConfig>;
    // Temporary state for the wizard that might not map 1:1 to config yet
    selectedTemplateId?: string;
    params: Record<string, ParamValue>; // Key-Value pairs for parameters/env vars
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
        commandLineService: {
            command: '',
            env: {},
            workingDirectory: ''
        } as any
    },
    params: {},
    webhooks: [],
    transformers: [],
};

const WizardContext = createContext<WizardContextType | undefined>(undefined);

/**
 * WizardProvider manages the state of the configuration wizard and provides
 * methods to navigate between steps and update the configuration.
 * @param props - The component props.
 * @param props.children - The child components.
 * @returns The rendered component.
 */
export function WizardProvider({ children }: { children: ReactNode }) {
    const [state, setState] = useState<WizardState>(defaultState);
    const [isHydrated, setIsHydrated] = useState(false);

    // Initial hydration
    useEffect(() => {
        const saved = sessionStorage.getItem('wizard_state');
        if (saved) {
            try {
                const parsed = JSON.parse(saved);
                // Migration logic: if params are strings, convert to objects
                if (parsed.params) {
                    const migratedParams: Record<string, ParamValue> = {};
                    Object.entries(parsed.params).forEach(([k, v]) => {
                        if (typeof v === 'string') {
                            migratedParams[k] = { type: 'plainText', value: v as string };
                        } else {
                            migratedParams[k] = v as ParamValue;
                        }
                    });
                    parsed.params = migratedParams;
                }
                setState(parsed);
            } catch (e) {
                console.error("Failed to hydrate wizard state", e);
            }
        }
        setIsHydrated(true);
    }, []);

    // Persistence
    useEffect(() => {
        if (isHydrated) {
            sessionStorage.setItem('wizard_state', JSON.stringify(state));
        }
    }, [state, isHydrated]);

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
            console.warn("Validation failed:", validation.error);
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

/**
 * useWizard is a hook to access the wizard context.
 * @returns The wizard context containing state and navigation methods.
 * @throws Error if used outside of a WizardProvider.
 */
export function useWizard() {
    const context = useContext(WizardContext);
    if (!context) {
        throw new Error('useWizard must be used within a WizardProvider');
    }
    return context;
}
