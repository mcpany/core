/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { ToolSafetyTable } from './tool-safety-table';
import { apiClient } from '@/lib/client';
import { ToolDefinition } from '@/lib/types';
import { vi, describe, it, expect } from 'vitest';

// Mock apiClient
vi.mock('@/lib/client', () => ({
    apiClient: {
        setToolStatus: vi.fn(),
    },
}));

// Mock useToast
vi.mock('@/hooks/use-toast', () => ({
    useToast: () => ({
        toast: vi.fn(),
    }),
}));

const mockTools: ToolDefinition[] = [
    {
        name: 'test-tool',
        description: 'A test tool',
        inputSchema: {},
    },
    {
        name: 'disabled-tool',
        description: 'A disabled tool',
        inputSchema: {},
        disable: true,
    } as any
];

describe('ToolSafetyTable', () => {
    it('renders a list of tools', () => {
        render(<ToolSafetyTable tools={mockTools} />);
        expect(screen.getByText('test-tool')).toBeInTheDocument();
        expect(screen.getByText('disabled-tool')).toBeInTheDocument();
    });

    it('displays correct status for enabled/disabled tools', () => {
        render(<ToolSafetyTable tools={mockTools} />);
        // test-tool is enabled (default)
        // disabled-tool is disabled
        // We can check for "Enabled" and "Disabled" text
        expect(screen.getAllByText('Enabled')).toHaveLength(1);
        expect(screen.getAllByText('Disabled')).toHaveLength(1);
    });

    it('calls setToolStatus when toggle is clicked', async () => {
        const onUpdate = vi.fn();
        render(<ToolSafetyTable tools={mockTools} onUpdate={onUpdate} />);

        // Find toggle for test-tool (which is enabled)
        // The Switch component usually has a role="switch"
        const switches = screen.getAllByRole('switch');
        const enabledSwitch = switches[0]; // Assuming order preserved

        fireEvent.click(enabledSwitch);

        await waitFor(() => {
            expect(apiClient.setToolStatus).toHaveBeenCalledWith('test-tool', true); // Expect disable=true
            expect(onUpdate).toHaveBeenCalled();
        });
    });
});
