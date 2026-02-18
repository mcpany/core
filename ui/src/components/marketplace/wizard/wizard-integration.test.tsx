/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { CreateConfigWizard } from './create-config-wizard';
import React from 'react';
import { vi, describe, it, expect } from 'vitest';

// Mock UI components that might cause issues in jsdom
vi.mock('@/components/ui/dialog', () => ({
    Dialog: ({ children, open }: any) => open ? <div>{children}</div> : null,
    DialogContent: ({ children }: any) => <div>{children}</div>,
    DialogHeader: ({ children }: any) => <div>{children}</div>,
    DialogTitle: ({ children }: any) => <div>{children}</div>,
    DialogDescription: ({ children }: any) => <div>{children}</div>,
    DialogFooter: ({ children }: any) => <div>{children}</div>,
}));

// Mock ResizeObserver
global.ResizeObserver = class ResizeObserver {
    observe() {}
    unobserve() {}
    disconnect() {}
};

// Mock the registry to ensure we have known data
vi.mock('@/lib/service-registry', () => ({
    SERVICE_REGISTRY: [{
        id: 'postgres',
        name: 'PostgreSQL',
        description: 'Postgres DB',
        command: 'npx -y @modelcontextprotocol/server-postgres',
        configurationSchema: {
            properties: {
                POSTGRES_URL: {
                    title: 'Connection URL',
                    type: 'string',
                    default: 'postgres://localhost'
                }
            }
        }
    }]
}));

// Mock API Client
const mockSaveTemplate = vi.fn().mockResolvedValue({});
const mockCreateProfile = vi.fn().mockResolvedValue({});
const mockGetServiceTemplates = vi.fn()
    .mockResolvedValueOnce([]) // First call: empty -> triggers seeding
    .mockResolvedValueOnce([   // Second call: returns seeded template
        {
            id: 'postgres',
            name: 'PostgreSQL',
            description: 'Postgres DB',
            serviceConfig: {
                commandLineService: {
                    command: 'npx -y @modelcontextprotocol/server-postgres'
                },
                configurationSchema: JSON.stringify({
                    properties: {
                        POSTGRES_URL: {
                            title: 'Connection URL',
                            type: 'string',
                            default: 'postgres://localhost'
                        }
                    }
                })
            }
        }
    ]);

vi.mock('@/lib/client', () => ({
    apiClient: {
        getServiceTemplates: () => mockGetServiceTemplates(),
        saveTemplate: (config: any) => mockSaveTemplate(config),
        createProfile: (profile: any) => mockCreateProfile(profile),
        listProfiles: vi.fn().mockResolvedValue([]),
        listCredentials: vi.fn().mockResolvedValue([]),
        getSystemStatus: vi.fn().mockResolvedValue({}),
        registerService: vi.fn().mockResolvedValue({}),
    }
}));

describe('CreateConfigWizard Integration', () => {
    it('should allow selecting a registry template and filling schema form', async () => {
        const onComplete = vi.fn();
        const onOpenChange = vi.fn();

        render(
            <CreateConfigWizard open={true} onOpenChange={onOpenChange} onComplete={onComplete} />
        );

        // Step 1: Select Service Type
        expect(screen.getByText('1. Select Service Type')).toBeInTheDocument();

        // The default selected is 'manual', so "Manual / Custom".
        const trigger = screen.getByRole('combobox');

        // Wait for loading to finish (select becomes enabled)
        await waitFor(() => {
            expect(trigger).not.toBeDisabled();
        });

        // Radix UI Select often requires click or pointerDown
        fireEvent.click(trigger);

        // Now options should be visible.
        const postgresOption = await screen.findByText('PostgreSQL');
        fireEvent.click(postgresOption);

        // Click Next
        const nextButton = screen.getByText('Next');
        fireEvent.click(nextButton);

        // Step 2: Configure Parameters
        // Should show "Configuration" title because Postgres has a schema.
        await waitFor(() => {
            expect(screen.getByText('Configuration')).toBeInTheDocument();
        });

        // Should see "Connection URL" input (from schema title)
        const urlInput = screen.getByLabelText(/Connection URL/i);
        expect(urlInput).toBeInTheDocument();

        // Fill it
        fireEvent.change(urlInput, { target: { value: 'postgres://test:test@localhost:5432/test' } });

        // Click Next
        fireEvent.click(screen.getByText('Next'));

        // Step 3: Webhooks (Skip)
        expect(screen.getByText('3. Webhooks & Transformers')).toBeInTheDocument();
        fireEvent.click(screen.getByText('Next'));

        // Step 4: Auth (Skip)
        expect(screen.getByText('4. Authentication')).toBeInTheDocument();
        fireEvent.click(screen.getByText('Next'));

        // Step 5: Review
        expect(screen.getByText('5. Review & Finish')).toBeInTheDocument();

        const finishButton = screen.getByText('Finish & Save to Local Marketplace');
        fireEvent.click(finishButton);

        await waitFor(() => {
            expect(onComplete).toHaveBeenCalled();
        });

        const config = onComplete.mock.calls[0][0];

        expect(config.commandLineService.command).toContain('server-postgres');
        expect(config.commandLineService.env['POSTGRES_URL'].plainText).toBe('postgres://test:test@localhost:5432/test');
    });
});
