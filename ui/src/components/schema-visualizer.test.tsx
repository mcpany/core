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

    it('collapses deep nodes (>1 level) by default', () => {
        const schema = {
            type: 'object',
            properties: {
                level1: {
                    type: 'object',
                    properties: {
                        level2: {
                            type: 'object',
                            properties: {
                                level3: { type: 'string' }
                            }
                        }
                    }
                }
            }
        };

        render(<SchemaVisualizer schema={schema} />);

        // Level 1 visible (root prop)
        expect(screen.getByText('level1')).toBeInTheDocument();

        // Level 2 visible (child of level 1)
        // Level 1 is level=0 in the loop? No.
        // SchemaVisualizer -> SchemaNode(level=0, name="level1")
        //   -> expands because 0 < 2
        //   -> renders children: SchemaNode(level=1, name="level2")
        //      -> expands because 1 < 2
        //      -> renders children: SchemaNode(level=2, name="level3") -- WAIT

        // Let's trace carefully:
        // Root loop (SchemaVisualizer): renders SchemaNode for "level1" with level=0.
        // SchemaNode(level=0): expanded=true (0<2). Renders children.
        // Children of level1: "level2". Renders SchemaNode for "level2" with level=1.
        // SchemaNode(level=1): expanded=true (1<2). Renders children.
        // Children of level2: "level3". Renders SchemaNode for "level3" with level=2.
        // SchemaNode(level=2): expanded=FALSE (2<2 is false). Renders ITSELF (the row "level3"), but NOT ITS CHILDREN.

        // So "level3" IS visible as a row.
        // But if "level3" had children, THEY would be hidden.

        expect(screen.getByText('level2')).toBeInTheDocument();
        expect(screen.getByText('level3')).toBeInTheDocument();

        // Let's go deeper to verify the collapse.
        // We need level4.
    });

    it('collapses very deep nodes (>2 levels) by default', () => {
         const schema = {
            type: 'object',
            properties: {
                l1: {
                    type: 'object',
                    properties: {
                        l2: {
                            type: 'object',
                            properties: {
                                l3: {
                                    type: 'object',
                                    properties: {
                                        l4: { type: 'string' }
                                    }
                                }
                            }
                        }
                    }
                }
            }
        };

        render(<SchemaVisualizer schema={schema} />);

        expect(screen.getByText('l1')).toBeInTheDocument(); // Level 0
        expect(screen.getByText('l2')).toBeInTheDocument(); // Level 1
        expect(screen.getByText('l3')).toBeInTheDocument(); // Level 2. Rendered, but collapsed?

        // SchemaNode(l3, level=2). expanded=false.
        // It renders the row for "l3".
        // It does NOT render children of l3.
        // Child of l3 is l4.

        expect(screen.queryByText('l4')).not.toBeInTheDocument();
    });
});
