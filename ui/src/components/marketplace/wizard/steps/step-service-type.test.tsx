
import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { StepServiceType } from './step-service-type';
import { WizardProvider } from '../wizard-context';
import { SERVICE_REGISTRY } from '@/lib/service-registry';
import { vi, describe, it, expect } from 'vitest';

// Mock SERVICE_REGISTRY
vi.mock('@/lib/service-registry', () => ({
  SERVICE_REGISTRY: [
    {
      id: 'test-service',
      name: 'Test Service',
      description: 'A test service from registry',
      command: 'npx test-service',
      configurationSchema: {
        type: 'object',
        properties: {
          TEST_VAR: {
            type: 'string',
            default: 'default_value'
          }
        }
      }
    }
  ]
}));

describe('StepServiceType', () => {
  it('renders registry templates', () => {
    render(
      <WizardProvider>
        <StepServiceType />
      </WizardProvider>
    );

    // Open select
    const trigger = screen.getByRole('combobox');
    fireEvent.click(trigger);

    // Check if Test Service is in the list
    expect(screen.getByText('Test Service')).toBeInTheDocument();
  });
});
