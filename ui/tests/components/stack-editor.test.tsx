/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */


import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { StackEditor } from "../../src/components/stacks/stack-editor";
import { apiClient } from "../../src/lib/client";

// Mock the ConfigEditor component because Monaco is hard to test in JSDOM
vi.mock('./config-editor', () => ({
  ConfigEditor: ({ value, onChange }: { value: string, onChange: (v: string) => void }) => (
    <textarea
      role="textbox"
      value={value}
      onChange={(e) => onChange(e.target.value)}
    />
  ),
}));

// Mock the apiClient
vi.mock('../../src/lib/client', () => ({
  apiClient: {
    getCollection: vi.fn(),
    saveCollection: vi.fn(),
  },
}));


describe('StackEditor', () => {
  const mockStackId = 'test-stack';
  const mockConfig = 'version: "1.0"\nservices:\n  test:\n    image: test/image';
  const mockCollection = {
    name: mockStackId,
    services: [
      { name: 'test', image: 'test/image' }
    ]
  };

  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(apiClient.getCollection).mockResolvedValue(mockCollection);
    vi.mocked(apiClient.saveCollection).mockResolvedValue({});
  });

  it('renders correctly and loads config', async () => {
    render(<StackEditor stackId={mockStackId} />);

    expect(screen.getByText('config.yaml')).toBeDefined();
    await waitFor(() => {
        // The component dumps the object to YAML, so we just check if it contains the image
        expect(screen.getByText(/test\/image/)).toBeDefined();
    });
  });

  it('validates valid YAML', async () => {
    render(<StackEditor stackId={mockStackId} />);

    await waitFor(() => {
         expect(screen.getByText('Valid YAML')).toBeDefined();
    });
  });

  it('validates valid YAML', async () => {
    render(<StackEditor stackId={mockStackId} />);

    await waitFor(() => {
         expect(screen.getByText('Valid YAML')).toBeDefined();
    });
  });
});
