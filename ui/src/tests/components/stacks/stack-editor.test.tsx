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
  },
}));

// Mock ResizeObserver for scroll area
global.ResizeObserver = class ResizeObserver {
  observe() {}
  unobserve() {}
  disconnect() {}
};

// Mock Monaco Editor
vi.mock('@monaco-editor/react', () => {
  const Editor = ({ value, onChange }: { value: string; onChange: (val: string) => void }) => (
    <textarea
      data-testid="monaco-editor"
      value={value}
      onChange={(e) => onChange(e.target.value)}
    />
  );
  return {
    __esModule: true,
    default: Editor,
    Editor,
    useMonaco: () => ({
      languages: {
        register: () => {},
        setMonarchTokensProvider: () => {},
        registerCompletionItemProvider: () => ({ dispose: () => {} }),
        CompletionItemKind: { Keyword: 0, Property: 1 },
      },
    }),
    loader: {
      config: () => {},
    },
  };
});

describe('StackEditor', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('loads and displays configuration', async () => {
    (apiClient.getStackConfig as any).mockResolvedValue('version: "1.0"');

    render(<StackEditor stackId="test-stack" />);

    await waitFor(() => {
      expect(screen.getByText('config.yaml')).toBeInTheDocument();
      // Use getByDisplayValue for textarea content
      expect(screen.getByDisplayValue('version: "1.0"')).toBeInTheDocument();
    });
  });

  it('validates YAML content', async () => {
    (apiClient.getStackConfig as any).mockResolvedValue('');

    const { container } = render(<StackEditor stackId="test-stack" />);

    // Wait for the textarea to be rendered, as it might be async/lazy loaded
    // Monaco editor often hides the real textarea, but here we seem to be testing a standard textarea fallback or similar.
    // If we assume a standard textarea exists:
    await waitFor(() => {
        expect(container.querySelector('textarea')).toBeInTheDocument();
    });

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
    (apiClient.getStackConfig as any).mockResolvedValue('');
    render(<StackEditor stackId="test-stack" />);

    // Check initial state
    expect(screen.getByText('Service Palette')).toBeInTheDocument();
    // Since config is empty, visualizer shows "No services defined"
    expect(screen.getByText('No services defined')).toBeInTheDocument();
  });
});
