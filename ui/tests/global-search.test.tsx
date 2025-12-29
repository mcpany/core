import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { GlobalSearch } from '../src/components/global-search';
import { useRouter } from 'next/navigation';
import { useTheme } from 'next-themes';

// Mock dependencies
jest.mock('next/navigation', () => ({
  useRouter: jest.fn(),
}));

jest.mock('next-themes', () => ({
  useTheme: jest.fn(),
}));

// Mock ResizeObserver for Radix UI
class ResizeObserver {
  observe() {}
  unobserve() {}
  disconnect() {}
}
global.ResizeObserver = ResizeObserver;

// Mock scrollIntoView which is missing in jsdom but used by cmdk
window.HTMLElement.prototype.scrollIntoView = function() {};

describe('GlobalSearch Component', () => {
  const mockPush = jest.fn();
  const mockSetTheme = jest.fn();

  beforeEach(() => {
    (useRouter as jest.Mock).mockReturnValue({ push: mockPush });
    (useTheme as jest.Mock).mockReturnValue({ setTheme: mockSetTheme });
    mockPush.mockClear();
    mockSetTheme.mockClear();
  });

  it('opens when Cmd+K is pressed', async () => {
    render(<GlobalSearch />);

    // Initially not visible
    expect(screen.queryByPlaceholderText('Type a command or search...')).not.toBeInTheDocument();

    fireEvent.keyDown(document, { key: 'k', metaKey: true });

    await waitFor(() => {
      expect(screen.getByPlaceholderText('Type a command or search...')).toBeInTheDocument();
    });
  });

  it('navigates to Dashboard when selected', async () => {
    render(<GlobalSearch />);
    fireEvent.keyDown(document, { key: 'k', metaKey: true });

    await waitFor(() => {
        expect(screen.getByPlaceholderText('Type a command or search...')).toBeInTheDocument();
    });

    // We need to wait for items to appear. cmdk might render them async or virtualized.
    const dashboardItem = await screen.findByText('Dashboard');
    fireEvent.click(dashboardItem);

    expect(mockPush).toHaveBeenCalledWith('/');
  });

  it('changes theme to dark when selected', async () => {
    render(<GlobalSearch />);
    fireEvent.keyDown(document, { key: 'k', metaKey: true });

    await waitFor(() => {
        expect(screen.getByPlaceholderText('Type a command or search...')).toBeInTheDocument();
    });

    const darkThemeItem = await screen.findByText('Dark');
    fireEvent.click(darkThemeItem);

    expect(mockSetTheme).toHaveBeenCalledWith('dark');
  });
});
