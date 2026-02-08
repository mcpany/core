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

// Mock URL
global.URL.createObjectURL = vi.fn();
global.URL.revokeObjectURL = vi.fn();

// Mock ResizeObserver
window.ResizeObserver = vi.fn().mockImplementation(() => ({
    observe: vi.fn(),
    unobserve: vi.fn(),
    disconnect: vi.fn(),
}));

describe('PlaygroundClient', () => {
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

  it('renders', () => {
      // Basic render test
      render(<PlaygroundClient />);
      expect(screen.getByText('Playground')).toBeInTheDocument();
  });

  it('displays execution duration after tool execution', async () => {
      // Mock executeTool to take some time
      (apiClient.executeTool as any).mockImplementation(async () => {
          await new Promise(resolve => setTimeout(resolve, 50)); // 50ms delay
          return { result: "success" };
      });

      render(<PlaygroundClient />);

      const input = screen.getByPlaceholderText(/e.g. calculator/);
      // Use fireEvent to simulate typing
      fireEvent.change(input, { target: { value: 'test_tool {}' } });
      fireEvent.keyDown(input, { key: 'Enter', code: 'Enter' });

      // Wait for the result
      await waitFor(() => {
          expect(screen.getByText(/Result \(test_tool\)/)).toBeInTheDocument();
      });

      // Check for duration badge (regex for "Xms")
      // It should be around 50ms, but we just check if any "ms" badge exists
      const durationBadge = screen.getByText(/\d+ms/);
      expect(durationBadge).toBeInTheDocument();
      // Optional: check it's visible
      expect(durationBadge).toBeVisible();
  });

  it('exports session history', async () => {
      render(<PlaygroundClient />);
      const exportBtn = screen.getByTitle('Export Session');
      fireEvent.click(exportBtn);
      expect(global.URL.createObjectURL).toHaveBeenCalled();
  });

  it('imports session history', async () => {
      render(<PlaygroundClient />);

      const file = new File([JSON.stringify([
          {
              id: "test-id",
              type: "assistant",
              content: "Imported Message",
              timestamp: new Date().toISOString()
          }
      ])], "session.json", { type: "application/json" });

      const importBtn = screen.getByTitle('Import Session');

      const createElementSpy = vi.spyOn(document, 'createElement');
      fireEvent.click(importBtn);

      const input = createElementSpy.mock.results.find(r => r.value instanceof HTMLInputElement && r.value.type === 'file')?.value as HTMLInputElement;
      expect(input).toBeDefined();

      Object.defineProperty(input, 'files', { value: [file] });
      fireEvent.change(input);

      await waitFor(() => {
          expect(screen.getByText("Imported Message")).toBeInTheDocument();
      });
  });
});
