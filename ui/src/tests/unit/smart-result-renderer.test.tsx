/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { render, screen, fireEvent } from '@testing-library/react';
import { SmartResultRenderer } from '@/components/playground/pro/smart-result-renderer';
import { describe, it, expect } from 'vitest';

describe('SmartResultRenderer', () => {
    it('renders Table for simple JSON array and toggles view', () => {
        const data = [
            { id: 1, name: 'Alice' },
            { id: 2, name: 'Bob' }
        ];
        render(<SmartResultRenderer result={data} />);

        // Check for headers
        expect(screen.getByText('id')).toBeDefined();

        // Check for Toggle Buttons
        const jsonButton = screen.getByRole('button', { name: /JSON/i });
        const tableButton = screen.getByRole('button', { name: /Table/i });
        expect(jsonButton).toBeDefined();
        expect(tableButton).toBeDefined();

        // Switch to Raw JSON
        fireEvent.click(jsonButton);

        // Table header should disappear (queryByText returns null if not found)
        expect(screen.queryByText('id')).toBeNull();

        // Switch back to Table
        fireEvent.click(tableButton);
        expect(screen.getByText('id')).toBeDefined();
    });

    it('renders Table for JSON array inside stdout string (Command output)', () => {
        const result = {
            stdout: JSON.stringify([
                { id: 101, status: 'Active' },
                { id: 102, status: 'Inactive' }
            ])
        };
        render(<SmartResultRenderer result={result} />);
        expect(screen.getByText('status')).toBeDefined();
        expect(screen.getByText('Active')).toBeDefined();
    });

    it('renders Table for CallToolResult structure with nested JSON', () => {
        // This simulates what mcpany returns for a command_line tool echo call
        const result = {
            content: [
                {
                    type: 'text',
                    text: JSON.stringify([
                         { sku: 'ABC', qty: 10 }
                    ])
                }
            ],
            isError: false
        };
        render(<SmartResultRenderer result={result} />);
        expect(screen.getByText('sku')).toBeDefined();
        expect(screen.getByText('ABC')).toBeDefined();
    });

    it('renders Properties view for non-array object data', () => {
        const data = { id: 1, name: 'Alice' };
        render(<SmartResultRenderer result={data} />);

        // Should have Properties button
        const propButton = screen.getByRole('button', { name: /Properties/i });
        expect(propButton).toBeDefined();

        // Should render key-value structure instead of array table
        // We can check if `table` element exists (since it uses Table under the hood for properties)
        // But it won't have standard column headers like a smart table.
        // So we can check if "id" and "1" appear as cells.
        expect(screen.getByText('id')).toBeDefined();
        expect(screen.getByText('1')).toBeDefined();
        expect(screen.getByText('name')).toBeDefined();
        expect(screen.getByText('Alice')).toBeDefined();
    });
});
