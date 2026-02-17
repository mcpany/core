/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import React from 'react';
import { render, screen } from '@testing-library/react';
import { StepServiceType } from './step-service-type';
import { vi, describe, it, expect, beforeEach } from 'vitest';
import * as WizardContext from '../wizard-context';

// Mock the wizard context hook
vi.mock('../wizard-context', () => ({
  useWizard: vi.fn()
}));

// Mock SERVICE_REGISTRY if needed, but it's better to use the real one to catch data issues
// However, if we suspect the real one causes crash, we might want to mock it partially?
// Let's assume real one is fine for now, or jest/vitest will load it.

describe('StepServiceType', () => {
  const mockUpdateState = vi.fn();
  const mockUpdateConfig = vi.fn();

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders successfully', () => {
    vi.mocked(WizardContext.useWizard).mockReturnValue({
      state: {
        config: { name: '' },
        params: {},
        currentStep: 0,
        selectedTemplateId: 'manual', // Default
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

    render(<StepServiceType />);

    // Check if critical elements are present
    expect(screen.getByText('Service Name')).toBeInTheDocument();
    expect(screen.getByText('Template')).toBeInTheDocument();
    // Use getAllByText because "Manual / Custom" appears in the Select trigger and the list/description
    expect(screen.getAllByText('Manual / Custom').length).toBeGreaterThan(0);
  });
});
