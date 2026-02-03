/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { render, screen } from '@testing-library/react';
import { describe, it, expect, vi } from 'vitest';
import { DiffViewer } from './diff-viewer';

// Mock Monaco Editor
vi.mock('@monaco-editor/react', () => ({
  DiffEditor: ({ original, modified }: { original: string, modified: string }) => (
    <div data-testid="diff-editor">
      <div data-testid="original">{original}</div>
      <div data-testid="modified">{modified}</div>
    </div>
  ),
}));

// Mock next-themes
vi.mock('next-themes', () => ({
  useTheme: () => ({ theme: 'light' }),
}));

describe('DiffViewer', () => {
  it('renders diff editor with correct content', () => {
    const original = 'key: value';
    const modified = 'key: newValue';

    render(<DiffViewer original={original} modified={modified} />);

    expect(screen.getByTestId('diff-editor')).toBeInTheDocument();
    expect(screen.getByTestId('original')).toHaveTextContent(original);
    expect(screen.getByTestId('modified')).toHaveTextContent(modified);
  });
});
