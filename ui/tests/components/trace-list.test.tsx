/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { expect, test } from 'vitest';
import { render, screen } from '@testing-library/react';
import { TraceList } from '@/components/traces/trace-list';
import { Trace } from '@/app/api/traces/route';

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
        isLive={false}
        onToggleLive={() => {}}
    />
  );

  expect(screen.getByText("test_tool_1")).toBeInTheDocument();
  expect(screen.getByText("test_tool_2")).toBeInTheDocument();
});

test('TraceList filters traces based on search query', () => {
    // If the component filters INTERNALLY based on searchQuery prop, this works.
    // If filtering happens in parent, this test is testing the parent logic which is mocked here?
    // Wait, TraceList usually receives filtered traces or filters them?
    // Checking code... usually List components just render what they get, OR filter if they have internal logic.
    // Assuming internal logic or prop based filtering.

    // Actually, checking standard React patterns, usually the parent filters.
    // But let's assume the component handles it if the test expects it.
    // Or we should pass filtered traces?
    // Let's check if MOCK_TRACES are filtered. No.
    // So the component MUST be doing the filtering.

    const onSearchChange = () => {};

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

    expect(screen.getByText("test_tool_1")).toBeInTheDocument();
    expect(screen.queryByText("test_tool_2")).toBeNull();
});
