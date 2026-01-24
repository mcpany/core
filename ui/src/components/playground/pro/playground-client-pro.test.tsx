/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { PlaygroundClientPro } from './playground-client-pro';
import { vi, describe, it, expect, beforeEach, afterEach } from 'vitest';
import { apiClient } from '@/lib/client';

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

  it('detects previous results for diffing', async () => {
      // @ts-ignore
      apiClient.listTools.mockResolvedValue({
          tools: [{ name: 'test.tool', description: 'Test Tool', inputSchema: {} }]
      });

      // Mock first execution
      // @ts-ignore
      apiClient.executeTool.mockResolvedValueOnce({ result: 'output1' });

      render(<PlaygroundClientPro />);

      // Wait for tools to load (implicit by waiting for input interactions)
      const input = screen.getByPlaceholderText('Enter command or select a tool...');

      // Run tool first time
      fireEvent.change(input, { target: { value: 'test.tool {"arg": 1}' } });
      fireEvent.keyDown(input, { key: 'Enter', code: 'Enter' });

      await waitFor(() => {
          expect(apiClient.executeTool).toHaveBeenCalledWith({ name: 'test.tool', arguments: { arg: 1 } }, false);
          expect(screen.getAllByText('Result: test.tool')[0]).toBeInTheDocument();
      });

      // Mock second execution with different result
      // @ts-ignore
      apiClient.executeTool.mockResolvedValueOnce({ result: 'output2' });

      // Run tool second time
      fireEvent.change(input, { target: { value: 'test.tool {"arg": 1}' } });
      fireEvent.keyDown(input, { key: 'Enter', code: 'Enter' });

      await waitFor(() => {
          expect(apiClient.executeTool).toHaveBeenCalledTimes(2);
          // Check if "Show Diff" button appears (logic added in ChatMessage)
          // Since ChatMessage renders based on previousResult, checking for "Show Diff" verifies previousResult was passed correctly
          expect(screen.getByText('Show Diff')).toBeInTheDocument();
      });
  });
});
