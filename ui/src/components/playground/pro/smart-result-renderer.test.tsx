/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { render, screen } from '@testing-library/react';
import { describe, it, expect, vi } from 'vitest';
import { SmartResultRenderer } from './smart-result-renderer';
import '@testing-library/jest-dom';

describe('SmartResultRenderer', () => {
    // 1. Existing Behavior: Simple array of objects -> Table
    it('renders a table for an array of simple objects', () => {
        const data = [
            { id: 1, name: 'Alice' },
            { id: 2, name: 'Bob' },
        ];
        render(<SmartResultRenderer result={data} />);

        // Should find table headers
        expect(screen.getByText('id')).toBeInTheDocument();
        expect(screen.getByText('name')).toBeInTheDocument();
        // Should find data
        expect(screen.getByText('Alice')).toBeInTheDocument();
        expect(screen.getByText('Bob')).toBeInTheDocument();
    });

    // 2. Existing Behavior: CallToolResult with text -> Table (via unwrapping)
    it('unwraps CallToolResult text content and renders table if applicable', () => {
        const result = {
            content: [
                {
                    type: 'text',
                    text: JSON.stringify([
                        { id: 1, status: 'ok' }
                    ])
                }
            ]
        };
        render(<SmartResultRenderer result={result} />);

        expect(screen.getByText('status')).toBeInTheDocument();
        expect(screen.getByText('ok')).toBeInTheDocument();
    });

    // 3. New Requirement: CallToolResult with Image -> Image tag
    it('renders an image for CallToolResult with image content', () => {
        const base64Data = 'iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mNk+A8AAQUBAScY42YAAAAASUVORK5CYII=';
        const result = {
            content: [
                {
                    type: 'image',
                    data: base64Data,
                    mimeType: 'image/png'
                }
            ]
        };
        render(<SmartResultRenderer result={result} />);

        const img = screen.getByRole('img');
        expect(img).toBeInTheDocument();
        expect(img).toHaveAttribute('src', `data:image/png;base64,${base64Data}`);
    });

    // 4. New Requirement: CallToolResult with Mixed Content -> Multiple items
    it('renders mixed text and image content', () => {
        const base64Data = 'iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mNk+A8AAQUBAScY42YAAAAASUVORK5CYII=';
        const result = {
            content: [
                {
                    type: 'text',
                    text: 'Here is an image:'
                },
                {
                    type: 'image',
                    data: base64Data,
                    mimeType: 'image/png'
                }
            ]
        };
        render(<SmartResultRenderer result={result} />);

        expect(screen.getByText('Here is an image:')).toBeInTheDocument();
        const img = screen.getByRole('img');
        expect(img).toBeInTheDocument();
        expect(img).toHaveAttribute('src', `data:image/png;base64,${base64Data}`);
    });

    // 5. New Requirement: Command Output (stdout) containing nested MCP Content -> Unwrap and Render Image
    it('unwraps command stdout containing MCP content and renders image', () => {
        const base64Data = 'iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mNk+A8AAQUBAScY42YAAAAASUVORK5CYII=';
        const nestedContent = {
            content: [
                {
                    type: 'image',
                    data: base64Data,
                    mimeType: 'image/png'
                }
            ]
        };

        const result = {
            command: 'echo ...',
            stdout: JSON.stringify(nestedContent)
        };

        render(<SmartResultRenderer result={result} />);

        const img = screen.getByRole('img');
        expect(img).toBeInTheDocument();
        expect(img).toHaveAttribute('src', `data:image/png;base64,${base64Data}`);
    });
});
