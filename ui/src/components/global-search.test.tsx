/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { GlobalSearch } from './global-search';
import { apiClient } from '@/lib/client';
import { useRouter } from 'next/navigation';
import { useTheme } from 'next-themes';

// Mock the dependent modules
vi.mock('@/lib/client', () => ({
  apiClient: {
    listServices: vi.fn(),
    listTools: vi.fn(),
    listResources: vi.fn(),
    listPrompts: vi.fn(),
  },
}));

vi.mock('next/navigation', () => ({
  useRouter: vi.fn(),
}));

vi.mock('next-themes', () => ({
  useTheme: vi.fn(),
}));

// Mock ResizeObserver for Radix UI
class ResizeObserverMock {
  observe() {}
  unobserve() {}
  disconnect() {}
}
global.ResizeObserver = ResizeObserverMock;

describe('GlobalSearch Component', () => {
  const mockPush = vi.fn();
  const mockSetTheme = vi.fn();

  beforeEach(() => {
    vi.clearAllMocks();
    (useRouter as any).mockReturnValue({ push: mockPush });
    (useTheme as any).mockReturnValue({ setTheme: mockSetTheme });

    // Default mock implementations
    (apiClient.listServices as any).mockResolvedValue({ services: [] });
    (apiClient.listTools as any).mockResolvedValue({ tools: [] });
    (apiClient.listResources as any).mockResolvedValue({ resources: [] });
    (apiClient.listPrompts as any).mockResolvedValue({ prompts: [] });
  });

  it('renders the search button initially', () => {
    render(<GlobalSearch />);
    expect(screen.getAllByText(/Search/i).length).toBeGreaterThan(0);
    expect(screen.getByText('⌘')).toBeInTheDocument();
    expect(screen.getByText('K')).toBeInTheDocument();
  });

  it('opens the command dialog when clicked', async () => {
    render(<GlobalSearch />);
    const button = screen.getByRole('button');
    fireEvent.click(button);

    await waitFor(() => {
        expect(screen.getByPlaceholderText('Type a command or search...')).toBeInTheDocument();
    });
  });

  it('opens the command dialog on Cmd+K', async () => {
    render(<GlobalSearch />);
    fireEvent.keyDown(document, { key: 'k', metaKey: true });

    await waitFor(() => {
        expect(screen.getByPlaceholderText('Type a command or search...')).toBeInTheDocument();
    });
  });

  it('fetches data when opened', async () => {
     render(<GlobalSearch />);
     fireEvent.click(screen.getByRole('button'));

     await waitFor(() => {
         expect(apiClient.listServices).toHaveBeenCalled();
         expect(apiClient.listTools).toHaveBeenCalled();
         expect(apiClient.listResources).toHaveBeenCalled();
         expect(apiClient.listPrompts).toHaveBeenCalled();
     });
  });

  it('navigates to dashboard when Dashboard is selected', async () => {
    render(<GlobalSearch />);
    fireEvent.click(screen.getByRole('button'));

    await waitFor(() => {
        const dashboardOption = screen.getByText('Dashboard');
        fireEvent.click(dashboardOption);
    });

    expect(mockPush).toHaveBeenCalledWith('/');
  });

   it('navigates to settings when Settings is selected', async () => {
    render(<GlobalSearch />);
    fireEvent.click(screen.getByRole('button'));

    await waitFor(() => {
        const settingsOption = screen.getByText('Settings');
        fireEvent.click(settingsOption);
    });

    expect(mockPush).toHaveBeenCalledWith('/settings');
  });


  it('changes theme when theme option is selected', async () => {
    render(<GlobalSearch />);
    fireEvent.click(screen.getByRole('button'));

    // We might need to type "Dark" to filter if it's not immediately visible or to be sure
    const input = screen.getByPlaceholderText('Type a command or search...');
    fireEvent.change(input, { target: { value: 'Dark' } });

    await waitFor(() => {
        const darkOption = screen.getByText('Dark');
        fireEvent.click(darkOption);
    });

    expect(mockSetTheme).toHaveBeenCalledWith('dark');
  });

  it('displays and navigates to tools', async () => {
      (apiClient.listTools as any).mockResolvedValue({
          tools: [{ name: 'test-tool', description: 'A test tool' }]
      });

      render(<GlobalSearch />);
      fireEvent.click(screen.getByRole('button'));

      await waitFor(() => {
          expect(screen.getByText('test-tool')).toBeInTheDocument();
      });

      fireEvent.click(screen.getByText('test-tool'));
      expect(mockPush).toHaveBeenCalledWith('/tools?name=test-tool');
  });
});
