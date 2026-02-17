/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { StepServiceType } from './step-service-type';
import { WizardProvider } from '../wizard-context';
import { SERVICE_REGISTRY } from '@/lib/service-registry';
import { vi, describe, it, expect } from 'vitest';

// Mock dependencies
vi.mock('@/lib/service-registry', () => ({
  SERVICE_REGISTRY: [
    {
      id: 'test-service',
      name: 'Test Service',
      description: 'A test service',
      command: 'npx test-service',
      configurationSchema: {
        properties: {
          TEST_VAR: { type: 'string', default: 'default_val' }
        }
      }
    }
  ]
}));

describe('StepServiceType', () => {
  it('renders templates from service registry', () => {
    render(
      <WizardProvider>
        <StepServiceType />
      </WizardProvider>
    );

    // Open the select
    const trigger = screen.getByRole('combobox');
    fireEvent.click(trigger);

    // Check if test service is in the list
    expect(screen.getByText('Test Service')).toBeInTheDocument();
  });
});
