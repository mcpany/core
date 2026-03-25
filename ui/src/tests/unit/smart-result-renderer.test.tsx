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

        // Check for Toggle Buttons (Top-level SmartResultRenderer toggles)
        const jsonButton = screen.getByRole('button', { name: 'JSON' });
        const tableButton = screen.getByRole('button', { name: 'Table' });
        expect(jsonButton).toBeDefined();
        expect(tableButton).toBeDefined();

        // Switch to Raw JSON
        fireEvent.click(jsonButton);

        // In JSON view, we use `<JsonView>` which defaults to smartTable=true.
        // It renders a table if it detects an array.
        // To verify we actually switched to the JsonView mode, we can look for
        // the JsonView's specific toggle buttons which only appear when it renders.
        const innerRawJsonButton = screen.getByRole('button', { name: /Raw/i });
        expect(innerRawJsonButton).toBeDefined();

        // Actually click "Raw" to completely remove the table view
        fireEvent.click(innerRawJsonButton);

        // Now there should be absolutely no table elements in the DOM.
        expect(screen.queryByRole('table')).toBeNull();

        // Switch back to Top-level Table
        fireEvent.click(tableButton);
        expect(screen.getByText('id')).toBeDefined();
        expect(screen.getByRole('table')).toBeDefined();
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

    it('renders Raw JSON for non-array data', () => {
        const data = { id: 1, name: 'Alice' };
        render(<SmartResultRenderer result={data} />);
        // Table headers should not exist
        const table = screen.queryByRole('table');
        expect(table).toBeNull();
    });
});
