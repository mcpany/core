/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { render, screen, fireEvent } from '@testing-library/react';
import { PlaygroundClientPro } from './playground-client-pro';
import { vi, describe, it, expect, beforeEach, afterEach } from 'vitest';

// Mock dependencies
vi.mock('@/lib/client', () => ({
  apiClient: {
    listTools: vi.fn().mockResolvedValue({ tools: [] }),
    executeTool: vi.fn(),
  },
}));

vi.mock("next/navigation", () => ({
  useSearchParams: () => new URLSearchParams(),
}));

vi.mock("@/hooks/use-mobile", () => ({
    useIsMobile: () => false
}));

// Mock scrollIntoView
window.HTMLElement.prototype.scrollIntoView = vi.fn();

// Mock ResizeObserver
window.ResizeObserver = vi.fn().mockImplementation(() => ({
    observe: vi.fn(),
    unobserve: vi.fn(),
    disconnect: vi.fn(),
}));

describe('PlaygroundClientPro Export', () => {
  beforeEach(() => {
    // Mock URL methods
    global.URL.createObjectURL = vi.fn().mockReturnValue('blob:test-url');
    global.URL.revokeObjectURL = vi.fn();

    // Mock localStorage
    const store: Record<string, string> = {};
    Object.defineProperty(window, 'localStorage', {
      value: {
        getItem: (key: string) => store[key] || null,
        setItem: (key: string, value: string) => {
          store[key] = value.toString();
        },
        clear: () => {
          for (const key in store) delete store[key];
        },
      },
      writable: true,
    });
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  it('renders export button', () => {
    render(<PlaygroundClientPro />);
    // "Export" button text might be hidden or inside the button.
    // The button has "Export" text and an icon.
    expect(screen.getByText('Export')).toBeInTheDocument();
  });

  it('triggers download on export click', () => {
    render(<PlaygroundClientPro />);

    // Helper to spy on link click
    const clickSpy = vi.fn();
    const originalCreateElement = document.createElement;

    vi.spyOn(document, 'createElement').mockImplementation((tagName) => {
        const el = originalCreateElement.call(document, tagName);
        if (tagName === 'a') {
            el.click = clickSpy;
        }
        return el;
    });

    const exportBtn = screen.getByText('Export');
    fireEvent.click(exportBtn);

    expect(global.URL.createObjectURL).toHaveBeenCalled();
    expect(clickSpy).toHaveBeenCalled();

    // Check if the blob was created with correct type
    const blob = (global.URL.createObjectURL as any).mock.calls[0][0];
    expect(blob.type).toBe('application/json');
  });
});
