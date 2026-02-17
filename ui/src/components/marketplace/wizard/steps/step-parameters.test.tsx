/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { StepParameters } from './step-parameters';
import { WizardProvider, useWizard } from '../wizard-context';
import { vi, describe, it, expect, beforeEach } from 'vitest';

// Mock dependencies
vi.mock('../wizard-context', () => ({
    useWizard: vi.fn(),
    WizardProvider: ({ children }: any) => <div>{children}</div>
}));

// Mock SchemaForm to isolate testing
vi.mock('@/components/marketplace/schema-form', () => ({
    SchemaForm: ({ value, onChange, schema }: any) => (
        <div data-testid="schema-form">
            <input
                data-testid="schema-input"
                value={value.test_key || ''}
                onChange={(e) => onChange({ ...value, test_key: e.target.value })}
            />
            <div data-testid="schema-json">{JSON.stringify(schema)}</div>
        </div>
    )
}));

describe('StepParameters', () => {
    beforeEach(() => {
        vi.resetAllMocks();
    });

    it('renders key-value editor when no schema is present', () => {
        // Mock useWizard
        const updateState = vi.fn();
        const updateConfig = vi.fn();
        const state = {
            params: { "EXISTING_VAR": "value" },
            config: { commandLineService: { command: "test-cmd" } }
        };

        (useWizard as any).mockReturnValue({
            state,
            updateState,
            updateConfig
        });

        render(<StepParameters />);

        expect(screen.getByText('Environment Variables / Parameters')).toBeDefined();
        expect(screen.getByDisplayValue('EXISTING_VAR')).toBeDefined();
        expect(screen.getByDisplayValue('value')).toBeDefined();
        expect(screen.queryByTestId('schema-form')).toBeNull();
    });

    it('renders SchemaForm when schema is present', () => {
        // Mock useWizard
        const updateState = vi.fn();
        const updateConfig = vi.fn();
        const schema = { properties: { test_key: { type: "string" } } };
        const state = {
            params: { "test_key": "initial" },
            config: {
                commandLineService: { command: "test-cmd" },
                configurationSchema: JSON.stringify(schema)
            }
        };

        (useWizard as any).mockReturnValue({
            state,
            updateState,
            updateConfig
        });

        render(<StepParameters />);

        expect(screen.getByTestId('schema-form')).toBeDefined();
        expect(screen.getByDisplayValue('initial')).toBeDefined();
        expect(screen.getByText('Configuration')).toBeDefined();
    });

    it('updates params and config when SchemaForm changes', () => {
        // Mock useWizard
        const updateState = vi.fn();
        const updateConfig = vi.fn();
        const schema = { properties: { test_key: { type: "string" } } };
        const state = {
            params: {},
            config: {
                commandLineService: { command: "test-cmd", env: {} },
                configurationSchema: JSON.stringify(schema)
            }
        };

        (useWizard as any).mockReturnValue({
            state,
            updateState,
            updateConfig
        });

        render(<StepParameters />);

        const input = screen.getByTestId('schema-input');
        fireEvent.change(input, { target: { value: 'new-value' } });

        expect(updateState).toHaveBeenCalledWith({ params: { test_key: 'new-value' } });
        // It also calls updateConfig to sync env
         expect(updateConfig).toHaveBeenCalledWith({
            commandLineService: {
                command: "test-cmd",
                env: { test_key: { plainText: "new-value" } }
            }
        });
    });
});
