/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import React from 'react';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { ShareCollectionDialog } from '../../../components/share-collection-dialog';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { apiClient } from '../../../lib/client';

// Mock apiClient
vi.mock('../../../lib/client', () => ({
    apiClient: {
        listServices: vi.fn(),
    },
}));

// Mock clipboard
Object.assign(navigator, {
  clipboard: {
    writeText: vi.fn(),
  },
});

describe('ShareCollectionDialog', () => {
    beforeEach(() => {
        vi.clearAllMocks();
    });

    it('should fetch services and allow generation', async () => {
        const user = userEvent.setup();

        (apiClient.listServices as any).mockResolvedValue([
            { name: "service-1", httpService: { address: "http://localhost:8080" } },
            { name: "service-2", mcpService: { address: "http://localhost:8081" } }
        ]);

        render(<ShareCollectionDialog />);

        // Open Dialog
        const trigger = screen.getByText('Share Your Config');
        await user.click(trigger);

        // Wait for loading to finish and services to appear
        await waitFor(() => {
             expect(screen.getByText('service-1')).toBeInTheDocument();
             expect(screen.getByText('service-2')).toBeInTheDocument();
        });

        // Select service-1
        // Shadcn checkbox usually has role="checkbox".
        // Index 0 is "Select All" header, 1 is service-1, 2 is service-2
        const checkboxes = screen.getAllByRole('checkbox');
        await user.click(checkboxes[1]);

        // Click Generate
        const generateBtn = screen.getByText('Generate Configuration');
        expect(generateBtn).toBeEnabled();
        await user.click(generateBtn);

        // Expect textarea with content
        await waitFor(() => {
            const textarea = screen.getByRole('textbox') as HTMLTextAreaElement;
            expect(textarea).toBeInTheDocument();
            expect(textarea.value).toContain('service-1');
            // service-2 was not selected
            expect(textarea.value).not.toContain('service-2');
        });
    });
});
