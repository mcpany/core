/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { SchemaVisualizer } from './schema-visualizer';

// Mock ResizeObserver
global.ResizeObserver = class ResizeObserver {
    observe() {}
    unobserve() {}
    disconnect() {}
};

describe('SchemaVisualizer', () => {
    it('renders "No input schema defined" when schema is empty or null', () => {
        const { rerender } = render(<SchemaVisualizer schema={null} />);
        expect(screen.getByText('No input schema defined.')).toBeInTheDocument();

        rerender(<SchemaVisualizer schema={{}} />);
        expect(screen.getByText('No input schema defined.')).toBeInTheDocument();
    });

    it('renders simple object properties', () => {
        const schema = {
            type: 'object',
            properties: {
                firstName: {
                    type: 'string',
                    description: 'The user first name',
                },
                age: {
                    type: 'integer',
                    description: 'The user age',
                },
            },
            required: ['firstName'],
        };

        render(<SchemaVisualizer schema={schema} />);

        // Check property names
        expect(screen.getByText('firstName')).toBeInTheDocument();
        expect(screen.getByText('age')).toBeInTheDocument();

        // Check types
        // Note: badges might be rendered in a specific way, looking for text content
        expect(screen.getAllByText('string').length).toBeGreaterThan(0);
        expect(screen.getAllByText('integer').length).toBeGreaterThan(0);

        // Check descriptions
        expect(screen.getByText('The user first name')).toBeInTheDocument();
        expect(screen.getByText('The user age')).toBeInTheDocument();

        // Check required asterisk (might be hard to target specifically, but we can check if it exists in DOM)
        // In our implementation, required adds a class or element.
        // We can check if "firstName" has the required style or asterisk nearby.
        // The asterisk is in a separate span with text "*"
        expect(screen.getByText('*')).toBeInTheDocument();
    });

    it('renders nested objects', () => {
        const schema = {
            type: 'object',
            properties: {
                address: {
                    type: 'object',
                    properties: {
                        street: { type: 'string' },
                        city: { type: 'string' },
                    },
                },
            },
        };

        render(<SchemaVisualizer schema={schema} />);

        expect(screen.getByText('address')).toBeInTheDocument();

        // Nested properties should be visible because 'expanded' is true by default
        expect(screen.getByText('street')).toBeInTheDocument();
        expect(screen.getByText('city')).toBeInTheDocument();
    });

    it('renders array items', () => {
        const schema = {
            type: 'object',
            properties: {
                tags: {
                    type: 'array',
                    items: {
                        type: 'string',
                        description: 'A tag',
                    },
                },
            },
        };

        render(<SchemaVisualizer schema={schema} />);

        expect(screen.getByText('tags')).toBeInTheDocument();
        // Array items usually rendered with name "items" in our component for the expanded view
        expect(screen.getByText('items')).toBeInTheDocument();
        expect(screen.getByText('A tag')).toBeInTheDocument();
    });

    it('renders deeply nested objects collapsed by default', () => {
        const schema = {
            type: 'object',
            properties: {
                level0: {
                    type: 'object',
                    properties: {
                        level1: {
                            type: 'object',
                            properties: {
                                level2: { type: 'string' },
                            },
                        },
                    },
                },
            },
        };

        render(<SchemaVisualizer schema={schema} />);

        // Level 0 should be visible
        expect(screen.getByText('level0')).toBeInTheDocument();

        // Level 1 should be visible (as child of expanded level 0)
        expect(screen.getByText('level1')).toBeInTheDocument();

        // Level 2 should NOT be visible initially (child of collapsed level 1)
        expect(screen.queryByText('level2')).not.toBeInTheDocument();

        // Find the expand button for level 1
        // Since level 1 is rendered, it has a row. The expand button is in the first cell.
        // We can find the row by text 'level1' and then find the button within it.
        // Or simpler: getAllByRole('button') and click the one corresponding to level 1.
        // Since level 0 is expanded, it has a ChevronDown. Level 1 is collapsed, it has a ChevronRight.
        // But testing-library by role is better.
        // Let's rely on the fact that only level 1 has a ChevronRight initially (level 0 has ChevronDown).
        // But how do we distinguish? We can check for the button associated with 'level1'.

        // Alternative: Click the button inside the row containing 'level1'.
        const level1Row = screen.getByText('level1').closest('tr');
        const expandButton = level1Row?.querySelector('button');
        expect(expandButton).toBeInTheDocument();

        if (expandButton) {
            fireEvent.click(expandButton);
        }

        // Now Level 2 should be visible
        expect(screen.getByText('level2')).toBeInTheDocument();
    });
});
