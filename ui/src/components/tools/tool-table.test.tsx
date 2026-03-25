/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import React from 'react';
import { render, screen } from '@testing-library/react';
import { ToolTable } from './tool-table';
import { vi } from 'vitest';

describe('ToolTable Empty State', () => {
    it('renders empty state when tools array is empty', () => {
        const mockIsPinned = vi.fn().mockReturnValue(false);
        const mockTogglePin = vi.fn();
        const mockToggleTool = vi.fn();
        const mockOpenInspector = vi.fn();

        render(
            <ToolTable
                tools={[]}
                isCompact={false}
                isPinned={mockIsPinned}
                togglePin={mockTogglePin}
                toggleTool={mockToggleTool}
                openInspector={mockOpenInspector}
            />
        );

        // Verify the empty state text is rendered
        expect(screen.getByText('No tools found')).toBeInTheDocument();
        expect(screen.getByText('No tools found for this service or matching your search.')).toBeInTheDocument();

        // Verify that only the table header and the single empty state row are rendered
        // 1 row in the thead + 1 empty state row in tbody = 2 rows total
        const rows = screen.getAllByRole('row');
        expect(rows.length).toBe(2);
    });
});
