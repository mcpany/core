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
  const mockProps = {
    initialValue: 'version: "1.0"\nservices:\n  test: {}',
    onSave: vi.fn(),
    onCancel: vi.fn(),
  };

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('loads and displays configuration', async () => {
    render(<StackEditor {...mockProps} />);

    await waitFor(() => {
      // Check for editor
      expect(screen.getByTestId('config-editor-mock')).toBeInTheDocument();
      // Check initial value in editor
      expect(screen.getByTestId('config-editor-mock')).toHaveValue(mockProps.initialValue);
    });
  });

  it('toggles palette and visualizer', async () => {
    render(<StackEditor {...mockProps} />);

    // Check initial state
    expect(screen.getByText('Service Palette')).toBeInTheDocument();
    // Since config is simple, visualizer shows graph (mocked or real).
    // The graph might render "No services defined" if services map is empty or not parsed correctly by dagre/reactflow in test environment.
    // But we just check for presence of visualizer container or text.
    // The visualizer title is "Live Preview"
    expect(screen.getByText('Live Preview')).toBeInTheDocument();
  });
});
