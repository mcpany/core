/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import React from 'react';
import { render, screen, fireEvent, waitFor, within } from '@testing-library/react';
import { CreateConfigWizard } from './create-config-wizard';
import { vi } from 'vitest';

// Mock dependencies
vi.mock('@/lib/client', () => ({
  apiClient: {
    listCredentials: vi.fn().mockResolvedValue([]),
    validateService: vi.fn().mockResolvedValue({ valid: true }),
  },
}));

vi.mock('@/hooks/use-toast', () => ({
  useToast: () => ({
    toast: vi.fn(),
  }),
}));

// Mock Lucide icons
vi.mock('lucide-react', async (importOriginal) => {
  const actual = await importOriginal();
  return {
    ...(actual as object),
  };
});

describe('CreateConfigWizard', () => {
  it('generates correct MCP Stdio configuration for PostgreSQL template', async () => {
    const handleComplete = vi.fn();
    render(<CreateConfigWizard open={true} onOpenChange={() => {}} onComplete={handleComplete} />);

    // Step 1: Select PostgreSQL Template
    const templateSelect = screen.getByRole('combobox');
    fireEvent.click(templateSelect);

    const postgresOption = screen.getByText('PostgreSQL Database');
    fireEvent.click(postgresOption);

    // Verify initial selection description
    expect(screen.getByText('Connect to a PostgreSQL database.')).toBeInTheDocument();

    // Click Next to go to Parameters
    fireEvent.click(screen.getByText('Next'));

    // Step 2: Parameters (This is where we expect the refactor to matter)
    // For now, just click Next to Webhooks
    fireEvent.click(screen.getByText('Next'));

    // Step 3: Webhooks -> Next to Auth
    fireEvent.click(screen.getByText('Next'));

    // Step 4: Auth -> Next to Review
    fireEvent.click(screen.getByText('Next'));

    // Step 5: Review -> Finish
    fireEvent.click(screen.getByText('Finish & Save to Local Marketplace'));

    expect(handleComplete).toHaveBeenCalled();
    const config = handleComplete.mock.calls[0][0];

    // ASSERTION: This should use mcpService with stdioConnection
    // Currently (before fix), it uses commandLineService.
    // We expect the fix to make this true.

    // Check if mcpService exists
    if (!config.mcpService) {
        throw new Error("Expected mcpService to be defined, but got: " + JSON.stringify(config, null, 2));
    }

    if (!config.mcpService.connectionType?.stdioConnection) {
         throw new Error("Expected mcpService.stdioConnection to be defined");
    }

    const stdio = config.mcpService.connectionType.stdioConnection;
    expect(stdio.command).toBe('npx');
    expect(stdio.args).toContain('@modelcontextprotocol/server-postgres');
    expect(stdio.env).toHaveProperty('POSTGRES_URL');
  });
});
