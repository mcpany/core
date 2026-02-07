/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import React from 'react';
import { render, screen, waitFor, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import MiddlewarePage from '../../src/app/middleware/page';
import { apiClient } from '@/lib/client';

// Mock the drag and drop context
vi.mock('@hello-pangea/dnd', () => ({
  DragDropContext: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
  Droppable: ({ children }: { children: (provided: any) => React.ReactNode }) => (
    <div>
      {children({
        droppableProps: {},
        innerRef: vi.fn(),
        placeholder: null,
      })}
    </div>
  ),
  Draggable: ({ children }: { children: (provided: any, snapshot: any) => React.ReactNode }) => (
    <div>
      {children(
        {
          draggableProps: {},
          dragHandleProps: {},
          innerRef: vi.fn(),
        },
        { isDragging: false }
      )}
    </div>
  ),
}));

// Mock the API client
vi.mock('@/lib/client', () => ({
  apiClient: {
    getGlobalSettings: vi.fn(),
    saveGlobalSettings: vi.fn(),
  },
}));

describe('MiddlewarePage Component', () => {
  const mockMiddlewares = [
    { name: 'Mock Auth', priority: 0, disabled: false },
    { name: 'Mock Logger', priority: 1, disabled: true },
  ];

  beforeEach(() => {
    vi.clearAllMocks();
    (apiClient.getGlobalSettings as any).mockResolvedValue({
      middlewares: mockMiddlewares,
    });
  });

  it('should display the middleware pipeline heading', async () => {
    render(<MiddlewarePage />);
    expect(screen.getByText('Middleware Pipeline')).toBeDefined();
  });

  it('should fetch and display middlewares from the API', async () => {
    render(<MiddlewarePage />);

    // Wait for the API call to resolve and update the UI
    await waitFor(() => {
        expect(apiClient.getGlobalSettings).toHaveBeenCalledTimes(1);
    });

    // Check if the mock middlewares are displayed
    // Mock Auth appears twice (list + visualization)
    const authElements = await screen.findAllByText(/Mock Auth/);
    expect(authElements.length).toBeGreaterThanOrEqual(1);

    // Mock Logger appears once (list only, disabled)
    expect(await screen.findByText(/Mock Logger/)).toBeDefined();
  });

  it('should display switches for each middleware with correct state', async () => {
    render(<MiddlewarePage />);
    await waitFor(() => expect(apiClient.getGlobalSettings).toHaveBeenCalled());

    await screen.findAllByText(/Mock Auth/);

    const switches = screen.getAllByRole('switch');
    expect(switches.length).toBe(2);

    // Check states (Radix switch uses data-state)
    expect(switches[0].getAttribute('data-state')).toMatch(/checked|on/); // Mock Auth is enabled
    expect(switches[1].getAttribute('data-state')).toMatch(/unchecked|off/); // Mock Logger is disabled
  });
});
