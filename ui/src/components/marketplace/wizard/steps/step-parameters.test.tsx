/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import React from 'react';
import { render, screen } from '@testing-library/react';
import { StepParameters } from './step-parameters';
import { vi, describe, it, expect, beforeEach } from 'vitest';
import * as WizardContext from '../wizard-context';

// Mock schema form
vi.mock('@/components/marketplace/schema-form', () => ({
  SchemaForm: () => <div data-testid="schema-form">Schema Form</div>
}));

// Mock the entire module
vi.mock('../wizard-context', () => ({
  useWizard: vi.fn()
}));

describe('StepParameters', () => {
  const mockUpdateState = vi.fn();
  const mockUpdateConfig = vi.fn();

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders SchemaForm when configurationSchema is present', () => {
    vi.mocked(WizardContext.useWizard).mockReturnValue({
      state: {
        params: {},
        config: {
          configurationSchema: JSON.stringify({
            properties: { foo: { type: 'string' } }
          }),
          commandLineService: { env: {} }
        },
        currentStep: 1,
        webhooks: [],
        transformers: []
      } as any,
      updateState: mockUpdateState,
      updateConfig: mockUpdateConfig,
      setStep: vi.fn(),
      nextStep: vi.fn(),
      prevStep: vi.fn(),
      reset: vi.fn(),
      validateStep: vi.fn()
    });

    render(<StepParameters />);
    expect(screen.getByTestId('schema-form')).toBeInTheDocument();
  });

  it('renders Key-Value table when configurationSchema is missing', () => {
    vi.mocked(WizardContext.useWizard).mockReturnValue({
      state: {
        params: {},
        config: {
            commandLineService: { env: {} }
        },
        currentStep: 1,
        webhooks: [],
        transformers: []
      } as any,
      updateState: mockUpdateState,
      updateConfig: mockUpdateConfig,
      setStep: vi.fn(),
      nextStep: vi.fn(),
      prevStep: vi.fn(),
      reset: vi.fn(),
      validateStep: vi.fn()
    });

    render(<StepParameters />);
    expect(screen.queryByTestId('schema-form')).not.toBeInTheDocument();
    expect(screen.getByText('Key')).toBeInTheDocument();
    expect(screen.getByText('Value')).toBeInTheDocument();
  });
});
