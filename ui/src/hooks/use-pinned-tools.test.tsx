/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { renderHook, act } from '@testing-library/react';
import { usePinnedTools } from './use-pinned-tools';
import { describe, it, expect, beforeEach, vi } from 'vitest';

const STORAGE_KEY = "mcpany-pinned-tools";

describe('usePinnedTools', () => {
  beforeEach(() => {
    window.localStorage.clear();
    vi.clearAllMocks();
  });

  it('should initialize with empty pinned tools', () => {
    const { result } = renderHook(() => usePinnedTools());
    expect(result.current.pinnedTools).toEqual([]);
  });

  it('should load pinned tools from local storage', () => {
    window.localStorage.setItem(STORAGE_KEY, JSON.stringify(['tool1']));
    const { result } = renderHook(() => usePinnedTools());
    expect(result.current.pinnedTools).toEqual(['tool1']);
  });

  it('should toggle pin', () => {
    const { result } = renderHook(() => usePinnedTools());

    act(() => {
      result.current.togglePin('tool1');
    });

    expect(result.current.pinnedTools).toEqual(['tool1']);
    expect(window.localStorage.getItem(STORAGE_KEY)).toEqual(JSON.stringify(['tool1']));

    act(() => {
      result.current.togglePin('tool1');
    });

    expect(result.current.pinnedTools).toEqual([]);
    expect(window.localStorage.getItem(STORAGE_KEY)).toEqual(JSON.stringify([]));
  });

  it('should check if a tool is pinned', () => {
    const { result } = renderHook(() => usePinnedTools());

    act(() => {
      result.current.togglePin('tool1');
    });

    expect(result.current.isPinned('tool1')).toBe(true);
    expect(result.current.isPinned('tool2')).toBe(false);
  });

  it('should bulk pin tools', () => {
    const { result } = renderHook(() => usePinnedTools());

    act(() => {
      result.current.bulkPin(['tool1', 'tool2']);
    });

    expect(result.current.pinnedTools).toEqual(['tool1', 'tool2']);
    expect(window.localStorage.getItem(STORAGE_KEY)).toEqual(JSON.stringify(['tool1', 'tool2']));

    // Should not duplicate
    act(() => {
      result.current.bulkPin(['tool2', 'tool3']);
    });
    // Set order is not guaranteed, but array implementation preserves order usually
    // We can check length and containment
    expect(result.current.pinnedTools).toHaveLength(3);
    expect(result.current.pinnedTools).toContain('tool1');
    expect(result.current.pinnedTools).toContain('tool2');
    expect(result.current.pinnedTools).toContain('tool3');
  });

  it('should bulk unpin tools', () => {
    window.localStorage.setItem(STORAGE_KEY, JSON.stringify(['tool1', 'tool2', 'tool3']));
    const { result } = renderHook(() => usePinnedTools());

    act(() => {
      result.current.bulkUnpin(['tool1', 'tool3']);
    });

    expect(result.current.pinnedTools).toEqual(['tool2']);
    expect(window.localStorage.getItem(STORAGE_KEY)).toEqual(JSON.stringify(['tool2']));
  });
});
