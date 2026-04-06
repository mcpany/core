/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { render, screen } from '@testing-library/react';
import { describe, it, expect } from 'vitest';
import { SmartResultRenderer } from './smart-result-renderer';

describe('SmartResultRenderer', () => {
    it('renders text-only JSON as a table', () => {
        const result = [
            { id: 1, name: 'Alice' },
            { id: 2, name: 'Bob' }
        ];
        render(<SmartResultRenderer result={result} />);
        expect(screen.getByText('Alice')).toBeInTheDocument();
        expect(screen.getByText('Bob')).toBeInTheDocument();
        // It renders a table
        expect(screen.getByRole('table')).toBeInTheDocument();
    });

    it('renders CallToolResult with image content', () => {
        const result = {
            content: [
                {
                    type: 'image',
                    data: 'base64data',
                    mimeType: 'image/png'
                }
            ]
        };
        render(<SmartResultRenderer result={result} />);

        const img = screen.queryByRole('img');
        expect(img).toBeInTheDocument();
        expect(img).toHaveAttribute('src', 'data:image/png;base64,base64data');
    });

    it('renders CallToolResult with mixed content (text + image)', () => {
        const result = {
            content: [
                { type: 'text', text: 'Here is a chart:' },
                {
                    type: 'image',
                    data: 'base64chart',
                    mimeType: 'image/jpeg'
                }
            ]
        };
        render(<SmartResultRenderer result={result} />);

        expect(screen.getByText('Here is a chart:')).toBeInTheDocument();

        const img = screen.queryByRole('img');
        expect(img).toBeInTheDocument();
        expect(img).toHaveAttribute('src', 'data:image/jpeg;base64,base64chart');
    });

    it('renders nested JSON content from command output correctly', () => {
        // Simulates `echo '[{"type": "image", ...}]'`
        const result = {
            stdout: JSON.stringify([
                {
                    type: 'image',
                    data: 'base64nested',
                    mimeType: 'image/gif'
                }
            ])
        };
        render(<SmartResultRenderer result={result} />);

        const img = screen.queryByRole('img');
        expect(img).toBeInTheDocument();
        expect(img).toHaveAttribute('src', 'data:image/gif;base64,base64nested');
    });

    it('renders generic JSON array from CLI stdout as a table (regression test)', () => {
        const result = {
            stdout: JSON.stringify([
                { id: 1, name: 'Item 1' },
                { id: 2, name: 'Item 2' }
            ])
        };
        render(<SmartResultRenderer result={result} />);
        expect(screen.getByRole('table')).toBeInTheDocument();
        expect(screen.getByText('Item 1')).toBeInTheDocument();
    });
});
