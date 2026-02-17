/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { CreateConfigWizard } from './create-config-wizard';
import { vi } from 'vitest';

// Mock dependencies
vi.mock('@/lib/client', () => ({
  apiClient: {
    saveTemplate: vi.fn(),
    listCredentials: vi.fn().mockResolvedValue([]),
  },
}));

vi.mock('@/hooks/use-toast', () => ({
    useToast: () => ({
      toast: vi.fn(),
    }),
  }));

// Mock SERVICE_REGISTRY
vi.mock('@/lib/service-registry', () => ({
  SERVICE_REGISTRY: [
    {
      id: "test-db",
      name: "Test DB",
      repo: "test/repo",
      command: "test-command",
      description: "Test Description",
      configurationSchema: {
        type: "object",
        properties: {
          "DB_URL": { type: "string", title: "Database URL", default: "default-url" }
        }
      }
    }
  ]
}));

// ResizeObserver mock
global.ResizeObserver = class ResizeObserver {
    observe() {}
    unobserve() {}
    disconnect() {}
};

describe('CreateConfigWizard', () => {
  it('renders and allows selecting a registry template', async () => {
    const onComplete = vi.fn();
    render(<CreateConfigWizard open={true} onOpenChange={() => {}} onComplete={onComplete} />);

    // Check if "Service Name" input is present (Step 1)
    expect(screen.getByLabelText(/Service Name/)).toBeInTheDocument();

    // Select the template via the select component
    // Select component in shadcn/ui is complex to test with fireEvent directly on the trigger sometimes
    // But let's try opening it.
    const trigger = screen.getByRole('combobox');
    fireEvent.click(trigger);

    // Select "Test DB"
    const option = await screen.findByText('Test DB');
    fireEvent.click(option);

    // Click Next
    fireEvent.click(screen.getByText('Next'));

    // Step 2: Parameters
    // Should show SchemaForm for DB_URL
    // Wait for the form to appear
    await waitFor(() => expect(screen.getByText('Database URL')).toBeInTheDocument());

    // Check default value
    const input = screen.getByLabelText('Database URL') as HTMLInputElement;
    expect(input.value).toBe('default-url');

    // Fill DB_URL
    fireEvent.change(input, { target: { value: 'postgres://localhost' } });

    // Click Next
    fireEvent.click(screen.getByText('Next'));

    // Step 3: Webhooks (skip)
    expect(screen.getByText('3. Webhooks & Transformers')).toBeInTheDocument();
    fireEvent.click(screen.getByText('Next'));

    // Step 4: Auth (skip)
    expect(screen.getByText('4. Authentication')).toBeInTheDocument();
    fireEvent.click(screen.getByText('Next'));

    // Step 5: Review
    expect(screen.getByText(/Review & Finish/)).toBeInTheDocument();

    // Click Finish
    fireEvent.click(screen.getByText('Finish & Save to Local Marketplace'));

    await waitFor(() => expect(onComplete).toHaveBeenCalled());

    // Verify the config passed to onComplete has the env var
    const config = onComplete.mock.calls[0][0];
    // commandLineService.env structure depends on step-parameters logic
    // which maps params to { plainText: v }
    expect(config.commandLineService.env["DB_URL"]).toEqual({ plainText: "postgres://localhost" });
    // Verify command is populated
    expect(config.commandLineService.command).toBe("test-command");
  });
});
