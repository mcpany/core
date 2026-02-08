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

        // Level 0 is expanded by default (level < 1)
        expect(screen.getByText('level0')).toBeInTheDocument();
        // Level 1 should be visible (child of Level 0)
        expect(screen.getByText('level1')).toBeInTheDocument();

        // Level 2 should NOT be visible (child of Level 1, which is collapsed by default)
        expect(screen.queryByText('level2')).not.toBeInTheDocument();

        // Find the expand button for level1.
        // Since level0 is also expandable, there are multiple buttons.
        // We need to find the one associated with level1.
        // The structure is TableRow -> TableCell -> div -> button.
        // We can find the row containing "level1" and then find the button within it.

        const level1Row = screen.getByText('level1').closest('tr');
        // We can look for the button with aria-label "Expand" (since it starts collapsed)
        // But simply finding the button in the row is enough.
        const expandButton = level1Row?.querySelector('button');
        expect(expandButton).toBeInTheDocument();

        // Click to expand
        fireEvent.click(expandButton!);

        // Now level2 should be visible
        expect(screen.getByText('level2')).toBeInTheDocument();
    });
});
