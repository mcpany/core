/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import React from 'react';
import { render, screen } from '@testing-library/react';
import { CreateConfigWizard } from './create-config-wizard';
import { vi } from 'vitest';

// Mock SERVICE_REGISTRY
vi.mock('@/lib/service-registry', () => ({
  SERVICE_REGISTRY: [
    {
      id: 'test-service',
      name: 'Test Service',
      repo: 'test/repo',
      command: 'npx -y test-service',
      description: 'A test service',
      configurationSchema: {
        type: 'object',
        properties: {
          TEST_VAR: {
            type: 'string',
            title: 'Test Variable',
            default: 'default-value'
          }
        }
      }
    }
  ]
}));

// Mock API Client
vi.mock('@/lib/client', () => ({
  apiClient: {
    saveTemplate: vi.fn(),
  },
}));

// Mock Toast
vi.mock('@/hooks/use-toast', () => ({
  useToast: () => ({
    toast: vi.fn(),
  }),
}));

describe('CreateConfigWizard', () => {
  it('renders and defaults to Manual template', async () => {
    const onComplete = vi.fn();
    const onOpenChange = vi.fn();

    render(<CreateConfigWizard open={true} onOpenChange={onOpenChange} onComplete={onComplete} />);

    // Check if wizard opens
    expect(screen.getByText('Create Upstream Service Config')).toBeInTheDocument();

    // "Manual / Custom" appears in the Select trigger and in the Card below it.
    const manualElements = screen.getAllByText('Manual / Custom');
    expect(manualElements.length).toBeGreaterThan(0);

    // Verify the description of the Manual template is visible
    expect(screen.getByText('Configure everything from scratch.')).toBeInTheDocument();
  });
});
