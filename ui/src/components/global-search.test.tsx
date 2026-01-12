/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import React from 'react';
import { render, screen, fireEvent, waitFor, act } from '@testing-library/react';
import { GlobalSearch } from './global-search';
import { vi } from 'vitest';
import { apiClient } from '@/lib/client';

// Mock the API client
vi.mock('@/lib/client', () => ({
  apiClient: {
    listServices: vi.fn().mockResolvedValue([
        { id: 'srv1', name: 'service-1', version: '1.0', disable: false },
        { id: 'srv2', name: 'service-2', version: '2.0', disable: true }
    ]),
    listTools: vi.fn().mockResolvedValue({ tools: [{ name: 'tool-1', description: 'A test tool' }] }),
    listResources: vi.fn().mockResolvedValue({ resources: [{ name: 'res-1', uri: 'file:///tmp/res1' }] }),
    listPrompts: vi.fn().mockResolvedValue({ prompts: [{ name: 'prompt-1' }] }),
    setServiceStatus: vi.fn().mockResolvedValue({}),
  },
}));

// Mock useRouter
const mockPush = vi.fn();
vi.mock('next/navigation', () => ({
  useRouter: () => ({
    push: mockPush,
  }),
}));

// Mock useTheme
vi.mock('next-themes', () => ({
  useTheme: () => ({
    setTheme: vi.fn(),
  }),
}));

// Mock useToast
const mockToast = vi.fn();
vi.mock('@/hooks/use-toast', () => ({
  useToast: () => ({
    toast: mockToast,
  }),
}));

// Mock window location and clipboard
const originalLocation = window.location;
const mockReload = vi.fn();
const mockWriteText = vi.fn();

Object.defineProperty(window, 'location', {
  configurable: true,
  value: { ...originalLocation, reload: mockReload, href: 'http://localhost/' },
});

Object.defineProperty(navigator, 'clipboard', {
  value: { writeText: mockWriteText },
});


// ResizeObserver mock
global.ResizeObserver = class ResizeObserver {
    observe() {}
    unobserve() {}
    disconnect() {}
};

// Mock pointer capture methods
Element.prototype.setPointerCapture = () => {};
Element.prototype.releasePointerCapture = () => {};

describe('GlobalSearch', () => {
    beforeEach(() => {
        vi.clearAllMocks();
    });

    it('renders the search button', () => {
        render(<GlobalSearch />);
        expect(screen.getByText(/Search or type >/i)).toBeInTheDocument();
    });

    it('opens dialog on click and lists items', async () => {
        render(<GlobalSearch />);

        await act(async () => {
            fireEvent.click(screen.getByText(/Search or type >/i));
        });

        // Wait for data to load
        await waitFor(() => {
            expect(screen.getByText('service-1')).toBeInTheDocument();
            expect(screen.getByText('tool-1')).toBeInTheDocument();
            expect(screen.getByText('res-1')).toBeInTheDocument();
        });
    });

    it('navigates when item is selected', async () => {
        render(<GlobalSearch />);

        await act(async () => {
            fireEvent.click(screen.getByText(/Search or type >/i));
        });

        await waitFor(() => {
            expect(screen.getByText('service-1')).toBeInTheDocument();
        });

        // Click the navigation item (first occurrence of service-1 text might be title or item, assuming item text)
        // Since we added actions, 'service-1' appears twice (navigate and restart).
        // CommandItem children: Icon + Text.
        // We can select by text.
        const serviceItems = screen.getAllByText('service-1');
        // The first one is likely the navigation one based on render order.
        fireEvent.click(serviceItems[0]);

        expect(mockPush).toHaveBeenCalledWith('/services?id=srv1');
    });

    it('triggers restart action', async () => {
         render(<GlobalSearch />);

        await act(async () => {
            fireEvent.click(screen.getByText(/Search or type >/i));
        });

        await waitFor(() => {
            expect(screen.getByText('Restart service-1')).toBeInTheDocument();
        });

        fireEvent.click(screen.getByText('Restart service-1'));

        // Should call setServiceStatus twice (disable then enable)
        await waitFor(() => {
             expect(apiClient.setServiceStatus).toHaveBeenCalledWith('service-1', true);
        });

        // We can't easily wait for the timeout in the component without fake timers,
        // but we can check if toast was called.
        expect(mockToast).toHaveBeenCalledWith(expect.objectContaining({
            title: "Restarting Service"
        }));
    });

    it('triggers copy URI action', async () => {
         render(<GlobalSearch />);

        await act(async () => {
            fireEvent.click(screen.getByText(/Search or type >/i));
        });

        await waitFor(() => {
            expect(screen.getByText('Copy URI: res-1')).toBeInTheDocument();
        });

        fireEvent.click(screen.getByText('Copy URI: res-1'));

        expect(mockWriteText).toHaveBeenCalledWith('file:///tmp/res1');
        expect(mockToast).toHaveBeenCalledWith(expect.objectContaining({
            title: "Copied to clipboard"
        }));
    });

    it('triggers reload window action', async () => {
         render(<GlobalSearch />);

        await act(async () => {
            fireEvent.click(screen.getByText(/Search or type >/i));
        });

        await waitFor(() => {
            expect(screen.getByText('Reload Window')).toBeInTheDocument();
        });

        fireEvent.click(screen.getByText('Reload Window'));

        expect(mockReload).toHaveBeenCalled();
    });
});
