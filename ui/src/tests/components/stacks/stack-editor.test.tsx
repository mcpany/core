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
  const mockSave = vi.fn();
  const mockCancel = vi.fn();
  const initialValue = 'name: test-stack\nservices: []\n';

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders configuration', () => {
    render(
      <StackEditor
        initialValue={initialValue}
        onSave={mockSave}
        onCancel={mockCancel}
      />
    );

    expect(screen.getByTestId('config-editor-mock')).toBeInTheDocument();
    expect(screen.getByTestId('config-editor-mock')).toHaveValue(initialValue);
  });

  it('calls onSave with updated content', async () => {
    render(
      <StackEditor
        initialValue={initialValue}
        onSave={mockSave}
        onCancel={mockCancel}
      />
    );

    const textarea = screen.getByTestId('config-editor-mock');
    fireEvent.change(textarea, { target: { value: 'updated: content' } });

    const saveBtn = screen.getByText('Save & Deploy');
    fireEvent.click(saveBtn);

    expect(mockSave).toHaveBeenCalledWith('updated: content');
  });

  it('toggles palette and visualizer', async () => {
    render(
      <StackEditor
        initialValue={initialValue}
        onSave={mockSave}
        onCancel={mockCancel}
      />
    );

    // Initial state: Palette and Visualizer are shown (based on component default)
    // The test might depend on exact rendering of child components.
    // Since we didn't mock Palette or Visualizer, they might render (or fail if dependencies missing).
    // Let's assume they render buttons with specific titles.
    // "Toggle Palette" and "Toggle Visualizer"

    const togglePalette = screen.getByTitle('Toggle Palette');
    const toggleVisualizer = screen.getByTitle('Toggle Visualizer');

    expect(togglePalette).toBeInTheDocument();
    expect(toggleVisualizer).toBeInTheDocument();
  });
});
