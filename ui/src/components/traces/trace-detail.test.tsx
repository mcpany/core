/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import React from 'react';
import { render, screen } from '@testing-library/react';
import { TraceDetail } from './trace-detail';
import { Trace, Span } from '@/app/api/traces/route';
import { vi } from 'vitest';

// Mock scrollIntoView
window.HTMLElement.prototype.scrollIntoView = vi.fn();

// Mock dependencies
vi.mock('next/navigation', () => ({
    useRouter: () => ({
        push: vi.fn(),
    }),
}));

describe('TraceDetail', () => {
    it('renders trace details correctly', () => {
        const span: Span = {
            id: '1',
            traceId: 'trace-1',
            parentId: undefined,
            name: 'Root Span',
            startTime: 1000,
            endTime: 1100,
            status: 'success',
            input: { foo: 'bar' },
            output: { result: 'ok' },
            type: 'tool',
            children: []
        };

        const trace: Trace = {
            id: 'trace-1',
            timestamp: 1000,
            totalDuration: 100,
            status: 'success',
            rootSpan: span,
        };

        render(<TraceDetail trace={trace} />);

        // Use getAllByText because it appears in the header and the waterfall
        expect(screen.getAllByText('Root Span').length).toBeGreaterThan(0);
        expect(screen.getAllByText('100ms').length).toBeGreaterThan(0);
        // Check input/output JSON presence (checking for keys)
        expect(screen.getByText(/"foo": "bar"/)).toBeInTheDocument();
        expect(screen.getByText(/"result": "ok"/)).toBeInTheDocument();
    });

    it('renders nested spans in waterfall', () => {
         const childSpan: Span = {
            id: '2',
            traceId: 'trace-1',
            parentId: '1',
            name: 'Child Span',
            startTime: 1010,
            endTime: 1050,
            status: 'success',
            type: 'service',
            children: []
        };

        const rootSpan: Span = {
            id: '1',
            traceId: 'trace-1',
            parentId: undefined,
            name: 'Root Span',
            startTime: 1000,
            endTime: 1100,
            status: 'success',
            type: 'tool',
            children: [childSpan]
        };

        const trace: Trace = {
            id: 'trace-1',
            timestamp: 1000,
            totalDuration: 100,
            status: 'success',
            rootSpan: rootSpan,
        };

        render(<TraceDetail trace={trace} />);

        expect(screen.getAllByText('Root Span').length).toBeGreaterThan(0);
        expect(screen.getAllByText('Child Span').length).toBeGreaterThan(0);
    });
});
