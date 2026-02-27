/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import { StackEditor } from "../../src/components/stacks/stack-editor";

// Mock the ConfigEditor component because Monaco is hard to test in JSDOM
vi.mock('../../src/components/stacks/config-editor', () => ({
  ConfigEditor: ({ value, onChange }: { value: string, onChange: (v: string) => void }) => (
    <textarea
      role="textbox"
      value={value}
      onChange={(e) => onChange(e.target.value)}
    />
  ),
}));

describe('StackEditor', () => {
  const mockConfig = 'version: "1.0"\nservices:\n  test:\n    image: test/image';

  const mockProps = {
      initialValue: mockConfig,
      onSave: vi.fn(),
      onCancel: vi.fn(),
  };

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders correctly and loads config', async () => {
    render(<StackEditor {...mockProps} />);

    // We check for the mocked editor value
    const editor = screen.getByRole('textbox');
    expect(editor).toBeDefined();
    expect(editor.textContent).toBe(mockConfig);
  });
});
