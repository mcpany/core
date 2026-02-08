/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { render, screen } from '@testing-library/react';
import { describe, it, expect } from 'vitest';
import { SmartResultRenderer } from './smart-result-renderer';

describe('SmartResultRenderer', () => {
    it('renders raw JSON when result is not recognized as MCP content', () => {
        const result = { foo: 'bar' };
        render(<SmartResultRenderer result={result} />);
        // It should render JSON view. We can check for text content.
        // JsonView renders tokens separately
        expect(screen.getByText(/"foo"/)).toBeInTheDocument();
        expect(screen.getByText(/"bar"/)).toBeInTheDocument();
    });

    it('renders a table for array of objects (text data)', () => {
        const result = {
            content: [
                {
                    type: 'text',
                    text: JSON.stringify([{ id: 1, name: 'Alice' }, { id: 2, name: 'Bob' }])
                }
            ]
        };
        render(<SmartResultRenderer result={result} />);
        // Should render table headers
        expect(screen.getByText('id')).toBeInTheDocument();
        expect(screen.getByText('name')).toBeInTheDocument();
        // Should render table content
        expect(screen.getByText('Alice')).toBeInTheDocument();
        expect(screen.getByText('Bob')).toBeInTheDocument();
    });

    it('renders an image when content has image type', () => {
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
        expect(img).toHaveAttribute('src', `data:image/png;base64,${base64Data}`);
    });

    it('renders image from nested JSON in text content (Command Output)', () => {
        const base64Data = 'iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mNk+A8AAQUBAScY42YAAAAASUVORK5CYII=';
        // Simulating command output where stdout is a JSON string of MCP content
        const result = {
            command: 'test',
            stdout: JSON.stringify([
                {
                    type: 'image',
                    data: base64Data,
                    mimeType: 'image/png'
                }
            ])
        };
        render(<SmartResultRenderer result={result} />);

        const img = screen.getByRole('img');
        expect(img).toBeInTheDocument();
        expect(img).toHaveAttribute('src', `data:image/png;base64,${base64Data}`);
    });
});
