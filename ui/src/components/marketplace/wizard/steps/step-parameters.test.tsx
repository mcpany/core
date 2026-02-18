
import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { StepParameters } from './step-parameters';
import { WizardProvider, WizardStep, WizardState, WizardContextType } from '../wizard-context';
import { UpstreamServiceConfig } from '@/lib/client';
import { vi, describe, it, expect } from 'vitest';

// Mock WizardContext to inject specific config
const mockWizardContextValue: WizardContextType = {
  state: {
    currentStep: WizardStep.PARAMETERS,
    config: {
      name: 'Test Service',
      configurationSchema: JSON.stringify({
        type: 'object',
        properties: {
          TEST_VAR: {
            type: 'string',
            title: 'Test Variable',
            default: 'default_value'
          }
        },
        required: ['TEST_VAR']
      }),
      commandLineService: {
        command: 'npx test',
        env: {}
      }
    } as Partial<UpstreamServiceConfig>,
    params: { TEST_VAR: 'default_value' },
    webhooks: [],
    transformers: [],
  } as WizardState,
  setStep: vi.fn(),
  updateConfig: vi.fn(),
  updateState: vi.fn(),
  nextStep: vi.fn(),
  prevStep: vi.fn(),
  reset: vi.fn(),
  validateStep: vi.fn()
};

// Plan B: Mock `useWizard` directly.
vi.mock('../wizard-context', async () => {
    const originalModule = await vi.importActual('../wizard-context') as any;
    return {
        ...originalModule,
        useWizard: () => mockWizardContextValue
    };
});

describe('StepParameters', () => {
  it('renders SchemaForm when schema is present', () => {
    render(<StepParameters />);

    // Should see "Configuration" header from schema mode
    expect(screen.getByText('Configuration')).toBeInTheDocument();

    // Should see the field from schema
    expect(screen.getByLabelText(/Test Variable/)).toBeInTheDocument();
    expect(screen.getByDisplayValue('default_value')).toBeInTheDocument();
  });

  it('updates params when input changes', () => {
    render(<StepParameters />);

    const input = screen.getByLabelText(/Test Variable/);
    fireEvent.change(input, { target: { value: 'new_value' } });

    expect(mockWizardContextValue.updateState).toHaveBeenCalledWith({
      params: expect.objectContaining({ TEST_VAR: 'new_value' })
    });

    // Also checks syncEnv was called (implicit in implementation, tested via updateConfig)
    expect(mockWizardContextValue.updateConfig).toHaveBeenCalledWith(expect.objectContaining({
        commandLineService: expect.objectContaining({
            env: expect.objectContaining({
                TEST_VAR: { plainText: 'new_value' }
            })
        })
    }));
  });
});
