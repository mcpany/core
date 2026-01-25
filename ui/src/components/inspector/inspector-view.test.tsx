import { render, screen, act } from '@testing-library/react';
import { InspectorView } from './inspector-view';
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import React from 'react';

// Mock ResizeObserver
global.ResizeObserver = vi.fn().mockImplementation(() => ({
  observe: vi.fn(),
  unobserve: vi.fn(),
  disconnect: vi.fn(),
}));

let mockWsInstance: any;

// Mock WebSocket
function MockWebSocket(url: string) {
    // @ts-ignore
    mockWsInstance = this;
    // @ts-ignore
    this.send = vi.fn();
    // @ts-ignore
    this.close = vi.fn();
    // @ts-ignore
    this.onopen = null;
    // @ts-ignore
    this.onclose = null;
    // @ts-ignore
    this.onmessage = null;
}

global.WebSocket = MockWebSocket as any;

// Mock ScrollArea
vi.mock("@/components/ui/scroll-area", () => ({
  ScrollArea: ({ children }: { children: React.ReactNode }) => <div data-testid="scroll-area">{children}</div>
}));

// Mock Resizable panels (simplified)
vi.mock("@/components/ui/resizable", () => ({
  ResizablePanelGroup: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
  ResizablePanel: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
  ResizableHandle: () => <div>|</div>
}));

describe('InspectorView', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockWsInstance = null;
    global.WebSocket = MockWebSocket as any;
  });

  it('renders correctly', () => {
    render(<InspectorView />);
    expect(screen.getByText('Traffic Inspector')).toBeInTheDocument();
    expect(screen.getByText('0 events')).toBeInTheDocument();
  });

  it('connects to websocket on mount', () => {
    render(<InspectorView />);
    expect(mockWsInstance).toBeTruthy();
  });

  it('displays traffic events', async () => {
    render(<InspectorView />);

    // Simulate WebSocket open
    if (mockWsInstance && mockWsInstance.onopen) {
        await act(async () => {
            (mockWsInstance.onopen as any)();
        });
    }

    // Simulate incoming log message
    const trafficLog = {
      id: '123',
      timestamp: new Date().toISOString(),
      level: 'INFO',
      message: 'Request completed',
      metadata: {
        method: 'tools/call',
        duration: '10ms',
        request_payload: JSON.stringify({ params: { name: 'test' } }),
        response_payload: JSON.stringify({ result: { content: 'ok' } }),
      }
    };

    if (mockWsInstance && mockWsInstance.onmessage) {
        await act(async () => {
            (mockWsInstance.onmessage as any)({ data: JSON.stringify(trafficLog) });
        });
    }

    // Use findAllByText for elements that appear multiple times (List + Detail)
    const methods = await screen.findAllByText('tools/call');
    expect(methods.length).toBeGreaterThan(0);

    const durations = await screen.findAllByText('10ms');
    expect(durations.length).toBeGreaterThan(0);
  });
});
