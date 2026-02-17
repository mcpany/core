/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { render, screen, fireEvent } from '@testing-library/react';
import { StepParameters } from './step-parameters';
import { WizardProvider, WizardStep } from '../wizard-context';
import React from 'react';
import { vi } from 'vitest';

// Mock SERVICE_REGISTRY
vi.mock('@/lib/service-registry', () => ({
    SERVICE_REGISTRY: []
}));

vi.mock('../wizard-context', async (importOriginal) => {
    const actual = await importOriginal();
    return {
        ...actual,
        useWizard: vi.fn(),
    };
});

import { useWizard } from '../wizard-context';

describe('StepParameters', () => {
    it('renders manual editor when no schema', () => {
        (useWizard as any).mockReturnValue({
            state: {
                config: { commandLineService: {} },
                params: {}
            },
            updateState: vi.fn(),
            updateConfig: vi.fn()
        });

        render(<StepParameters />);

        expect(screen.getByText('Add Environment Variable')).toBeInTheDocument();
        expect(screen.queryByText('Smart Configuration')).not.toBeInTheDocument();
    });

    it('renders schema form when schema present', () => {
        const schema = JSON.stringify({
            properties: {
                API_KEY: { type: 'string', title: 'API Key' }
            }
        });

        (useWizard as any).mockReturnValue({
            state: {
                config: {
                    commandLineService: {},
                    configurationSchema: schema
                },
                params: {}
            },
            updateState: vi.fn(),
            updateConfig: vi.fn()
        });

        render(<StepParameters />);

        expect(screen.getByText('Smart Configuration')).toBeInTheDocument();
        // Label text might include * or be inside a label element
        expect(screen.getByText('API Key')).toBeInTheDocument();
    });
});
