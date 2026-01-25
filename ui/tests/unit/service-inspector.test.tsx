/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import React from 'react';
import { render, screen, waitFor } from '@testing-library/react';
import { ServiceInspector } from '../../src/components/services/service-inspector';
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';

// Mock fetch global
global.fetch = vi.fn();

// Mock scrollIntoView
Element.prototype.scrollIntoView = vi.fn();

// Mock useRouter
vi.mock('next/navigation', () => ({
  useRouter: () => ({
    push: vi.fn(),
    replace: vi.fn(),
    prefetch: vi.fn(),
  }),
}));

describe('ServiceInspector', () => {
  const mockTraces = [
    {
      id: 't1',
      timestamp: new Date().toISOString(),
      totalDuration: 100,
      status: 'success',
      trigger: 'user',
      rootSpan: {
        id: 's1',
        name: 'call_my_tool',
        type: 'tool',
        startTime: Date.now(),
        endTime: Date.now() + 100,
        status: 'success',
        input: {
          method: 'tools/call',
          params: {
            name: 'my_tool',
            arguments: {}
          }
        },
        children: []
      }
    },
    {
      id: 't2',
      timestamp: new Date().toISOString(),
      totalDuration: 50,
      status: 'error',
      trigger: 'user',
      rootSpan: {
        id: 's2',
        name: 'call_other_tool',
        type: 'tool',
        startTime: Date.now(),
        endTime: Date.now() + 50,
        status: 'error',
        input: {
          method: 'tools/call',
          params: {
            name: 'other_tool',
            arguments: {}
          }
        },
        children: []
      }
    }
  ];

  beforeEach(() => {
    (global.fetch as any).mockResolvedValue({
      ok: true,
      json: async () => mockTraces
    });
  });

  afterEach(() => {
    vi.resetAllMocks();
  });

  it('renders traces filtered by tool names', async () => {
    render(<ServiceInspector serviceId="svc-1" toolNames={['my_tool']} />);

    await waitFor(() => {
      // TraceList renders buttons with text inside spans
      // Just searching for text should find it
      expect(screen.getByText('call_my_tool')).toBeDefined();
    });

    // t2 should be filtered out because 'other_tool' is not in toolNames
    expect(screen.queryByText('call_other_tool')).toBeNull();
  });
});
