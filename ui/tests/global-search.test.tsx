import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { GlobalSearch } from '../src/components/global-search';
import { useRouter } from 'next/navigation';
import { useTheme } from 'next-themes';

// Mock lucide-react to avoid ESM issues
jest.mock('lucide-react', () => ({
  Search: () => <svg data-testid="icon-search" />,
  Calculator: () => <svg />,
  Calendar: () => <svg />,
  CreditCard: () => <svg />,
  Settings: () => <svg />,
  Smile: () => <svg />,
  User: () => <svg />,
  LayoutDashboard: () => <svg />,
  ScrollText: () => <svg />,
  Terminal: () => <svg />,
  Blocks: () => <svg />,
  Sun: () => <svg />,
  Moon: () => <svg />,
  Laptop: () => <svg />,
  X: () => <svg />, // Added X because it's used in Dialog
}));

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

// Mock scrollIntoView
window.HTMLElement.prototype.scrollIntoView = jest.fn();

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
    // Radix Dialog renders nothing when closed by default (unless forceMount is used, but typically it portals)
    // However, for unit testing without full DOM env sometimes it's tricky.
    // Let's rely on firing the event and checking if the input appears.

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

    const dashboardItem = screen.getByText('Dashboard');
    fireEvent.click(dashboardItem);

    expect(mockPush).toHaveBeenCalledWith('/');
  });

  it('changes theme to dark when selected', async () => {
    render(<GlobalSearch />);
    fireEvent.keyDown(document, { key: 'k', metaKey: true });

    await waitFor(() => {
        expect(screen.getByPlaceholderText('Type a command or search...')).toBeInTheDocument();
    });

    const darkThemeItem = screen.getByText('Dark');
    fireEvent.click(darkThemeItem);

    expect(mockSetTheme).toHaveBeenCalledWith('dark');
  });
});
