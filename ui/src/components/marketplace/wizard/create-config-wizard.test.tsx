/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { CreateConfigWizard } from './create-config-wizard';
import { vi, describe, it, expect } from 'vitest';

// Mock dependencies
vi.mock('@/hooks/use-toast', () => ({
  useToast: () => ({
    toast: vi.fn(),
  }),
}));

// Mock Lucide icons to avoid issues with dynamic imports or missing icons in test env
vi.mock('lucide-react', async (importOriginal) => {
  const actual = await importOriginal();
  return {
    ...(actual as object),
  };
});

// Mock ResizeObserver for Resizable panels if needed (Wizard doesn't use them but good practice)
global.ResizeObserver = class ResizeObserver {
  observe() {}
  unobserve() {}
  disconnect() {}
};

describe('CreateConfigWizard Integration', () => {
  it('selects PostgreSQL template and generates McpService config', async () => {
    const onComplete = vi.fn();
    const onOpenChange = vi.fn();

    render(<CreateConfigWizard open={true} onOpenChange={onOpenChange} onComplete={onComplete} />);

    // Step 1: Select PostgreSQL
    // Wait for dialog to open
    await waitFor(() => {
        expect(screen.getByText('1. Select Service Type')).toBeInTheDocument();
    });

    // Open the Select dropdown
    const trigger = screen.getByRole('combobox');
    fireEvent.click(trigger);

    // Select PostgreSQL
    const postgresOption = await screen.findByText('PostgreSQL Database');
    fireEvent.click(postgresOption);

    // Click Next
    const nextButton = screen.getByText('Next');
    fireEvent.click(nextButton);

    // Step 2: Parameters
    await waitFor(() => {
        expect(screen.getByText('2. Configure Parameters')).toBeInTheDocument();
    });

    // Check if command is populated
    // Note: In current implementation (before refactor), StepParameters renders "Command" label and input.
    // In new implementation, it should also render "Command".
    const commandInput = screen.getByDisplayValue(/npx/);
    expect(commandInput).toBeInTheDocument();

    // Proceed to Review
    fireEvent.click(screen.getByText('Next')); // To Webhooks
    await waitFor(() => expect(screen.getByText('3. Webhooks & Transformers')).toBeInTheDocument());

    fireEvent.click(screen.getByText('Next')); // To Auth
    await waitFor(() => expect(screen.getByText('4. Authentication')).toBeInTheDocument());

    fireEvent.click(screen.getByText('Next')); // To Review
    await waitFor(() => expect(screen.getByText('5. Review & Finish')).toBeInTheDocument());

    // Finish
    fireEvent.click(screen.getByText('Finish & Save to Local Marketplace'));

    await waitFor(() => {
        expect(onComplete).toHaveBeenCalled();
    });

    const config = onComplete.mock.calls[0][0];

    // Assertions for expected structure (Refactored)
    // It SHOULD be mcpService.stdioConnection
    expect(config.mcpService).toBeDefined();
    expect(config.mcpService.stdioConnection).toBeDefined();
    expect(config.mcpService.stdioConnection.command).toContain('npx');
    expect(config.mcpService.stdioConnection.args).toBeDefined();

    // It should NOT be commandLineService
    expect(config.commandLineService).toBeUndefined();
  });
});
