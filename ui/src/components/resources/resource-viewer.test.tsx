/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { render, screen } from '@testing-library/react';
import { describe, it, expect, vi } from 'vitest';
import { ResourceViewer } from './resource-viewer';
import { ResourceContent } from '@/lib/client';

// Mock syntax highlighter since it might cause issues in JSDOM
vi.mock('react-syntax-highlighter/dist/esm/light', () => {
    const MockHighlighter = ({ children }: { children: React.ReactNode }) => <pre data-testid="code-block">{children}</pre>;
    // Mock static methods like registerLanguage
    MockHighlighter.registerLanguage = vi.fn();
    return {
        default: MockHighlighter
    };
});

describe('ResourceViewer', () => {
  it('renders loading state', () => {
    render(<ResourceViewer content={null} loading={true} />);
    expect(screen.getByText('Loading content...')).toBeInTheDocument();
  });

  it('renders empty state', () => {
    render(<ResourceViewer content={null} loading={false} />);
    expect(screen.getByText('Select a resource to view its content.')).toBeInTheDocument();
  });

  it('renders JSON content', () => {
    const content: ResourceContent = {
        uri: 'file:///config.json',
        mimeType: 'application/json',
        text: '{"foo": "bar"}'
    };
    render(<ResourceViewer content={content} loading={false} />);
    expect(screen.getByTestId('code-block')).toHaveTextContent('{"foo": "bar"}');
  });

  it('renders Markdown content', () => {
    const content: ResourceContent = {
        uri: 'file:///README.md',
        mimeType: 'text/markdown',
        text: '# Hello'
    };
    render(<ResourceViewer content={content} loading={false} />);
    expect(screen.getByTestId('code-block')).toHaveTextContent('# Hello');
  });

  it('renders Plain text content', () => {
     const content: ResourceContent = {
        uri: 'file:///log.txt',
        mimeType: 'text/plain',
        text: 'Just some logs'
    };
    render(<ResourceViewer content={content} loading={false} />);
    expect(screen.getByTestId('code-block')).toHaveTextContent('Just some logs');
  });
});
