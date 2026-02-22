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
    listTemplates: vi.fn().mockResolvedValue([]),
    saveTemplate: vi.fn().mockResolvedValue({}),
  },
}));

vi.mock('@/hooks/use-toast', () => ({
  useWizard: () => ({
    state: { config: { name: 'test' } },
    updateConfig: vi.fn(),
  }),
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

// Mock ResizeObserver
global.ResizeObserver = class ResizeObserver {
  observe() {}
  unobserve() {}
  disconnect() {}
};

describe('CreateConfigWizard Flow', () => {
  it('generates correct MCP Stdio configuration for PostgreSQL template', async () => {
    const onComplete = vi.fn();
    const onOpenChange = vi.fn();

    render(<CreateConfigWizard open={true} onOpenChange={onOpenChange} onComplete={onComplete} />);

    // Step 1: Service Type
    expect(screen.getByText('1. Select Service Type')).toBeInTheDocument();

    // Select PostgreSQL Template
    // Use getByRole for combobox to find the Select trigger specifically
    const trigger = screen.getByRole('combobox');
    fireEvent.click(trigger);

    const postgresOption = await screen.findByText('PostgreSQL Database');
    fireEvent.click(postgresOption);

    // Verify template description loaded
    expect(screen.getByText('Connect to a PostgreSQL database.')).toBeInTheDocument();

    // Click Next
    fireEvent.click(screen.getByText('Next'));

    // Step 2: Parameters
    expect(screen.getByText('2. Configure Parameters')).toBeInTheDocument();

    // Check for Stdio-specific UI elements to confirm we are in Stdio mode
    expect(screen.getByText('Arguments')).toBeInTheDocument();
    expect(screen.getByText('Add Argument')).toBeInTheDocument();

    // Click Next
    fireEvent.click(screen.getByText('Next'));

    // Step 3: Webhooks
    expect(screen.getByText('3. Webhooks & Transformers')).toBeInTheDocument();
    fireEvent.click(screen.getByText('Next'));

    // Step 4: Auth
    expect(screen.getByText('4. Authentication')).toBeInTheDocument();
    fireEvent.click(screen.getByText('Next'));

    // Step 5: Review
    expect(screen.getByText('5. Review & Finish')).toBeInTheDocument();

    // Click Finish
    fireEvent.click(screen.getByText('Finish & Save to Local Marketplace'));

    // Check the output
    await waitFor(() => {
        expect(onComplete).toHaveBeenCalled();
    });

    const generatedConfig = onComplete.mock.calls[0][0];

    // ASSERTION: Should use mcpService with stdioConnection
    expect(generatedConfig.mcpService).toBeDefined();
    expect(generatedConfig.mcpService.connectionType.stdioConnection).toBeDefined();

    // Should NOT use commandLineService
    expect(generatedConfig.commandLineService).toBeUndefined();

    // Check Command
    expect(generatedConfig.mcpService.connectionType.stdioConnection.command).toContain('npx');

    // Check Args (Should be array)
    expect(Array.isArray(generatedConfig.mcpService.connectionType.stdioConnection.args)).toBe(true);
    expect(generatedConfig.mcpService.connectionType.stdioConnection.args).toContain('@modelcontextprotocol/server-postgres');

    // Check Env
    expect(generatedConfig.mcpService.connectionType.stdioConnection.env['POSTGRES_URL']).toBeDefined();
  });
});
