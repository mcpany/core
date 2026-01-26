/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { ToolTable } from './tool-table';
import { vi } from 'vitest';
import { ToolDefinition } from '@proto/config/v1/tool';
import { ToolAnalytics } from '@/lib/client';

// Mock dependencies
vi.mock('@/lib/tokens', () => ({
    estimateTokens: vi.fn().mockReturnValue(100),
    formatTokenCount: vi.fn().mockReturnValue('100'),
}));

describe('ToolTable', () => {
    const mockTools: ToolDefinition[] = [
        {
            name: 'tool-1',
            description: 'Description 1',
            serviceId: 'service-1',
            disable: false,
        },
        {
            name: 'tool-2',
            description: 'Description 2',
            serviceId: 'service-2',
            disable: true,
        },
    ];

    const mockUsageStats: Record<string, ToolAnalytics> = {
        'tool-1@service-1': {
            name: 'tool-1',
            serviceId: 'service-1',
            totalCalls: 10,
            successRate: 95.5,
        }
    };

    const mockIsPinned = vi.fn().mockReturnValue(false);
    const mockTogglePin = vi.fn();
    const mockToggleTool = vi.fn();
    const mockOpenInspector = vi.fn();

    beforeEach(() => {
        vi.clearAllMocks();
    });

    it('renders tools correctly', () => {
        render(
            <ToolTable
                tools={mockTools}
                isCompact={false}
                isPinned={mockIsPinned}
                togglePin={mockTogglePin}
                toggleTool={mockToggleTool}
                openInspector={mockOpenInspector}
                usageStats={mockUsageStats}
            />
        );

        expect(screen.getByText('tool-1')).toBeInTheDocument();
        expect(screen.getByText('Description 1')).toBeInTheDocument();
        expect(screen.getByText('service-1')).toBeInTheDocument();

        expect(screen.getByText('tool-2')).toBeInTheDocument();
        expect(screen.getByText('Description 2')).toBeInTheDocument();
        expect(screen.getByText('service-2')).toBeInTheDocument();

        // Check stats
        expect(screen.getByText('10')).toBeInTheDocument(); // Total calls
        expect(screen.getByText('95.5%')).toBeInTheDocument(); // Success rate
    });

    it('handles interactions', () => {
        render(
            <ToolTable
                tools={mockTools}
                isCompact={false}
                isPinned={mockIsPinned}
                togglePin={mockTogglePin}
                toggleTool={mockToggleTool}
                openInspector={mockOpenInspector}
                usageStats={mockUsageStats}
            />
        );

        // Click pin
        const pinButtons = screen.getAllByLabelText(/Pin tool-/);
        fireEvent.click(pinButtons[0]);
        expect(mockTogglePin).toHaveBeenCalledWith('tool-1');

        // Toggle tool (Switch)
        // Switch is usually a button with role="switch"
        const switches = screen.getAllByRole('switch');
        fireEvent.click(switches[0]);
        // tool-1 is enabled (disable: false), so clicking it should toggle to disable: true
        expect(mockToggleTool).toHaveBeenCalledWith('tool-1', true);

        // Open Inspector
        const inspectButtons = screen.getAllByText('Inspect');
        fireEvent.click(inspectButtons[0]);
        expect(mockOpenInspector).toHaveBeenCalledWith(mockTools[0]);
    });

    it('renders compact mode', () => {
        const { container } = render(
            <ToolTable
                tools={mockTools}
                isCompact={true}
                isPinned={mockIsPinned}
                togglePin={mockTogglePin}
                toggleTool={mockToggleTool}
                openInspector={mockOpenInspector}
                usageStats={mockUsageStats}
            />
        );

        // Check for compact classes, e.g., h-8
        const rows = container.querySelectorAll('tr');
        // First row is header, others are body. Check body row.
        expect(rows[1]).toHaveClass('h-8');
    });
});
