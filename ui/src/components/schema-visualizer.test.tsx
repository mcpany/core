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
});
