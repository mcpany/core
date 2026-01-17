/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { ResourcePreviewModal } from './resource-preview-modal';
import { apiClient, ResourceDefinition, ResourceContent } from '@/lib/client';
import { useToast } from '@/hooks/use-toast';

// Mock dependencies
vi.mock('@/lib/client', () => ({
  apiClient: {
    readResource: vi.fn(),
  },
}));

vi.mock('@/hooks/use-toast', () => ({
  useToast: vi.fn(() => ({ toast: vi.fn() })),
}));

// Mock syntax highlighter
vi.mock('react-syntax-highlighter/dist/esm/light', () => {
    const MockHighlighter = ({ children }: { children: React.ReactNode }) => <pre data-testid="code-block">{children}</pre>;
    MockHighlighter.registerLanguage = vi.fn();
    return { default: MockHighlighter };
});

// Mock URL.createObjectURL
global.URL.createObjectURL = vi.fn(() => 'blob:mock-url');
global.URL.revokeObjectURL = vi.fn();

// Mock ResizeObserver for Radix Dialog
global.ResizeObserver = class ResizeObserver {
  observe() {}
  unobserve() {}
  disconnect() {}
};

describe('ResourcePreviewModal', () => {
  const mockResource: ResourceDefinition = {
    uri: 'file:///test.json',
    name: 'test.json',
    mimeType: 'application/json',
  };

  const mockContent: ResourceContent = {
    uri: 'file:///test.json',
    mimeType: 'application/json',
    text: '{"foo": "bar"}',
  };

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders nothing when closed', () => {
    render(
      <ResourcePreviewModal
        isOpen={false}
        onClose={vi.fn()}
        resource={mockResource}
      />
    );
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
  });

  it('renders content provided in initialContent', () => {
    render(
      <ResourcePreviewModal
        isOpen={true}
        onClose={vi.fn()}
        resource={mockResource}
        initialContent={mockContent}
      />
    );
    expect(screen.getByText('test.json')).toBeInTheDocument();
    expect(screen.getByTestId('code-block')).toHaveTextContent('{"foo": "bar"}');
    expect(apiClient.readResource).not.toHaveBeenCalled();
  });

  it('fetches content when initialContent is not provided', async () => {
    // @ts-expect-error Mocking
    apiClient.readResource.mockResolvedValueOnce({ contents: [mockContent] });

    render(
      <ResourcePreviewModal
        isOpen={true}
        onClose={vi.fn()}
        resource={mockResource}
      />
    );

    expect(screen.getByText('Loading content...')).toBeInTheDocument();

    await waitFor(() => {
        expect(screen.getByTestId('code-block')).toHaveTextContent('{"foo": "bar"}');
    });

    expect(apiClient.readResource).toHaveBeenCalledWith(mockResource.uri);
  });

  it('fetches content when initialContent uri does not match resource uri', async () => {
    // @ts-expect-error Mocking
    apiClient.readResource.mockResolvedValueOnce({ contents: [mockContent] });

    render(
      <ResourcePreviewModal
        isOpen={true}
        onClose={vi.fn()}
        resource={mockResource}
        initialContent={{ ...mockContent, uri: 'other' }}
      />
    );

    await waitFor(() => {
        expect(apiClient.readResource).toHaveBeenCalledWith(mockResource.uri);
    });
  });

  it('handles fetch error', async () => {
     // @ts-expect-error Mocking
    apiClient.readResource.mockRejectedValueOnce(new Error('Fetch failed'));
    const mockToast = vi.fn();
    // @ts-expect-error Mocking
    vi.mocked(useToast).mockReturnValue({ toast: mockToast });

    render(
      <ResourcePreviewModal
        isOpen={true}
        onClose={vi.fn()}
        resource={mockResource}
      />
    );

    await waitFor(() => {
        expect(mockToast).toHaveBeenCalledWith(expect.objectContaining({
            title: 'Error',
            variant: 'destructive'
        }));
    });
  });

  it('calls onClose when close button is clicked', () => {
    const onClose = vi.fn();
    render(
      <ResourcePreviewModal
        isOpen={true}
        onClose={onClose}
        resource={mockResource}
        initialContent={mockContent}
      />
    );

    const closeButton = screen.getByTitle('Close');
    fireEvent.click(closeButton);
    expect(onClose).toHaveBeenCalled();
  });

  it('handles download for binary content', async () => {
    const binaryContent: ResourceContent = {
        uri: 'file:///image.png',
        mimeType: 'image/png',
        blob: 'AQID' // 0x01, 0x02, 0x03
    };

    const mockResourceImg = { ...mockResource, name: 'image.png', mimeType: 'image/png', uri: 'file:///image.png' };

    // Mock document.createElement('a') and click
    const clickMock = vi.fn();
    const linkMock = { href: '', download: '', click: clickMock } as unknown as HTMLAnchorElement;

    const originalCreateElement = document.createElement.bind(document);
    vi.spyOn(document, 'createElement').mockImplementation((tagName: string, options?: ElementCreationOptions) => {
        if (tagName === 'a') {
            return linkMock;
        }
        return originalCreateElement(tagName, options);
    });

    render(
        <ResourcePreviewModal
            isOpen={true}
            onClose={vi.fn()}
            resource={mockResourceImg}
            initialContent={binaryContent}
        />
    );

    const downloadBtn = screen.getByRole('button', { name: /download/i });
    fireEvent.click(downloadBtn);

    expect(global.URL.createObjectURL).toHaveBeenCalled();
    // Check if blob was created correctly
    // @ts-expect-error Mocking
    const blob = global.URL.createObjectURL.mock.calls[global.URL.createObjectURL.mock.calls.length - 1][0];
    expect(blob).toBeInstanceOf(Blob);
    expect(blob.type).toBe('image/png');
    expect(blob.size).toBe(3);

    expect(linkMock.href).toBe('blob:mock-url');
    expect(linkMock.download).toBe('image.png');
    expect(clickMock).toHaveBeenCalled();
  });
});
