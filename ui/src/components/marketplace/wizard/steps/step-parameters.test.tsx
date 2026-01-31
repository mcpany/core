/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { StepParameters } from './step-parameters';
import { WizardProvider, useWizard } from '../wizard-context';
import { vi, describe, it, expect } from 'vitest';

// Component to expose wizard state for assertion
const StateSpy = ({ onStateChange }: { onStateChange: (state: any) => void }) => {
    const { state } = useWizard();
    React.useEffect(() => {
        onStateChange(state);
    }, [state, onStateChange]);
    return null;
};

// Mock ScrollArea to avoid ResizeObserver errors in JSDOM
vi.mock('@/components/ui/scroll-area', () => ({
    ScrollArea: ({ children }: { children: React.ReactNode }) => <div>{children}</div>
}));

describe('StepParameters', () => {
    beforeEach(() => {
        sessionStorage.clear();
    });

    it('adds a parameter and updates config', async () => {
        const handleStateChange = vi.fn();

        render(
            <WizardProvider>
                <StateSpy onStateChange={handleStateChange} />
                <StepParameters />
            </WizardProvider>
        );

        // Click Add Parameter
        fireEvent.click(screen.getByText('Add Parameter'));

        // Input Key (placeholder "VAR_NAME")
        const keyInput = screen.getByPlaceholderText('VAR_NAME');
        fireEvent.change(keyInput, { target: { value: 'TEST_VAR' } });

        // Input Value (placeholder "Value")
        // Note: Initially it is 'plainText' so placeholder is 'Value'
        const valueInput = screen.getByPlaceholderText('Value');
        fireEvent.change(valueInput, { target: { value: 'some-value' } });

        // Check state
        await waitFor(() => {
             const calls = handleStateChange.mock.calls;
             const latestState = calls[calls.length - 1][0];
             expect(latestState.params['TEST_VAR']).toEqual({ type: 'plainText', value: 'some-value' });
             expect(latestState.config.commandLineService.env['TEST_VAR']).toEqual({ plainText: 'some-value' });
        });
    });

    it('sets environment variable type', async () => {
        const handleStateChange = vi.fn();

        render(
            <WizardProvider>
                <StateSpy onStateChange={handleStateChange} />
                <StepParameters />
            </WizardProvider>
        );

        // Click Add Parameter
        fireEvent.click(screen.getByText('Add Parameter'));

        // Input Key
        const keyInput = screen.getByPlaceholderText('VAR_NAME');
        fireEvent.change(keyInput, { target: { value: 'HOST_VAR' } });

        // Change Type (Select)
        // Radix UI Select trigger usually has role 'combobox'
        const triggers = screen.getAllByRole('combobox');
        fireEvent.click(triggers[0]);

        // Option "Host Env Var"
        // Wait for portal content
        await waitFor(() => {
            expect(screen.getByText('Host Env Var')).toBeInTheDocument();
        });
        fireEvent.click(screen.getByText('Host Env Var'));

        // Input Value (placeholder changes to "HOST_VAR_NAME")
        const valueInput = screen.getByPlaceholderText('HOST_VAR_NAME');
        fireEvent.change(valueInput, { target: { value: 'ACTUAL_HOST_VAR' } });

        // Check state
        await waitFor(() => {
             const calls = handleStateChange.mock.calls;
             const latestState = calls[calls.length - 1][0];
             expect(latestState.params['HOST_VAR']).toEqual({ type: 'environmentVariable', value: 'ACTUAL_HOST_VAR' });
             expect(latestState.config.commandLineService.env['HOST_VAR']).toEqual({ environmentVariable: 'ACTUAL_HOST_VAR' });
        });
    });
});
