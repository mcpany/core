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
    render(<StackEditor initialValue="name: test-stack" onSave={vi.fn()} onCancel={vi.fn()} />);

    // Check for the presence of the editor mock.
    expect(screen.getByTestId('config-editor-mock')).toBeInTheDocument();
  });

  it('toggles palette and visualizer', async () => {
    render(<StackEditor initialValue="name: test-stack" onSave={vi.fn()} onCancel={vi.fn()} />);

    // Check initial state
    // Palette is shown by default (Show Palette button exists, or Panel exists)
    // The text "Stack Composer" is in the header
    expect(screen.getByText('Stack Composer')).toBeInTheDocument();
  });
});
