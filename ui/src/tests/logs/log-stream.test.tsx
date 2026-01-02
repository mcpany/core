/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { render, screen, fireEvent, waitFor, act } from '@testing-library/react';
import { LogStream, LogEntry } from '@/components/logs/log-stream';

// Mock EventSource
class MockEventSource {
  onopen: (() => void) | null = null;
  onmessage: ((event: MessageEvent) => void) | null = null;
  onerror: ((event: Event) => void) | null = null;
  close = vi.fn();
  url: string;

  constructor(url: string) {
    this.url = url;
  }

  // Helper to simulate incoming message
  emitMessage(data: LogEntry) {
    if (this.onmessage) {
      this.onmessage({ data: JSON.stringify(data) } as MessageEvent);
    }
  }

  // Helper to simulate open
  emitOpen() {
    if (this.onopen) {
      this.onopen();
    }
  }
}

// Mock scrollIntoView
window.HTMLElement.prototype.scrollIntoView = vi.fn();

describe('LogStream', () => {
  let mockEventSource: MockEventSource;

  beforeEach(() => {
    // Correctly mock the global EventSource constructor to be a class
    vi.stubGlobal('EventSource', class {
      constructor(url: string) {
        mockEventSource = new MockEventSource(url);
        return mockEventSource;
      }
    });
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it('renders correctly and connects to stream', async () => {
    render(<LogStream />);

    expect(screen.getByText('Live Stream')).toBeDefined();
    // Use regex to match partial text or handle flexible spacing
    expect(screen.getByText(/Disconnected/i)).toBeDefined();

    // Simulate connection
    expect(mockEventSource).toBeDefined();

    // Wrap state update in act
    act(() => {
        mockEventSource.emitOpen();
    });

    await waitFor(() => {
        expect(screen.getByText(/Connected/i)).toBeDefined();
    });
  });

  it('displays logs received from stream', async () => {
    render(<LogStream />);
    act(() => {
        mockEventSource.emitOpen();
    });

    const logEntry: LogEntry = {
      id: '1',
      timestamp: new Date().toISOString(),
      level: 'INFO',
      message: 'Test log message',
      source: 'test-source'
    };

    act(() => {
        mockEventSource.emitMessage(logEntry);
    });

    await waitFor(() => {
      expect(screen.getByText('Test log message')).toBeDefined();
      expect(screen.getByText('test-source')).toBeDefined();
    });
  });

  it('filters logs by search query', async () => {
    render(<LogStream />);
    act(() => {
        mockEventSource.emitOpen();
    });

    const log1: LogEntry = { id: '1', timestamp: new Date().toISOString(), level: 'INFO', message: 'Apple', source: 's1' };
    const log2: LogEntry = { id: '2', timestamp: new Date().toISOString(), level: 'INFO', message: 'Banana', source: 's1' };

    act(() => {
        mockEventSource.emitMessage(log1);
        mockEventSource.emitMessage(log2);
    });

    await waitFor(() => {
      expect(screen.getByText('Apple')).toBeDefined();
      expect(screen.getByText('Banana')).toBeDefined();
    });

    const searchInput = screen.getByPlaceholderText('Filter logs...');

    act(() => {
        fireEvent.change(searchInput, { target: { value: 'Apple' } });
    });

    await waitFor(() => {
        expect(screen.getByText('Apple')).toBeDefined();
        expect(screen.queryByText('Banana')).toBeNull();
    });
  });

  it('clears logs when clear button is clicked', async () => {
    render(<LogStream />);
    act(() => {
        mockEventSource.emitOpen();
    });

    const logEntry: LogEntry = {
      id: '1',
      timestamp: new Date().toISOString(),
      level: 'INFO',
      message: 'To be cleared',
      source: 'test-source'
    };

    act(() => {
        mockEventSource.emitMessage(logEntry);
    });

    await waitFor(() => {
      expect(screen.getByText('To be cleared')).toBeDefined();
    });

    const clearButton = screen.getByText('Clear');

    act(() => {
        fireEvent.click(clearButton);
    });

    await waitFor(() => {
      expect(screen.queryByText('To be cleared')).toBeNull();
    });
  });
});
