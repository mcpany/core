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
    // Mock URL.createObjectURL
    const mockCreateObjectURL = vi.fn();
    const mockRevokeObjectURL = vi.fn();
    global.URL.createObjectURL = mockCreateObjectURL;
    global.URL.revokeObjectURL = mockRevokeObjectURL;

    // Mock anchor click
    const mockClick = vi.fn();
    const mockAnchor = { href: '', download: '', click: mockClick };

    // Save original
    const originalCreateElement = document.createElement.bind(document);

    const mockCreateElement = vi.spyOn(document, 'createElement').mockImplementation((tagName: string, options) => {
        if (tagName === 'a') {
            return mockAnchor as any;
        }
        return originalCreateElement(tagName, options);
    });

    const originalAppendChild = document.body.appendChild.bind(document.body);
    const mockAppendChild = vi.spyOn(document.body, 'appendChild').mockImplementation((node) => {
        if (node === mockAnchor) return node;
        return originalAppendChild(node);
    });

    const originalRemoveChild = document.body.removeChild.bind(document.body);
    const mockRemoveChild = vi.spyOn(document.body, 'removeChild').mockImplementation((node) => {
        if (node === mockAnchor) return node;
        return originalRemoveChild(node);
    });

    render(<PlaygroundClient />);

    // Trigger export
    const exportBtn = screen.getByTitle('Export Session');
    fireEvent.click(exportBtn);

    expect(mockCreateObjectURL).toHaveBeenCalled();
    expect(mockCreateElement).toHaveBeenCalledWith('a');
    expect(mockAnchor.download).toMatch(/playground-session-.*\.json/);
    expect(mockClick).toHaveBeenCalled();

    // cleanup
    mockCreateElement.mockRestore();
    mockAppendChild.mockRestore();
    mockRemoveChild.mockRestore();
  });

  it('imports session history', async () => {
    render(<PlaygroundClient />);

    const file = new File([JSON.stringify([{
        id: "test-id",
        type: "user",
        content: "Imported Message",
        timestamp: new Date().toISOString()
    }])], "session.json", { type: "application/json" });

    // The hidden input
    // eslint-disable-next-line testing-library/no-node-access
    const fileInput = document.querySelector('input[type="file"]');
    if (!fileInput) throw new Error("File input not found");

    await waitFor(() => {
        fireEvent.change(fileInput, { target: { files: [file] } });
    });

    await waitFor(() => {
        expect(screen.getByText('Imported Message')).toBeInTheDocument();
    });
  });
});
