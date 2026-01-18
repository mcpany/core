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

// Mock Monaco Editor
vi.mock('@monaco-editor/react', () => ({
  default: ({ value, onChange }: { value: string, onChange: (value: string) => void }) => {
    return (
      <textarea
        data-testid="monaco-editor-mock"
        value={value}
        onChange={(e) => onChange(e.target.value)}
      />
    );
  },
  useMonaco: vi.fn().mockReturnValue({
      languages: {
          register: vi.fn(),
          registerCompletionItemProvider: vi.fn().mockReturnValue({ dispose: vi.fn() }),
          setMonarchTokensProvider: vi.fn(),
          json: {
              modeConfiguration: {}
          }
      },
      editor: {
          defineTheme: vi.fn(),
      }
  }),
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
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    (apiClient.getStackConfig as any).mockResolvedValue('version: "1.0"');

    render(<StackEditor stackId="test-stack" />);

    await waitFor(() => {
      expect(screen.getByText('config.yaml')).toBeInTheDocument();
      // Use getByDisplayValue for textarea content
      expect(screen.getByDisplayValue('version: "1.0"')).toBeInTheDocument();
    });
  });

  it('validates YAML content', async () => {
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    (apiClient.getStackConfig as any).mockResolvedValue('');

    render(<StackEditor stackId="test-stack" />);

    // Find textarea by selector if role is elusive
    await waitFor(() => {
        expect(screen.getByTestId('monaco-editor-mock')).toBeInTheDocument();
    });
    const textarea = screen.getByTestId('monaco-editor-mock');

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
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    (apiClient.getStackConfig as any).mockResolvedValue('');
    render(<StackEditor stackId="test-stack" />);

    // Check initial state
    expect(screen.getByText('Service Palette')).toBeInTheDocument();
    // Since config is empty, visualizer shows "No services defined"
    expect(screen.getByText('No services defined')).toBeInTheDocument();
  });
});
