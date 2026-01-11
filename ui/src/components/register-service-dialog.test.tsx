/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { RegisterServiceDialog } from './register-service-dialog';
import { SERVICE_TEMPLATES } from '@/lib/templates';
import { vi } from 'vitest';

// Mock the API client
vi.mock('@/lib/client', () => ({
  apiClient: {
    registerService: vi.fn(),
    updateService: vi.fn(),
    listCredentials: vi.fn().mockResolvedValue([]),
  },
}));

// Mock useToast
vi.mock('@/hooks/use-toast', () => ({
  useToast: () => ({
    toast: vi.fn(),
  }),
}));

// Mock Lucide icons component to avoid import issues in tests?
// No, they should be fine if installed.
// But Button uses Radix which might have issues in JSDOM environment if not polyfilled properly for PointerEvents etc.
// Let's assume standard JSDOM setup in vitest config.

describe('RegisterServiceDialog', () => {
  it('renders the trigger button', () => {
    render(<RegisterServiceDialog />);
    expect(screen.getByText('Register Service')).toBeInTheDocument();
  });

  it('shows template selection when opened for new service', () => {
    render(<RegisterServiceDialog />);
    fireEvent.click(screen.getByText('Register Service'));
    expect(screen.getByText('Select Service Template')).toBeInTheDocument();

    // Check if templates are rendered
    SERVICE_TEMPLATES.forEach(template => {
       expect(screen.getByText(template.name)).toBeInTheDocument();
    });
  });

  it('populates form when template is selected', async () => {
    render(<RegisterServiceDialog />);
    fireEvent.click(screen.getByText('Register Service'));

    // Click on PostgreSQL template
    const postgresTemplate = SERVICE_TEMPLATES.find(t => t.id === 'postgres');
    expect(postgresTemplate).toBeDefined();
    if (!postgresTemplate) return;

    fireEvent.click(screen.getByText(postgresTemplate.name));

    // Should switch to Configure Service view
    expect(screen.getByText('Configure Service')).toBeInTheDocument();

    // Check if name field is populated
    const nameInput = screen.getByLabelText('Service Name') as HTMLInputElement;
    expect(nameInput.value).toBe(postgresTemplate.config.name);

    // Check if command is populated (switch to Command Line if needed, but the template pre-sets it?)
    // The form default type depends on the template config.
    // PostgreSQL uses command_line_service.
    // Wait, the select component might not be easily queryable by label text for value.
    // But the command input should be visible if type is command_line.

    await waitFor(() => {
        expect(screen.getByLabelText('Command')).toBeInTheDocument();
    });

    const commandInput = screen.getByLabelText('Command') as HTMLInputElement;
    expect(commandInput.value).toBe(postgresTemplate.config.commandLineService?.command);
  });

  it('allows going back to templates', () => {
      render(<RegisterServiceDialog />);
      fireEvent.click(screen.getByText('Register Service'));

      // Select a template
      fireEvent.click(screen.getByText('Custom Service'));
      expect(screen.getByText('Configure Service')).toBeInTheDocument();

      // Click Back button
      fireEvent.click(screen.getByLabelText('Back to templates'));

      expect(screen.getByText('Select Service Template')).toBeInTheDocument();
  });
});
