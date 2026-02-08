/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { expect, test, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import { TraceList } from '@/components/traces/trace-list';
import { Trace } from '@/types/trace';

// Mock react-virtuoso for tests as it doesn't render well in jsdom without layout
vi.mock('react-virtuoso', () => ({
  Virtuoso: ({ data, itemContent }: any) => {
    return (
      <div data-testid="virtual-list">
        {data.map((item: any, index: number) => itemContent(index, item))}
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

test('TraceList renders traces', () => {
  render(
    <TraceList
        traces={MOCK_TRACES}
        selectedId={null}
        onSelect={() => {}}
        searchQuery=""
        onSearchChange={() => {}}
        isPaused={false}
        onTogglePaused={() => {}}
        onRefresh={() => {}}
        isConnected={true}
    />
  );

  expect(screen.getByText("test_tool_1")).toBeDefined();
  expect(screen.getByText("test_tool_2")).toBeDefined();
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
          isPaused={false}
          onTogglePaused={() => {}}
          onRefresh={() => {}}
          isConnected={true}
      />
    );

    expect(screen.queryByText("test_tool_1")).toBeDefined();
    expect(screen.queryByText("test_tool_2")).toBeNull();
});
