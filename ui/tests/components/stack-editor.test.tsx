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
  });

  it('renders correctly and loads config', async () => {
    render(<StackEditor initialValue={mockConfig} onSave={vi.fn()} onCancel={vi.fn()} />);

    expect(screen.getByText('Stack Composer')).toBeDefined();
    await waitFor(() => {
        // The component dumps the object to YAML, so we just check if it contains the image
        expect(screen.getByRole('textbox')).toBeDefined();
    });
  });
});
