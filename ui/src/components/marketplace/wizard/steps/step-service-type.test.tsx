/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { render, screen, fireEvent } from '@testing-library/react';
import { StepServiceType } from './step-service-type';
import { WizardProvider } from '../wizard-context';
import React from 'react';
import { vi } from 'vitest';

// Mock SERVICE_REGISTRY to control test data
vi.mock('@/lib/service-registry', () => ({
    SERVICE_REGISTRY: [
        {
            id: 'mock-postgres',
            name: 'Mock Postgres',
            description: 'Mock DB',
            config: {
                commandLineService: {
                    command: 'mock-cmd'
                }
            },
            configurationSchema: {
                properties: {
                    POSTGRES_URL: { default: 'postgres://mock' }
                }
            }
        }
    ]
}));

describe('StepServiceType', () => {
    it('renders static and registry templates', () => {
        render(
            <WizardProvider>
                <StepServiceType />
            </WizardProvider>
        );

        // Check if select trigger is present
        const trigger = screen.getByRole('combobox');
        fireEvent.click(trigger);

        // Use getAllByText for "Manual / Custom" since it appears in trigger, list, and card
        expect(screen.getAllByText('Manual / Custom').length).toBeGreaterThan(0);

        // Registry
        expect(screen.getByText('Mock Postgres')).toBeInTheDocument();
    });

    it('updates description when template selected', () => {
        render(
            <WizardProvider>
                <StepServiceType />
            </WizardProvider>
        );

        const trigger = screen.getByRole('combobox');
        fireEvent.click(trigger);
        const option = screen.getByText('Mock Postgres');
        fireEvent.click(option);

        // Verify description update
        expect(screen.getByText('Mock DB')).toBeInTheDocument();
    });
});
