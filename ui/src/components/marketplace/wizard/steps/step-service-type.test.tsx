/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { StepServiceType } from './step-service-type';
import { useWizard } from '../wizard-context';
import { vi, describe, it, expect, beforeEach } from 'vitest';
import { SERVICE_REGISTRY } from '@/lib/service-registry';

// Mock dependencies
vi.mock('../wizard-context', () => ({
    useWizard: vi.fn(),
}));

vi.mock('@/components/ui/card', () => ({
    Card: ({ children }: any) => <div>{children}</div>,
    CardHeader: ({ children }: any) => <div>{children}</div>,
    CardTitle: ({ children }: any) => <div>{children}</div>,
    CardDescription: ({ children }: any) => <div>{children}</div>,
}));

vi.mock('@/lib/service-registry', () => ({
    SERVICE_REGISTRY: [
        {
            id: 'test-service',
            name: 'Test Service',
            description: 'Test Description',
            command: 'test-command',
            configurationSchema: {
                properties: {
                    TEST_VAR: {
                        type: 'string',
                        default: 'default-value'
                    }
                }
            }
        }
    ]
}));

describe('StepServiceType', () => {
    beforeEach(() => {
        vi.resetAllMocks();
    });

    it('renders templates from registry', () => {
        const updateState = vi.fn();
        const updateConfig = vi.fn();
        const state = { config: {}, selectedTemplateId: null };

        (useWizard as any).mockReturnValue({
            state,
            updateState,
            updateConfig
        });

        render(<StepServiceType />);

        // Open select
        const trigger = screen.getByRole('combobox');
        fireEvent.click(trigger);

        expect(screen.getByText('Test Service')).toBeDefined();
        expect(screen.getAllByText('Manual / Custom').length).toBeGreaterThan(0);
    });

    it('populates defaults when template is selected', async () => {
        const updateState = vi.fn();
        const updateConfig = vi.fn();
        const state = { config: {}, selectedTemplateId: null };

        (useWizard as any).mockReturnValue({
            state,
            updateState,
            updateConfig
        });

        render(<StepServiceType />);

        // Open select
        const trigger = screen.getByRole('combobox');
        fireEvent.click(trigger);

        // Click the item
        const item = screen.getByText('Test Service');
        fireEvent.click(item);

        expect(updateState).toHaveBeenCalledWith(expect.objectContaining({
            selectedTemplateId: 'test-service',
            params: { TEST_VAR: 'default-value' }
        }));

        expect(updateConfig).toHaveBeenCalledWith(expect.objectContaining({
            commandLineService: expect.objectContaining({
                command: 'test-command'
            }),
            configurationSchema: expect.stringContaining('TEST_VAR')
        }));
    });
});
