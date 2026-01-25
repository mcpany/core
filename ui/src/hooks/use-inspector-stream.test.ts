/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { renderHook, act } from "@testing-library/react";
import { useInspectorStream } from "./use-inspector-stream";
import { vi, describe, it, expect, beforeAll, afterAll } from "vitest";

// Mock WebSocket
class MockWebSocket {
  onopen: (() => void) | null = null;
  onclose: (() => void) | null = null;
  onmessage: ((event: { data: string }) => void) | null = null;
  onerror: ((event: unknown) => void) | null = null;
  close = vi.fn();

  constructor(url: string) {}
}

global.WebSocket = MockWebSocket as any;

describe("useInspectorStream", () => {
  let originalWebSocket: any;

  beforeAll(() => {
    originalWebSocket = global.WebSocket;
    global.WebSocket = MockWebSocket as any;
  });

  afterAll(() => {
    global.WebSocket = originalWebSocket;
  });

  it("should initialize with default state", () => {
    const { result } = renderHook(() => useInspectorStream());
    expect(result.current.messages).toEqual([]);
    expect(result.current.isConnected).toBe(false);
    expect(result.current.isPaused).toBe(false);
    expect(result.current.isSimulating).toBe(false);
  });

  it("should start and stop simulation", () => {
    vi.useFakeTimers();
    const { result } = renderHook(() => useInspectorStream());

    act(() => {
      result.current.startSimulation();
    });

    expect(result.current.isSimulating).toBe(true);

    act(() => {
      vi.advanceTimersByTime(1100);
    });

    expect(result.current.messages.length).toBeGreaterThan(0);

    act(() => {
      result.current.stopSimulation();
    });

    expect(result.current.isSimulating).toBe(false);
    vi.useRealTimers();
  });
});
