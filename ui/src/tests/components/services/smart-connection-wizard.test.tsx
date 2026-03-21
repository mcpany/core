/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { SmartConnectionWizard } from '@/components/services/smart-connection-wizard';
import { apiClient } from '@/lib/client';
import { ServiceTemplate } from '@/lib/templates';
import { Server } from 'lucide-react';

// Mock the API client
vi.mock('@/lib/client', () => ({
    apiClient: {
        validateService: vi.fn()
    }
}));

describe('SmartConnectionWizard', () => {
    const mockTemplate: ServiceTemplate = {
        id: 'test-template',
        name: 'Test Template',
        description: 'A template for testing',
        icon: Server,
        config: {
            name: 'default-test-name'
        },
        fields: [
            {
                name: 'apiKey',
                label: 'API Key',
                placeholder: 'Enter key',
                key: 'httpService.headers.Authorization'
            }
        ]
    };

    const mockOnCancel = vi.fn();
    const mockOnComplete = vi.fn();

    beforeEach(() => {
        vi.clearAllMocks();
    });

    it('renders step 1 with template fields', () => {
        render(
            <SmartConnectionWizard
                template={mockTemplate}
                onCancel={mockOnCancel}
                onComplete={mockOnComplete}
            />
        );

        expect(screen.getByText('Test Template')).toBeInTheDocument();
        expect(screen.getByLabelText('Service Name')).toHaveValue('default-test-name');
        expect(screen.getByLabelText('API Key')).toBeInTheDocument();
    });

    it('handles successful validation and moves to step 3', async () => {
        // Setup successful mock response
        const mockValidateService = apiClient.validateService as ReturnType<typeof vi.fn>;
        mockValidateService.mockResolvedValueOnce({
            valid: true,
            discoveredTools: [{ name: 'test_tool', description: 'A test tool' }]
        });

        render(
            <SmartConnectionWizard
                template={mockTemplate}
                onCancel={mockOnCancel}
                onComplete={mockOnComplete}
            />
        );

        // Fill required fields
        fireEvent.change(screen.getByLabelText('API Key'), { target: { value: 'secret123' } });

        // Click connect
        fireEvent.click(screen.getByRole('button', { name: /Connect/i }));

        // Verify API was called
        await waitFor(() => {
            expect(apiClient.validateService).toHaveBeenCalledTimes(1);
        });

        // Verify Step 3 is shown
        await waitFor(() => {
            expect(screen.getByText('Connection Successful')).toBeInTheDocument();
            // Tools are rendered by a subcomponent, but we should see the title
            expect(screen.getByText(/Discovered Tools/i)).toBeInTheDocument();
        });

        // Click Finish
        fireEvent.click(screen.getByRole('button', { name: /Save & Finish/i }));

        expect(mockOnComplete).toHaveBeenCalledTimes(1);
        const calledConfig = mockOnComplete.mock.calls[0][0];
        expect(calledConfig.name).toBe('default-test-name');
        expect((calledConfig.httpService as { headers: Record<string, string> }).headers.Authorization).toBe('secret123');
    });

    it('displays error message when validation fails', async () => {
        // Setup failed mock response
        const mockValidateService = apiClient.validateService as ReturnType<typeof vi.fn>;
        mockValidateService.mockResolvedValueOnce({
            valid: false,
            error: 'Invalid API Key'
        });

        render(
            <SmartConnectionWizard
                template={mockTemplate}
                onCancel={mockOnCancel}
                onComplete={mockOnComplete}
            />
        );

        // Fill required fields
        fireEvent.change(screen.getByLabelText('API Key'), { target: { value: 'bad_key' } });

        // Click connect
        fireEvent.click(screen.getByRole('button', { name: /Connect/i }));

        // Verify Step 2 Error is shown
        await waitFor(() => {
            expect(screen.getByText('Connection Failed')).toBeInTheDocument();
            expect(screen.getByText('Invalid API Key')).toBeInTheDocument();
        });

        // Verify back button works
        fireEvent.click(screen.getByRole('button', { name: /Back to Edit/i }));
        expect(screen.getByLabelText('Service Name')).toBeInTheDocument(); // Step 1
    });
});