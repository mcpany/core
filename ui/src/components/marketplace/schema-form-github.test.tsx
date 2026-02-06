/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import React from 'react';
import { render, screen } from '@testing-library/react';
import { SchemaForm } from './schema-form';
import { vi } from 'vitest';

// Mock the GitHubTokenField to verify it's rendered
vi.mock('@/components/marketplace/fields/github-token-field', () => ({
  GitHubTokenField: (props: any) => <div data-testid="github-token-field" data-id={props.id}>{props.title}</div>
}));

describe('SchemaForm with GitHubTokenField', () => {
    const mockSchema = {
        type: "object",
        properties: {
            "GITHUB_PERSONAL_ACCESS_TOKEN": {
                type: "string",
                title: "My GitHub Token",
                description: "Enter PAT"
            },
            "other_field": {
                type: "string",
                title: "Other"
            }
        },
        required: ["GITHUB_PERSONAL_ACCESS_TOKEN"]
    };

    it('renders GitHubTokenField for specific key', () => {
        const onChange = vi.fn();
        render(<SchemaForm schema={mockSchema} value={{}} onChange={onChange} />);

        const customField = screen.getByTestId('github-token-field');
        expect(customField).toBeInTheDocument();
        expect(customField).toHaveTextContent("My GitHub Token");
    });

    it('renders normal input for other fields', () => {
        const onChange = vi.fn();
        render(<SchemaForm schema={mockSchema} value={{}} onChange={onChange} />);

        expect(screen.getByLabelText("Other")).toBeInTheDocument();
    });
});
