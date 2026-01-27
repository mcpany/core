/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { expect, test, describe } from 'vitest';
import { render, screen } from '@testing-library/react';
import { SequenceDiagram } from '@/components/traces/sequence-diagram';
import { Trace } from '@/types/trace';

describe('SequenceDiagram', () => {
    const mockTrace: Trace = {
        id: 'trace-123',
        timestamp: new Date().toISOString(),
        totalDuration: 100,
        status: 'success',
        trigger: 'user',
        rootSpan: {
            id: 'span-1',
            name: 'root',
            type: 'core',
            startTime: 1000,
            endTime: 1100,
            status: 'success',
            children: [
                {
                    id: 'span-2',
                    name: 'tool-call-1',
                    type: 'tool',
                    startTime: 1010,
                    endTime: 1090,
                    status: 'success',
                }
            ]
        }
    };

    test('renders actors correctly', () => {
        render(<SequenceDiagram trace={mockTrace} />);

        expect(screen.getByText('User / Client')).toBeDefined();
        expect(screen.getByText('MCP Core')).toBeDefined();
        // The tool actor label matches the span name for type='tool'
        expect(screen.getAllByText('tool-call-1').length).toBeGreaterThan(0);
    });

    test('renders messages correctly', () => {
        render(<SequenceDiagram trace={mockTrace} />);

        // Root message
        expect(screen.getAllByText(/root/)).toBeDefined();
        // Return message from tool (label might be 'return')
        expect(screen.getAllByText(/return/)).toBeDefined();
    });
});
