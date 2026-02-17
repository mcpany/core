/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { vi } from 'vitest';
import { WizardProvider, useWizard } from './wizard-context';
import { StepServiceType } from './steps/step-service-type';
import { StepParameters } from './steps/step-parameters';

// Mock SERVICE_REGISTRY
vi.mock('@/lib/service-registry', () => ({
  SERVICE_REGISTRY: [
    {
      id: 'test-postgres',
      name: 'Test PostgreSQL',
      description: 'A test database service',
      command: 'npx test-postgres',
      configurationSchema: {
        type: 'object',
        properties: {
          DB_URL: {
            type: 'string',
            title: 'Database URL',
            default: 'postgres://localhost:5432/test'
          },
          READ_ONLY: {
            type: 'boolean',
            title: 'Read Only',
            default: false,
            description: 'Enable read-only mode'
          }
        },
        required: ['DB_URL']
      }
    }
  ]
}));

// Helper component to debug state
function StateDebugger() {
  const { state } = useWizard();
  return (
    <div data-testid="debug-state">
      <span data-testid="config-schema">{state.config.configurationSchema}</span>
      <span data-testid="params-db-url">{state.params.DB_URL}</span>
    </div>
  );
}

describe('Create Config Wizard Integration', () => {
  it('loads dynamic templates and renders schema form', async () => {
    const TestComponent = () => (
      <WizardProvider>
        <StepServiceType />
        <StepParameters />
        <StateDebugger />
      </WizardProvider>
    );

    render(<TestComponent />);

    // 1. Check if the dynamic template is in the list
    const selectTrigger = screen.getByRole('combobox');
    fireEvent.click(selectTrigger);

    const option = await screen.findByText('Test PostgreSQL');
    expect(option).toBeInTheDocument();

    // 2. Select the template
    fireEvent.click(option);

    // 3. Verify that schema and default params are loaded
    await waitFor(() => {
       const schemaElem = screen.getByTestId('config-schema');
       expect(schemaElem.textContent).toContain('DB_URL');
       const paramElem = screen.getByTestId('params-db-url');
       expect(paramElem.textContent).toBe('postgres://localhost:5432/test');
    });

    // 4. Verify SchemaForm renders
    // It should show "Database URL" label from schema title
    expect(screen.getByText('Database URL')).toBeInTheDocument();
    expect(screen.getByDisplayValue('postgres://localhost:5432/test')).toBeInTheDocument();

    // 5. Verify Boolean field rendering (Checkbox)
    expect(screen.getByText('Enable read-only mode')).toBeInTheDocument();
  });
});
