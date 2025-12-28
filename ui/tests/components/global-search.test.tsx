import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import '@testing-library/jest-dom';
import { GlobalSearch } from '@/components/global-search';

// Mock next/navigation
jest.mock('next/navigation', () => ({
  useRouter: () => ({
    push: jest.fn(),
  }),
}));

// Mock next-themes
jest.mock('next-themes', () => ({
  useTheme: () => ({
    setTheme: jest.fn(),
  }),
}));

// Mock lucide-react to avoid issues with SVG rendering in tests
jest.mock('lucide-react', () => ({
  Search: () => <div data-testid="icon-search" />,
  LayoutDashboard: () => <div data-testid="icon-dashboard" />,
  Briefcase: () => <div data-testid="icon-briefcase" />,
  Terminal: () => <div data-testid="icon-terminal" />,
  Play: () => <div data-testid="icon-play" />,
  Settings: () => <div data-testid="icon-settings" />,
  Sun: () => <div data-testid="icon-sun" />,
  Moon: () => <div data-testid="icon-moon" />,
  Laptop: () => <div data-testid="icon-laptop" />,
  Wrench: () => <div />,
  FileText: () => <div />,
  MessageSquare: () => <div />,
  Workflow: () => <div />,
  Webhook: () => <div />,
  X: () => <div data-testid="icon-x" />,
}));

// Mock resize observer
global.ResizeObserver = class ResizeObserver {
  observe() {}
  unobserve() {}
  disconnect() {}
};

// Mock scrollIntoView
window.HTMLElement.prototype.scrollIntoView = jest.fn();

// Mock cmdk since it relies on DOM APIs not fully present in JSDOM
// However, typically we want to test the interaction.
// Since cmdk is a complex component, full unit testing might be flaky in JSDOM without
// extensive mocking. But let's try to test the opening logic which is controlled by our component.

describe('GlobalSearch Component', () => {
  it('is initially hidden', () => {
    render(<GlobalSearch />);
    const dialog = screen.queryByRole('dialog');
    expect(dialog).not.toBeInTheDocument();
  });

  it('opens when Cmd+K is pressed', () => {
    render(<GlobalSearch />);

    fireEvent.keyDown(document, { key: 'k', metaKey: true });

    const dialog = screen.getByRole('dialog');
    expect(dialog).toBeInTheDocument();
  });

  it('opens when Ctrl+K is pressed', () => {
    render(<GlobalSearch />);

    fireEvent.keyDown(document, { key: 'k', ctrlKey: true });

    const dialog = screen.getByRole('dialog');
    expect(dialog).toBeInTheDocument();
  });

  it('displays navigation options when open', () => {
    render(<GlobalSearch />);
    fireEvent.keyDown(document, { key: 'k', metaKey: true });

    expect(screen.getByText('Dashboard')).toBeInTheDocument();
    expect(screen.getByText('Services')).toBeInTheDocument();
    expect(screen.getByText('Logs')).toBeInTheDocument();
  });
});
