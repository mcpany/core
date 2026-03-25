/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { renderHook, act } from '@testing-library/react';
import { usePolling } from './use-polling';
import { vi, describe, it, expect, beforeEach, afterEach } from 'vitest';

describe('usePolling', () => {
  beforeEach(() => {
    vi.useFakeTimers();
    // Reset document.hidden mock
    Object.defineProperty(document, 'hidden', {
      configurable: true,
      get: () => false,
    });
  });

  afterEach(() => {
    vi.clearAllTimers();
    vi.useRealTimers();
  });

  it('polls at the specified interval', () => {
    const callback = vi.fn();
    renderHook(() => usePolling(callback, 1000));

    expect(callback).not.toHaveBeenCalled();

    act(() => {
      vi.advanceTimersByTime(1000);
    });

    expect(callback).toHaveBeenCalledTimes(1);

    act(() => {
      vi.advanceTimersByTime(2000);
    });

    expect(callback).toHaveBeenCalledTimes(3);
  });

  it('stops polling when document is hidden', () => {
    const callback = vi.fn();
    renderHook(() => usePolling(callback, 1000));

    act(() => {
      vi.advanceTimersByTime(1000);
    });
    expect(callback).toHaveBeenCalledTimes(1);

    // Simulate hiding the document
    Object.defineProperty(document, 'hidden', {
      configurable: true,
      get: () => true,
    });
    const event = new Event('visibilitychange');
    document.dispatchEvent(event);

    act(() => {
      vi.advanceTimersByTime(5000);
    });

    // Should not have been called again
    expect(callback).toHaveBeenCalledTimes(1);
  });

  it('resumes polling when document becomes visible', () => {
    const callback = vi.fn();
    renderHook(() => usePolling(callback, 1000));

    // Initially visible, 1 call after 1s
    act(() => {
      vi.advanceTimersByTime(1000);
    });
    expect(callback).toHaveBeenCalledTimes(1);

    // Hide
    Object.defineProperty(document, 'hidden', {
      configurable: true,
      get: () => true,
    });
    document.dispatchEvent(new Event('visibilitychange'));

    act(() => {
      vi.advanceTimersByTime(5000);
    });
    expect(callback).toHaveBeenCalledTimes(1);

    // Show again
    Object.defineProperty(document, 'hidden', {
      configurable: true,
      get: () => false,
    });
    document.dispatchEvent(new Event('visibilitychange'));

    // Should call immediately on resume
    expect(callback).toHaveBeenCalledTimes(2); // +1 immediate

    // And continue polling
    act(() => {
      vi.advanceTimersByTime(1000);
    });
    expect(callback).toHaveBeenCalledTimes(3);
  });

  it('does not poll if delay is null', () => {
    const callback = vi.fn();
    renderHook(() => usePolling(callback, null));

    act(() => {
      vi.advanceTimersByTime(5000);
    });

    expect(callback).not.toHaveBeenCalled();
  });
});
