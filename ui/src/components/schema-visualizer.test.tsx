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

    it('renders deep nested objects collapsed by default', () => {
        const schema = {
            type: 'object',
            properties: {
                address: {
                    type: 'object',
                    properties: {
                        details: {
                            type: 'object',
                            properties: {
                                zip: { type: 'string' }
                            }
                        }
                    },
                },
            },
        };

        render(<SchemaVisualizer schema={schema} />);

        // address (L0) is expanded.
        expect(screen.getByText('address')).toBeInTheDocument();
        // details (L1) is rendered because address is expanded.
        expect(screen.getByText('details')).toBeInTheDocument();

        // zip (L2) should NOT be visible because details (L1) is collapsed.
        expect(screen.queryByText('zip')).not.toBeInTheDocument();

        // Click expand on 'details' row.
        // We need to find the expand button. It's a button with ChevronRight.
        // It's in the row for 'details'.

        // The button is inside the cell for 'details'.
        // We can find the button by its role or simply by querying inside the row.
        const detailsCell = screen.getByText('details').closest('td');
        // The button is a sibling or parent? No, details is in a span inside a div.
        // <div className="flex items-center gap-2"><button>...</button><span>details</span></div>
        const expandButton = detailsCell?.querySelector('button');

        if (expandButton) {
            fireEvent.click(expandButton);
        } else {
             throw new Error("Expand button not found");
        }

        // Now zip should be visible
        expect(screen.getByText('zip')).toBeInTheDocument();
    });
});
