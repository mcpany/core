/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { PlaygroundClient } from './playground-client';
import { vi, describe, it, expect, beforeEach, afterEach } from 'vitest';
import { apiClient } from '@/lib/client';

// Mock dependencies
vi.mock('@/lib/client', () => ({
  apiClient: {
    listTools: vi.fn().mockResolvedValue({ tools: [] }),
    executeTool: vi.fn(),
  },
}));

// Mock scrollIntoView
window.HTMLElement.prototype.scrollIntoView = vi.fn();

// Mock ResizeObserver
window.ResizeObserver = vi.fn().mockImplementation(() => ({
    observe: vi.fn(),
    unobserve: vi.fn(),
    disconnect: vi.fn(),
}));

describe('PlaygroundClient Diff Feature', () => {
  const originalLocation = window.location;

  beforeEach(() => {
    // @ts-expect-error Mocking window.location
    delete window.location;
    // @ts-expect-error Mocking window.location
    window.location = { ...originalLocation, search: '' };
  });

  afterEach(() => {
    window.location = originalLocation;
    vi.clearAllMocks();
  });

  it('displays "Show Changes" button when tool output changes', async () => {
      // Mock executeTool to return different results
      const mockExecute = apiClient.executeTool as any;

      // First run: Result A
      mockExecute.mockResolvedValueOnce({ result: { output: "Result A" }, traceId: "1" });

      // Second run: Result B
      mockExecute.mockResolvedValueOnce({ result: { output: "Result B" }, traceId: "2" });

      render(<PlaygroundClient />);

      const input = screen.getByPlaceholderText(/e.g. calculator/);

      // --- Run 1 ---
      fireEvent.change(input, { target: { value: 'test_tool {}' } });
      fireEvent.keyDown(input, { key: 'Enter', code: 'Enter' });

      // Wait for first result
      await waitFor(() => {
          expect(screen.getByText(/"Result A"/)).toBeInTheDocument();
      });

      // Ensure "Show Changes" is NOT present yet
      expect(screen.queryByText('Show Changes')).not.toBeInTheDocument();

      // --- Run 2 (Same tool, same args) ---
      fireEvent.change(input, { target: { value: 'test_tool {}' } });
      fireEvent.keyDown(input, { key: 'Enter', code: 'Enter' });

      // Wait for second result
      await waitFor(() => {
          expect(screen.getByText(/"Result B"/)).toBeInTheDocument();
      });

      // Now "Show Changes" should appear
      const showChangesBtn = await screen.findByText('Show Changes');
      expect(showChangesBtn).toBeInTheDocument();

      // Click it to see diff
      fireEvent.click(showChangesBtn);

      // Expect Dialog to show both results
      await waitFor(() => {
        expect(screen.getByText('Output Differences')).toBeInTheDocument();
        expect(screen.getByText('Previous Output')).toBeInTheDocument();
        expect(screen.getByText('Current Output')).toBeInTheDocument();
      });
  });

  it('does NOT display "Show Changes" if output is identical', async () => {
      // Mock executeTool to return SAME results
      const mockExecute = apiClient.executeTool as any;
      mockExecute.mockResolvedValue({ result: { output: "Result A" }, traceId: "1" }); // Always returns A

      render(<PlaygroundClient />);
      const input = screen.getByPlaceholderText(/e.g. calculator/);

      // Run 1
      fireEvent.change(input, { target: { value: 'test_tool {}' } });
      fireEvent.keyDown(input, { key: 'Enter', code: 'Enter' });
      await waitFor(() => expect(screen.getByText(/"Result A"/)).toBeInTheDocument());

      // Run 2
      fireEvent.change(input, { target: { value: 'test_tool {}' } });
      fireEvent.keyDown(input, { key: 'Enter', code: 'Enter' });

      // Wait for second result
      await waitFor(() => {
          const results = screen.getAllByText(/"Result A"/);
          expect(results.length).toBeGreaterThanOrEqual(2);
      });

      // Should NOT have "Show Changes"
      expect(screen.queryByText('Show Changes')).not.toBeInTheDocument();
  });
});
