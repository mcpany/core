/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { expect, test, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import { TraceList } from '@/components/traces/trace-list';
import { Trace } from '@/app/api/traces/route';

// Mock Virtuoso to render items directly
vi.mock('react-virtuoso', () => ({
  Virtuoso: ({ data, itemContent }: any) => {
    return (
      <div>
        {data.map((item: any, index: number) => (
          <div key={item.id || index}>
            {itemContent(index, item)}
          </div>
        ))}
      </div>
    );
  }
}));

// Mock traces
const MOCK_TRACES: Trace[] = [
    {
        id: "tr_1",
        rootSpan: { id: "sp_1", name: "test_tool_1", type: "tool", startTime: 1000, endTime: 1100, status: "success" },
        timestamp: new Date().toISOString(),
        totalDuration: 100,
        status: "success",
        trigger: "user"
    },
    {
        id: "tr_2",
        rootSpan: { id: "sp_2", name: "test_tool_2", type: "tool", startTime: 2000, endTime: 2200, status: "error" },
        timestamp: new Date().toISOString(),
        totalDuration: 200,
        status: "error",
        trigger: "webhook"
    }
];

test('TraceList renders traces', async () => {
  render(
    <TraceList
        traces={MOCK_TRACES}
        selectedId={null}
        onSelect={() => {}}
        searchQuery=""
        onSearchChange={() => {}}
        isLive={false}
        onToggleLive={() => {}}
    />
  );

  // Use findByText for asynchronous rendering (Virtuoso)
  expect(await screen.findByText("test_tool_1")).toBeDefined();
  expect(await screen.findByText("test_tool_2")).toBeDefined();
});

test('TraceList filters traces based on search query', () => {
    const onSearchChange = () => {}; // Mock

    render(
      <TraceList
          traces={MOCK_TRACES}
          selectedId={null}
          onSelect={() => {}}
          searchQuery="test_tool_1"
          onSearchChange={onSearchChange}
          isLive={false}
          onToggleLive={() => {}}
      />
    );

    expect(screen.queryByText("test_tool_1")).toBeDefined();
    expect(screen.queryByText("test_tool_2")).toBeNull();
});
