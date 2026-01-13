/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { GlobalSearch } from '@/components/global-search';
import { ShortcutsProvider } from '@/components/shortcuts/shortcuts-provider';
import { vi } from 'vitest';

// Mock the useRouter hook
const pushMock = vi.fn();
vi.mock('next/navigation', () => ({
  useRouter: () => ({
    push: pushMock,
  }),
}));

// Mock the useTheme hook
const setThemeMock = vi.fn();
vi.mock('next-themes', () => ({
  useTheme: () => ({
    setTheme: setThemeMock,
  }),
}));

// Mock the apiClient
vi.mock('@/lib/client', () => ({
  apiClient: {
    listServices: vi.fn().mockResolvedValue({
      services: [{ id: '1', name: 'Service A', version: '1.0' }]
    }),
    listTools: vi.fn().mockResolvedValue({
      tools: [{ name: 'Tool A', description: 'Description A' }]
    }),
    listResources: vi.fn().mockResolvedValue({
      resources: [{ uri: 'resource:a', name: 'Resource A' }]
    }),
    listPrompts: vi.fn().mockResolvedValue({
      prompts: [{ name: 'Prompt A' }]
    })
  }
}));

// Mock the Command components
vi.mock('@/components/ui/command', () => {
  return {
    CommandDialog: ({ children, open }: { children: React.ReactNode; open: boolean }) => open ? <div data-testid="command-dialog">{children}</div> : null,
    CommandInput: ({ placeholder, value, onValueChange }: { placeholder?: string; value?: string; onValueChange: (v: string) => void }) => (
      <input
        placeholder={placeholder}
        value={value}
        onChange={(e) => onValueChange(e.target.value)}
        data-testid="command-input"
      />
    ),
    CommandList: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
    CommandEmpty: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
    CommandGroup: ({ children, heading }: { children: React.ReactNode; heading?: string }) => (
      <div>
        <h3>{heading}</h3>
        {children}
      </div>
    ),
    CommandItem: ({ children, onSelect, value }: { children: React.ReactNode; onSelect?: () => void; value?: string }) => (
      <div onClick={onSelect} data-value={value} data-testid="command-item">
        {children}
      </div>
    ),
    CommandSeparator: () => <hr />,
  }
});

describe('GlobalSearch', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders the search button', () => {
    render(
      <ShortcutsProvider>
        <GlobalSearch />
      </ShortcutsProvider>
    );
    expect(screen.getByText(/Search or type >/i)).toBeInTheDocument();
  });

  it('opens the dialog when button is clicked', async () => {
    render(
      <ShortcutsProvider>
        <GlobalSearch />
      </ShortcutsProvider>
    );
    fireEvent.click(screen.getByText(/Search or type >/i));

    await waitFor(() => {
        expect(screen.getByTestId('command-dialog')).toBeInTheDocument();
    });
  });

  it('opens the dialog on Cmd+K', async () => {
    // Mock navigator.platform to be Mac so Cmd+K works as expected with metaKey
    Object.defineProperty(navigator, 'platform', {
        value: 'MacIntel',
        configurable: true
    });

    render(
      <ShortcutsProvider>
        <GlobalSearch />
      </ShortcutsProvider>
    );
    // Simulate keyboard event with Meta key (Command)
    // The implementation checks for e.key and modifiers.
    // Ensure window event listener is triggered.
    const event = new KeyboardEvent('keydown', {
        key: 'k',
        metaKey: true,
        bubbles: true,
        cancelable: true
    });
    window.dispatchEvent(event);

    await waitFor(() => {
        expect(screen.getByTestId('command-dialog')).toBeInTheDocument();
    });
  });

  it('displays suggestions and items', async () => {
    render(
      <ShortcutsProvider>
        <GlobalSearch />
      </ShortcutsProvider>
    );
    fireEvent.click(screen.getByText(/Search or type >/i));

    await waitFor(() => {
       expect(screen.getAllByText('Dashboard')).toHaveLength(1);
    });

    await waitFor(() => {
       expect(screen.getAllByText('Services')).toHaveLength(2); // One in suggestions, one in header
    });
  });

  it('navigates when an item is selected', async () => {
    render(
      <ShortcutsProvider>
        <GlobalSearch />
      </ShortcutsProvider>
    );
    fireEvent.click(screen.getByText(/Search or type >/i));

    await waitFor(() => {
       expect(screen.getAllByText('Dashboard')).toHaveLength(1);
    });

    fireEvent.click(screen.getByText('Dashboard'));
    expect(pushMock).toHaveBeenCalledWith('/');
  });

  it('navigates to service detail', async () => {
    render(
      <ShortcutsProvider>
        <GlobalSearch />
      </ShortcutsProvider>
    );
    fireEvent.click(screen.getByText(/Search or type >/i));

    await waitFor(() => {
        expect(screen.getByText('Service A')).toBeInTheDocument();
    });

    fireEvent.click(screen.getByText('Service A'));
    expect(pushMock).toHaveBeenCalledWith('/services?id=1');
  });

   it('changes theme', async () => {
    render(
      <ShortcutsProvider>
        <GlobalSearch />
      </ShortcutsProvider>
    );
    fireEvent.click(screen.getByText(/Search or type >/i));

    await waitFor(() => {
        expect(screen.getByText('Light')).toBeInTheDocument();
    });

    fireEvent.click(screen.getByText('Light'));
    expect(setThemeMock).toHaveBeenCalledWith('light');
  });
});
