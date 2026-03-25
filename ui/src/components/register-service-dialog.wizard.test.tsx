/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { RegisterServiceDialog } from './register-service-dialog';
import { SERVICE_TEMPLATES } from '@/lib/templates';
import { vi } from 'vitest';

// Mock dependencies
vi.mock('@/lib/client', () => ({
  apiClient: {
    registerService: vi.fn(),
    updateService: vi.fn(),
    listCredentials: vi.fn().mockResolvedValue([]),
    validateService: vi.fn().mockResolvedValue({ valid: true, message: "Valid" }),
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
    // Add any specific mocks if needed
  };
});

describe('RegisterServiceDialog Wizard', () => {
  it('opens wizard for templates with fields', async () => {
    render(<RegisterServiceDialog />);
    fireEvent.click(screen.getByText('Register Service'));

    // Select PostgreSQL (which has fields)
    const postgresTemplate = SERVICE_TEMPLATES.find(t => t.id === 'postgres');
    expect(postgresTemplate).toBeDefined();
    fireEvent.click(screen.getByText(postgresTemplate!.name));

    // Should show Wizard view
    expect(screen.getByText(`Configure ${postgresTemplate!.name}`)).toBeInTheDocument();

    // Should show the field input
    expect(screen.getByLabelText('PostgreSQL Connection String')).toBeInTheDocument();
  });

  it('substitutes values and populates form', async () => {
    render(<RegisterServiceDialog />);
    fireEvent.click(screen.getByText('Register Service'));

    // Select PostgreSQL
    const postgresTemplate = SERVICE_TEMPLATES.find(t => t.id === 'postgres');
    expect(postgresTemplate).toBeDefined();
    fireEvent.click(screen.getByText(postgresTemplate!.name));

    // Fill the field
    const connectionString = "postgresql://test:test@localhost:5432/testdb";
    fireEvent.change(screen.getByLabelText('PostgreSQL Connection String'), { target: { value: connectionString } });

    // Click Continue
    fireEvent.click(screen.getByText('Continue'));

    // Should switch to main Form view
    expect(screen.getByText('Configure Service')).toBeInTheDocument();

    // Check if command is populated with substitution
    await waitFor(() => {
        expect(screen.getByLabelText('Command')).toBeInTheDocument();
    });

    const commandInput = screen.getByLabelText('Command') as HTMLInputElement;
    const expectedCommand = `npx -y @modelcontextprotocol/server-postgres ${connectionString}`;
    expect(commandInput.value).toBe(expectedCommand);
  });
});
