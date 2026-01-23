/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import React from 'react';
import { render, screen } from '@testing-library/react';
import { TraceList } from '../../components/traces/trace-list';
import { describe, it, expect, vi } from 'vitest';
import { Trace } from '../../app/api/traces/route';

const mockTraces: Trace[] = [
    {
        id: "trace-1",
        rootSpan: {
            id: "span-1",
            name: "GET /api/v1/test",
            type: "tool",
            startTime: Date.now(),
            endTime: Date.now() + 100,
            status: "success",
            children: [],
        },
        timestamp: new Date().toISOString(),
        totalDuration: 100,
        status: "success",
        trigger: "user",
    },
    {
        id: "trace-2",
        rootSpan: {
            id: "span-2",
            name: "POST /api/v1/create",
            type: "tool",
            startTime: Date.now(),
            endTime: Date.now() + 200,
            status: "error",
            children: [],
        },
        timestamp: new Date().toISOString(),
        totalDuration: 200,
        status: "error",
        trigger: "user",
    }
];

// Mock ResizeObserver
global.ResizeObserver = class ResizeObserver {
    observe() {}
    unobserve() {}
    disconnect() {}
};

describe('TraceList', () => {
    it('should render trace list items', () => {
        render(
            <TraceList
                traces={mockTraces}
                selectedId={null}
                onSelect={vi.fn()}
                searchQuery=""
                onSearchChange={vi.fn()}
                isLive={false}
                onToggleLive={vi.fn()}
            />
        );

        expect(screen.getByText('GET /api/v1/test')).toBeInTheDocument();
        expect(screen.getByText('POST /api/v1/create')).toBeInTheDocument();
    });

    it('should filter by search query', () => {
        render(
            <TraceList
                traces={mockTraces}
                selectedId={null}
                onSelect={vi.fn()}
                searchQuery="create"
                onSearchChange={vi.fn()}
                isLive={false}
                onToggleLive={vi.fn()}
            />
        );

        expect(screen.queryByText('GET /api/v1/test')).not.toBeInTheDocument();
        expect(screen.getByText('POST /api/v1/create')).toBeInTheDocument();
    });

    it('should show empty state', () => {
        render(
            <TraceList
                traces={[]}
                selectedId={null}
                onSelect={vi.fn()}
                searchQuery=""
                onSearchChange={vi.fn()}
                isLive={false}
                onToggleLive={vi.fn()}
            />
        );

        expect(screen.getByText('No traces found.')).toBeInTheDocument();
    });
});
