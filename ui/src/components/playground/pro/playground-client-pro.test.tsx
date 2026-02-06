/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { PlaygroundClientPro } from './playground-client-pro';
import { vi, describe, it, expect, beforeEach, afterEach } from 'vitest';

// Mock dependencies
vi.mock('@/lib/client', () => ({
  apiClient: {
    listTools: vi.fn().mockResolvedValue({ tools: [] }),
    listPrompts: vi.fn().mockResolvedValue({ prompts: [] }),
    executeTool: vi.fn(),
    executePrompt: vi.fn(),
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
    useSearchParams: () => new URLSearchParams(),
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

      // Find the Import button to verify it exists
      const importButton = screen.getByText('Import');
      expect(importButton).toBeInTheDocument();

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
});
