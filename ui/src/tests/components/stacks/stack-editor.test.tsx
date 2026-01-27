/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { StackEditor } from '@/components/stacks/stack-editor';
import { apiClient } from '@/lib/client';
import { vi } from 'vitest';

// Mock the API client
vi.mock('@/lib/client', () => ({
  apiClient: {
    getStackConfig: vi.fn(),
    saveStackConfig: vi.fn(),
    getCollection: vi.fn(),
    saveCollection: vi.fn(),
  },
}));

// Mock Monaco Editor
vi.mock('@monaco-editor/react', () => ({
  default: ({ value, onChange }: any) => (
    <textarea
      data-testid="monaco-editor"
      value={value}
      onChange={(e) => onChange(e.target.value)}
    />
  ),
  useMonaco: () => null,
}));

// Mock ResizeObserver for scroll area
global.ResizeObserver = class ResizeObserver {
  observe() {}
  unobserve() {}
  disconnect() {}
};

describe('StackEditor', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('loads and displays configuration', async () => {
    (apiClient.getCollection as any).mockResolvedValue({ version: "1.0" });

    render(<StackEditor stackId="test-stack" />);

    await waitFor(() => {
      expect(screen.getByText('config.yaml')).toBeInTheDocument();
      // Use getByDisplayValue for textarea content
      // js-yaml might quote strings differently, so we use regex
      expect(screen.getByDisplayValue(/version: ['"]?1\.0['"]?/)).toBeInTheDocument();
    });
  });

  it('validates YAML content', async () => {
    (apiClient.getCollection as any).mockResolvedValue({});

    const { container } = render(<StackEditor stackId="test-stack" />);

    // Find textarea by selector if role is elusive
    const textarea = container.querySelector('textarea');
    if (!textarea) throw new Error('Textarea not found');

    // Valid YAML
    fireEvent.change(textarea, { target: { value: 'key: value' } });
    await waitFor(() => {
        expect(screen.getByText('Valid YAML')).toBeInTheDocument();
    });

    // Invalid YAML
    fireEvent.change(textarea, { target: { value: 'key: "unclosed quote' } });

    await waitFor(() => {
         expect(screen.getByText('Invalid YAML')).toBeInTheDocument();
    });
  });

  it('toggles palette and visualizer', async () => {
    (apiClient.getCollection as any).mockResolvedValue({});
    render(<StackEditor stackId="test-stack" />);

    // Check initial state
    expect(screen.getByText('Service Palette')).toBeInTheDocument();
    // Since config is empty, visualizer shows "No services defined"
    expect(screen.getByText('No services defined')).toBeInTheDocument();
  });
});
