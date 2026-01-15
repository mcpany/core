/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { describe, it, expect, vi } from 'vitest';
import { ResourcePreviewModal } from './resource-preview-modal';
import { apiClient } from '@/lib/client';

// Mock dependencies
vi.mock('@/lib/client', () => ({
  apiClient: {
    readResource: vi.fn(),
  },
}));

// Mock ResourceViewer
vi.mock('./resource-viewer', () => ({
  ResourceViewer: ({ loading, content }: { loading: boolean, content: any }) => (
    <div data-testid="resource-viewer">
      {loading ? 'Loading...' : (content ? content.text : 'No content')}
    </div>
  ),
}));

// Mock Dialog components (radix-ui often needs mocks in JSDOM if not fully polyfilled, but let's try basic rendering)
// Actually, shadcn/ui Dialog uses Portal, which renders outside the container.
// Testing Library's `screen` handles this, but we need to ensure DialogContent is rendered.

describe('ResourcePreviewModal', () => {
  const mockResource = {
    uri: 'file:///test.json',
    name: 'test.json',
    mimeType: 'application/json'
  };

  const mockContent = {
    uri: 'file:///test.json',
    mimeType: 'application/json',
    text: '{"test": true}'
  };

  it('renders nothing when not open', () => {
    render(
      <ResourcePreviewModal
        isOpen={false}
        onOpenChange={vi.fn()}
        resource={mockResource}
      />
    );
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
  });

  it('renders modal with initial content', () => {
    render(
      <ResourcePreviewModal
        isOpen={true}
        onOpenChange={vi.fn()}
        resource={mockResource}
        initialContent={mockContent}
      />
    );
    expect(screen.getByRole('dialog')).toBeInTheDocument();
    expect(screen.getByText('test.json')).toBeInTheDocument();
    expect(screen.getByTestId('resource-viewer')).toHaveTextContent('{"test": true}');
    expect(apiClient.readResource).not.toHaveBeenCalled();
  });

  it('fetches content when not provided', async () => {
    // @ts-expect-error Mocking partial implementation
    apiClient.readResource.mockResolvedValueOnce({
        contents: [mockContent]
    });

    render(
      <ResourcePreviewModal
        isOpen={true}
        onOpenChange={vi.fn()}
        resource={mockResource}
      />
    );

    expect(screen.getByRole('dialog')).toBeInTheDocument();
    // It might show loading first
    expect(screen.getByTestId('resource-viewer')).toBeInTheDocument();

    await waitFor(() => {
        expect(apiClient.readResource).toHaveBeenCalledWith(mockResource.uri);
        expect(screen.getByTestId('resource-viewer')).toHaveTextContent('{"test": true}');
    });
  });
});
