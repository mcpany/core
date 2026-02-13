/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { PlaygroundClientPro } from './playground-client-pro';
import { vi, describe, it, expect, beforeEach } from 'vitest';
import { apiClient } from '@/lib/client';

// Hoist mocks to be accessible inside vi.mock
const { mockedSearchParams } = vi.hoisted(() => {
    return { mockedSearchParams: vi.fn() };
});

// Mock dependencies
vi.mock('@/lib/client', () => ({
  apiClient: {
    listTools: vi.fn().mockResolvedValue({ tools: [] }),
    executeTool: vi.fn(),
  },
}));

vi.mock('@/hooks/use-toast', () => ({
    useToast: () => ({
        toast: vi.fn(),
    }),
}));

vi.mock('@/hooks/use-local-storage', async () => {
    const React = await import('react');
    return {
        useLocalStorage: (key: string, initialValue: any) => {
            const [state, setState] = React.useState(initialValue);
            return [state, setState, true];
        }
    };
});

vi.mock('@/hooks/use-mobile', () => ({
    useIsMobile: () => false,
}));

vi.mock('next/navigation', () => ({
    useSearchParams: mockedSearchParams,
}));

// Mock scrollIntoView
window.HTMLElement.prototype.scrollIntoView = vi.fn();

// Mock ResizeObserver
window.ResizeObserver = vi.fn().mockImplementation(() => ({
    observe: vi.fn(),
    unobserve: vi.fn(),
    disconnect: vi.fn(),
}));

describe('PlaygroundClientPro', () => {
  beforeEach(() => {
      vi.clearAllMocks();
      mockedSearchParams.mockReturnValue(new URLSearchParams());
  });

  it('renders correctly', () => {
      render(<PlaygroundClientPro />);
      expect(screen.getByText('Console')).toBeInTheDocument();
  });

  it('imports history when file is selected', async () => {
      const { container } = render(<PlaygroundClientPro />);

      // Create a mock file
      const messages = [
          {
              id: "test-1",
              type: "user",
              content: "imported command",
              timestamp: new Date().toISOString()
          },
          {
              id: "test-2",
              type: "assistant",
              content: "imported response",
              timestamp: new Date().toISOString()
          }
      ];
      const file = new File([JSON.stringify(messages)], "history.json", { type: "application/json" });

      // Find the hidden input
      const input = container.querySelector('input[type="file"]') as HTMLInputElement;
      expect(input).toBeInTheDocument();

      // Simulate file upload
      await waitFor(() => {
          fireEvent.change(input, { target: { files: [file] } });
      });

      // Assert that messages are loaded
      await waitFor(() => {
          expect(screen.getByText('imported command')).toBeInTheDocument();
          expect(screen.getByText('imported response')).toBeInTheDocument();
      });
  });

  it('opens tool form with pre-filled arguments from URL params', async () => {
      // Setup mock tool
      const mockTool = {
          name: 'weather',
          description: 'Get weather',
          inputSchema: {
              type: 'object',
              properties: {
                  city: { type: 'string' }
              }
          }
      };

      // Mock listTools to return our tool
      (apiClient.listTools as any).mockResolvedValue({ tools: [mockTool] });

      // Setup URL params
      const params = new URLSearchParams();
      params.set('tool', 'weather');
      params.set('args', JSON.stringify({ city: 'London' }));
      mockedSearchParams.mockReturnValue(params);

      render(<PlaygroundClientPro />);

      // Wait for tools to load and dialog to open
      await waitFor(() => {
          // Check if the input field is pre-filled with "London"
          // This confirms the dialog opened and form initialized correctly
          expect(screen.getByDisplayValue('London')).toBeInTheDocument();
      });
  });
});
