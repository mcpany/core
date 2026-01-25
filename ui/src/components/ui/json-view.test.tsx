/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import React from 'react';
import { render, screen } from '@testing-library/react';
import { JsonView } from './json-view';
import userEvent from '@testing-library/user-event';
import { vi } from 'vitest';

describe('JsonView', () => {
    it('renders json data', () => {
        const data = { foo: 'bar', num: 123 };
        render(<JsonView data={data} />);

        // It renders inside a pre, so we check for the text
        expect(screen.getByText(/"foo": "bar"/)).toBeInTheDocument();
        expect(screen.getByText(/"num": 123/)).toBeInTheDocument();
    });

    it('renders null when data is null/undefined', () => {
        const { rerender } = render(<JsonView data={null} />);
        expect(screen.getByText('null')).toBeInTheDocument();

        rerender(<JsonView data={undefined} />);
        expect(screen.getByText('null')).toBeInTheDocument();
    });

    it('copies to clipboard on click', async () => {
        const user = userEvent.setup();
        const data = { foo: 'bar' };

        // Mock clipboard
        const writeTextMock = vi.fn().mockResolvedValue(undefined);
        Object.defineProperty(navigator, 'clipboard', {
            value: {
                writeText: writeTextMock,
            },
            writable: true,
        });

        render(<JsonView data={data} />);

        // Button with title "Copy JSON"
        const button = screen.getByTitle('Copy JSON');
        await user.click(button);

        expect(writeTextMock).toHaveBeenCalledWith(JSON.stringify(data, null, 2));
    });
});
