/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { GlobalSearch } from '@/components/global-search';
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
    CommandDialog: ({ children, open }: any) => open ? <div data-testid="command-dialog">{children}</div> : null,
    CommandInput: ({ placeholder, value, onValueChange }: any) => (
      <input
        placeholder={placeholder}
        value={value}
        onChange={(e) => onValueChange(e.target.value)}
        data-testid="command-input"
      />
    ),
    CommandList: ({ children }: any) => <div>{children}</div>,
    CommandEmpty: ({ children }: any) => <div>{children}</div>,
    CommandGroup: ({ children, heading }: any) => (
      <div>
        <h3>{heading}</h3>
        {children}
      </div>
    ),
    CommandItem: ({ children, onSelect, value }: any) => (
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
    render(<GlobalSearch />);
    expect(screen.getByText(/Search feature.../i)).toBeInTheDocument();
  });

  it('opens the dialog when button is clicked', async () => {
    render(<GlobalSearch />);
    fireEvent.click(screen.getByText(/Search feature.../i));

    await waitFor(() => {
        expect(screen.getByTestId('command-dialog')).toBeInTheDocument();
    });
  });

  it('opens the dialog on Cmd+K', async () => {
    render(<GlobalSearch />);
    fireEvent.keyDown(document, { key: 'k', metaKey: true });

    await waitFor(() => {
        expect(screen.getByTestId('command-dialog')).toBeInTheDocument();
    });
  });

  it('displays suggestions and items', async () => {
    render(<GlobalSearch />);
    fireEvent.click(screen.getByText(/Search feature.../i));

    await waitFor(() => {
       expect(screen.getAllByText('Dashboard')).toHaveLength(1);
    });

    await waitFor(() => {
       expect(screen.getAllByText('Services')).toHaveLength(2); // One in suggestions, one in header
    });
  });

  it('navigates when an item is selected', async () => {
    render(<GlobalSearch />);
    fireEvent.click(screen.getByText(/Search feature.../i));

    await waitFor(() => {
       expect(screen.getAllByText('Dashboard')).toHaveLength(1);
    });

    fireEvent.click(screen.getByText('Dashboard'));
    expect(pushMock).toHaveBeenCalledWith('/');
  });

  it('navigates to service detail', async () => {
    render(<GlobalSearch />);
    fireEvent.click(screen.getByText(/Search feature.../i));

    await waitFor(() => {
        expect(screen.getByText('Service A')).toBeInTheDocument();
    });

    fireEvent.click(screen.getByText('Service A'));
    expect(pushMock).toHaveBeenCalledWith('/services?id=1');
  });

   it('changes theme', async () => {
    render(<GlobalSearch />);
    fireEvent.click(screen.getByText(/Search feature.../i));

    await waitFor(() => {
        expect(screen.getByText('Light')).toBeInTheDocument();
    });

    fireEvent.click(screen.getByText('Light'));
    expect(setThemeMock).toHaveBeenCalledWith('light');
  });
});
