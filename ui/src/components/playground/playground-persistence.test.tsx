/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { PlaygroundClient } from './playground-client';
import { vi, describe, it, expect, beforeEach } from 'vitest';
import { apiClient } from '@/lib/client';

// Mock dependencies
vi.mock('@/lib/client', () => ({
  apiClient: {
    listTools: vi.fn().mockResolvedValue({ tools: [] }),
    executeTool: vi.fn().mockResolvedValue({ success: true }),
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

describe('PlaygroundClient Persistence & Replay', () => {
  beforeEach(() => {
    localStorage.clear();
    vi.clearAllMocks();
  });

  it('loads messages from localStorage', async () => {
    const savedMessages = [
        {
            id: 'test-1',
            type: 'user',
            content: 'test command',
            timestamp: new Date().toISOString()
        }
    ];
    localStorage.setItem('playground_messages', JSON.stringify(savedMessages));

    render(<PlaygroundClient />);

    expect(await screen.findByText('test command')).toBeInTheDocument();
  });

  it('persists messages to localStorage when adding a new message', async () => {
    render(<PlaygroundClient />);

    const input = screen.getByPlaceholderText(/e.g. calculator/);
    fireEvent.change(input, { target: { value: 'echo {}' } });
    fireEvent.keyDown(input, { key: 'Enter' });

    await waitFor(() => {
        const saved = localStorage.getItem('playground_messages');
        expect(saved).toBeTruthy();
        expect(saved).toContain('echo {}');
    });
  });

  it('triggers execution when Replay is clicked', async () => {
      // Setup initial state with a tool call
      const savedMessages = [
          {
              id: 'msg-1',
              type: 'tool-call',
              toolName: 'echo',
              toolArgs: { msg: 'hello' },
              timestamp: new Date().toISOString()
          }
      ];
      localStorage.setItem('playground_messages', JSON.stringify(savedMessages));

      render(<PlaygroundClient />);

      // Find the replay button. It has title "Replay this tool call"
      const replayButton = await screen.findByTitle('Replay this tool call');
      fireEvent.click(replayButton);

      // It should trigger executeTool
      await waitFor(() => {
          expect(apiClient.executeTool).toHaveBeenCalledWith({
              name: 'echo',
              arguments: { msg: 'hello' }
          });
      });
  });
});
