/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import React, { createContext, useContext, useState, ReactNode, useEffect } from 'react';
import { UpstreamServiceConfig } from '@/lib/client';

/**
 * Defines the sequence of steps in the configuration wizard.
 */
export enum WizardStep {
    SERVICE_TYPE = 0,
    PARAMETERS = 1,
    WEBHOOKS = 2,
    AUTH = 3,
    REVIEW = 4,
}

/**
 * Defines the comprehensive state of the configuration wizard, including current step,
 * the service configuration being built, and temporary UI state.
 */
export interface WizardState {
    /** The current active step in the wizard flow. */
    currentStep: WizardStep;
    /** The partial service configuration being assembled. */
    config: Partial<UpstreamServiceConfig>;
    /** ID of the selected template if one was chosen. */
    selectedTemplateId?: string;
    /** Key-Value pairs for parameters or environment variables collected during the process. */
    params: Record<string, string>;
    /** List of webhooks to be configured. */
    webhooks: any[]; // TODO: Define webhook type
    /** List of transformers to be applied. */
    transformers: any[];
    /** The selected authentication mode (local config or creating a new credential). */
    authType?: 'local' | 'new';
    /** ID of the credential if selected. */
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
 * Manages the state of the configuration wizard and provides methods to navigate
 * between steps and update the configuration.
 *
 * It handles:
 * - State initialization and hydration from sessionStorage.
 * - Validation logic for each step.
 * - Navigation (Next/Prev).
 *
 * @param props - The component props.
 * @param props.children - The child components.
 * @returns The rendered WizardProvider.
 *
 * @remarks
 * Side Effects:
 * - Reads from `sessionStorage` on mount to restore state.
 * - Writes to `sessionStorage` on state changes to persist progress.
 */
export function WizardProvider({ children }: { children: ReactNode }) {
    const [state, setState] = useState<WizardState>(defaultState);
    const [isHydrated, setIsHydrated] = useState(false);

    // Initial hydration
    useEffect(() => {
        const saved = sessionStorage.getItem('wizard_state');
        if (saved) {
            try {
                setState(JSON.parse(saved));
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
 * Hook to access the wizard context.
 *
 * @returns The wizard context containing state and navigation methods.
 * @throws {Error} If used outside of a WizardProvider.
 */
export function useWizard() {
    const context = useContext(WizardContext);
    if (!context) {
        throw new Error('useWizard must be used within a WizardProvider');
    }
    return context;
}
