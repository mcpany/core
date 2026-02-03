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
});
