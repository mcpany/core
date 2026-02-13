/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { StackEditor } from '@/components/stacks/stack-editor';
import { vi } from 'vitest';

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
  const defaultProps = {
    initialValue: 'name: test-stack\nservices: []',
    onSave: vi.fn(),
    onCancel: vi.fn(),
  };

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('loads and displays configuration', async () => {
    render(<StackEditor {...defaultProps} />);

    await waitFor(() => {
      expect(screen.getByText('Stack Composer')).toBeInTheDocument();
      expect(screen.getByTestId('config-editor-mock')).toBeInTheDocument();
    });
  });

  it('validates YAML content', async () => {
    render(<StackEditor {...defaultProps} />);

    await waitFor(() => expect(screen.getByTestId('config-editor-mock')).toBeInTheDocument());
    const textarea = screen.getByTestId('config-editor-mock');

    // Valid YAML (empty services)
    fireEvent.change(textarea, { target: { value: 'key: value' } });
    await waitFor(() => {
        expect(screen.getByText('No services defined')).toBeInTheDocument();
    });

    // Invalid YAML
    fireEvent.change(textarea, { target: { value: 'key: "unclosed quote' } });

    await waitFor(() => {
         expect(screen.getByText('YAML Syntax Error')).toBeInTheDocument();
    });
  });

  it('toggles palette and visualizer', async () => {
    render(<StackEditor {...defaultProps} />);

    // Check initial state
    expect(screen.getByTitle('Toggle Palette')).toBeInTheDocument();
    // Since config is empty, visualizer shows "No services defined"
    expect(screen.getByText('No services defined')).toBeInTheDocument();
  });
});
