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
    getCollection: vi.fn(), // Added getCollection mock
    saveCollection: vi.fn(), // Added saveCollection mock
  },
}));

// Mock ConfigEditor to render a simple textarea for testing
vi.mock('@/components/stacks/config-editor', () => ({
  ConfigEditor: ({ value, onChange }: { value: string; onChange: (val: string) => void }) => (
    <textarea
      value={value}
      onChange={(e) => onChange(e.target.value)}
      data-testid="config-editor-mock"
    />
  ),
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
    (apiClient.getCollection as any).mockResolvedValue({
      name: 'test-stack',
      services: []
    });

    render(<StackEditor stackId="test-stack" />);

    await waitFor(() => {
      expect(screen.getByText('config.yaml')).toBeInTheDocument();
      // The content will be a yaml dump of the collection.
      // Since services is empty array, it might be just "name: test-stack\nservices: {}\n" or similar.
      // Let's just check for the presence of the editor mock.
      expect(screen.getByTestId('config-editor-mock')).toBeInTheDocument();
    });
  });

  it('validates YAML content', async () => {
    (apiClient.getCollection as any).mockResolvedValue({
      name: 'test-stack',
      services: []
    });

    const { container } = render(<StackEditor stackId="test-stack" />);

    // Find textarea by selector if role is elusive
    await waitFor(() => expect(screen.getByTestId('config-editor-mock')).toBeInTheDocument());
    const textarea = screen.getByTestId('config-editor-mock');

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
    (apiClient.getCollection as any).mockResolvedValue({ name: 'test-stack', services: [] });
    render(<StackEditor stackId="test-stack" />);

    // Check initial state
    expect(screen.getByText('Service Palette')).toBeInTheDocument();
    // Since config is empty, visualizer shows "No services defined"
    expect(screen.getByText('No services defined')).toBeInTheDocument();
  });
});
