/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { render, screen, fireEvent } from '@testing-library/react';
import { JsonTree } from './json-tree';
import React from 'react';
import { vi } from 'vitest';

// Mock clipboard
Object.assign(navigator, {
  clipboard: {
    writeText: vi.fn(),
  },
});

describe('JsonTree', () => {
    const sampleData = {
        name: "test",
        count: 42,
        isActive: true,
        nested: {
            foo: "bar"
        },
        list: [1, 2]
    };

    it('renders basic types correctly', () => {
        render(<JsonTree data={sampleData} />);
        expect(screen.getByText('"test"')).toBeInTheDocument();
        expect(screen.getByText('42')).toBeInTheDocument();
        expect(screen.getByText('true')).toBeInTheDocument();
    });

    it('expands nested objects by default if depth matches', () => {
        // Depth 2 means 0 and 1 are expanded.
        render(<JsonTree data={sampleData} defaultExpandDepth={2} />);
        expect(screen.getByText('"bar"')).toBeInTheDocument();
    });

    it('collapses nested objects by default if depth is low', () => {
        // Depth 1 means root (0) is expanded, but nested (1) is collapsed.
        render(<JsonTree data={sampleData} defaultExpandDepth={1} />);

        // Root is expanded, so "nested:" key is visible.
        expect(screen.getByText('nested:')).toBeInTheDocument();

        // "nested" object is collapsed, so "bar" is not visible.
        expect(screen.queryByText('"bar"')).not.toBeInTheDocument();
    });

    it('expands when clicked', () => {
        render(<JsonTree data={sampleData} defaultExpandDepth={1} />);

        const key = screen.getByText('nested:');
        fireEvent.click(key);

        expect(screen.getByText('"bar"')).toBeInTheDocument();
    });
});
