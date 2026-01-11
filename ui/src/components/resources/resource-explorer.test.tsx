/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { describe, it, expect, vi } from 'vitest';
import { ResourceExplorer } from './resource-explorer';
import { apiClient } from '@/lib/client';
import { MOCK_RESOURCES } from '@/mocks/resources';

// Mock dependencies
vi.mock('@/lib/client', () => ({
  apiClient: {
    listResources: vi.fn(),
    readResource: vi.fn(),
  },
}));

// Mock syntax highlighter since it might cause issues in JSDOM
vi.mock('react-syntax-highlighter', () => ({
  default: ({ children }: { children: React.ReactNode }) => <pre data-testid="code-block">{children}</pre>,
}));

describe('ResourceExplorer', () => {
  it('renders loading state initially', async () => {
    // @ts-ignore
    apiClient.listResources.mockResolvedValueOnce({ resources: [] });

    render(<ResourceExplorer />);
    // Initial render might show loading or empty state depending on how fast useEffect runs
    // Here we check if it calls the API
    expect(apiClient.listResources).toHaveBeenCalled();
  });

  it('renders list of resources', async () => {
    // @ts-ignore
    apiClient.listResources.mockResolvedValueOnce({ resources: MOCK_RESOURCES });

    render(<ResourceExplorer />);

    await waitFor(() => {
        expect(screen.getByText('config.json')).toBeInTheDocument();
        expect(screen.getByText('README.md')).toBeInTheDocument();
    });
  });

  it('filters resources based on search query', async () => {
    // @ts-ignore
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
    // @ts-ignore
    apiClient.listResources.mockResolvedValueOnce({ resources: MOCK_RESOURCES });
    // @ts-ignore
    apiClient.readResource.mockResolvedValueOnce({
        contents: [{ uri: 'file:///app/config.json', mimeType: 'application/json', text: '{"test": true}' }]
    });

    render(<ResourceExplorer />);

    await waitFor(() => {
        expect(screen.getByText('config.json')).toBeInTheDocument();
    });

    fireEvent.click(screen.getByText('config.json'));

    await waitFor(() => {
        expect(apiClient.readResource).toHaveBeenCalledWith('file:///app/config.json');
        expect(screen.getByTestId('code-block')).toHaveTextContent('{"test": true}');
    });
  });
});
