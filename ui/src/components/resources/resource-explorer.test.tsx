/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { describe, it, expect, vi } from 'vitest';
import { ResourceExplorer } from './resource-explorer';
import { apiClient } from '@/lib/client';

const MOCK_RESOURCES = [
    { uri: 'file:///app/config.json', name: 'config.json', mimeType: 'application/json' },
    { uri: 'file:///app/README.md', name: 'README.md', mimeType: 'text/markdown' },
    { uri: 'postgres://db/users', name: 'users', mimeType: 'application/sql' }
];

// Mock dependencies
vi.mock('@/lib/client', () => ({
  apiClient: {
    listResources: vi.fn(),
    readResource: vi.fn(),
  },
}));

// Mock syntax highlighter since it might cause issues in JSDOM
vi.mock('react-syntax-highlighter/dist/esm/light', () => {
/**
 * MockHighlighter component.
 * @param props - The component props.
 * @param props.children - The child components.
 * @returns The rendered component.
 */
    const MockHighlighter = ({ children }: { children: React.ReactNode }) => <pre data-testid="code-block">{children}</pre>;
    // Mock static methods like registerLanguage
    MockHighlighter.registerLanguage = vi.fn();
    return {
        default: MockHighlighter
    };
});

describe('ResourceExplorer', () => {
  it('renders loading state initially', async () => {
    // @ts-expect-error Mocking partial implementation
    apiClient.listResources.mockResolvedValueOnce({ resources: [] });

    render(<ResourceExplorer />);
    // Initial render might show loading or empty state depending on how fast useEffect runs
    // Here we check if it calls the API
    expect(apiClient.listResources).toHaveBeenCalled();
  });

  it('renders list of resources', async () => {
    // @ts-expect-error Mocking partial implementation
    apiClient.listResources.mockResolvedValueOnce({ resources: MOCK_RESOURCES });

    render(<ResourceExplorer />);

    await waitFor(() => {
        expect(screen.getByText('config.json')).toBeInTheDocument();
        expect(screen.getByText('README.md')).toBeInTheDocument();
    });
  });

  it('filters resources based on search query', async () => {
    // @ts-expect-error Mocking partial implementation
    apiClient.listResources.mockResolvedValueOnce({ resources: MOCK_RESOURCES });

    render(<ResourceExplorer />);

    await waitFor(() => {
        expect(screen.getByText('config.json')).toBeInTheDocument();
    });

    const searchInput = screen.getByPlaceholderText('Search resources...');
    fireEvent.change(searchInput, { target: { value: 'json' } });

    expect(screen.getByText('config.json')).toBeInTheDocument();
    expect(screen.queryByText('README.md')).not.toBeInTheDocument();
  });

  it('selects a resource and shows content', async () => {
    // @ts-expect-error Mocking partial implementation
    apiClient.listResources.mockResolvedValueOnce({ resources: MOCK_RESOURCES });
    // @ts-expect-error Mocking partial implementation
    apiClient.readResource.mockResolvedValueOnce({
        // Use text/plain so it falls back to syntax highlighter code-block
        contents: [{ uri: 'file:///app/README.md', mimeType: 'text/plain', text: 'raw data' }]
    });

    render(<ResourceExplorer />);

    await waitFor(() => {
        expect(screen.getByText('README.md')).toBeInTheDocument();
    });

    fireEvent.click(screen.getByText('README.md'));

    await waitFor(() => {
        expect(apiClient.readResource).toHaveBeenCalledWith('file:///app/README.md');
        expect(screen.getByTestId('code-block')).toHaveTextContent('raw data');
    });
  });
});
