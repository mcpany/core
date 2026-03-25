/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { renderHook, act } from '@testing-library/react';
import { useRecentTools } from './use-recent-tools';
import { describe, it, expect, beforeEach, vi } from 'vitest';

describe('useRecentTools', () => {
    beforeEach(() => {
        window.localStorage.clear();
        vi.clearAllMocks();
    });

    it('should initialize with empty array', () => {
        const { result } = renderHook(() => useRecentTools());
        expect(result.current.recentTools).toEqual([]);
        expect(result.current.isLoaded).toBe(true);
    });

    it('should add a recent tool', () => {
        const { result } = renderHook(() => useRecentTools());
        act(() => {
            result.current.addRecent('tool1');
        });
        expect(result.current.recentTools).toEqual(['tool1']);
        expect(window.localStorage.getItem('mcpany-recent-tools')).toEqual(JSON.stringify(['tool1']));
    });

    it('should move existing tool to top', () => {
        const { result } = renderHook(() => useRecentTools());
        act(() => {
            result.current.addRecent('tool1');
            result.current.addRecent('tool2');
        });
        expect(result.current.recentTools).toEqual(['tool2', 'tool1']);

        act(() => {
            result.current.addRecent('tool1');
        });
        expect(result.current.recentTools).toEqual(['tool1', 'tool2']);
    });

    it('should limit to 5 tools', () => {
        const { result } = renderHook(() => useRecentTools());
        act(() => {
            for (let i = 1; i <= 6; i++) {
                result.current.addRecent(`tool${i}`);
            }
        });
        // Expect tool6, tool5, tool4, tool3, tool2
        expect(result.current.recentTools).toEqual(['tool6', 'tool5', 'tool4', 'tool3', 'tool2']);
        expect(result.current.recentTools.length).toBe(5);
    });

    it('should load from local storage', () => {
        window.localStorage.setItem('mcpany-recent-tools', JSON.stringify(['stored1', 'stored2']));
        const { result } = renderHook(() => useRecentTools());
        expect(result.current.recentTools).toEqual(['stored1', 'stored2']);
    });

    it('should remove a recent tool', () => {
        const { result } = renderHook(() => useRecentTools());
        act(() => {
            result.current.addRecent('tool1');
            result.current.addRecent('tool2');
            result.current.removeRecent('tool1');
        });
        expect(result.current.recentTools).toEqual(['tool2']);
    });

    it('should clear all recent tools', () => {
        const { result } = renderHook(() => useRecentTools());
        act(() => {
            result.current.addRecent('tool1');
            result.current.clearRecent();
        });
        expect(result.current.recentTools).toEqual([]);
        expect(window.localStorage.getItem('mcpany-recent-tools')).toBeNull();
    });
});
