/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { render, screen, fireEvent } from '@testing-library/react';
import { ToolSidebar } from './tool-sidebar';
import { ExtendedToolDefinition } from '@/lib/client';
import { vi, describe, it, expect } from 'vitest';

const mockTools: ExtendedToolDefinition[] = [
    { name: 'tool1', serviceId: 'serviceA', tags: ['tag1'], description: 'd1' } as any,
    { name: 'tool2', serviceId: 'serviceB', tags: ['tag2'], description: 'd2' } as any,
    { name: 'tool3', serviceId: 'serviceA', tags: ['tag1', 'tag3'], description: 'd3' } as any,
];

describe('ToolSidebar', () => {
    it('renders services and tags', () => {
        render(<ToolSidebar tools={mockTools} onSelectTool={() => {}} />);
        expect(screen.getAllByText('serviceA')[0]).toBeInTheDocument();
        expect(screen.getAllByText('serviceB')[0]).toBeInTheDocument();
        expect(screen.getAllByText('tag1')[0]).toBeInTheDocument();
        expect(screen.getAllByText('tag2')[0]).toBeInTheDocument();
        expect(screen.getAllByText('tag3')[0]).toBeInTheDocument();
    });

    it('filters by service', () => {
        render(<ToolSidebar tools={mockTools} onSelectTool={() => {}} />);
        // Click the service badge (it appears in filter list first)
        const badges = screen.getAllByText('serviceA');
        fireEvent.click(badges[0]);

        expect(screen.getByText('tool1')).toBeInTheDocument();
        expect(screen.getByText('tool3')).toBeInTheDocument();
        expect(screen.queryByText('tool2')).not.toBeInTheDocument();
    });

    it('filters by tag', () => {
        render(<ToolSidebar tools={mockTools} onSelectTool={() => {}} />);
        // Click the tag badge (tags are also rendered in tool list? No, only in filter list for now in my mock?)
        // In my mock, tags are not rendered in the tool card in the component implementation I saw.
        // Wait, ToolSidebar renders:
        // <div ...>{tool.name}</div>
        // <Badge ...>{tool.serviceId}</Badge>
        // It does NOT render tags in the tool card in the code I modified.
        // So getByText('tag1') might find only one if it's not in the card.
        // But let's be safe.

        const badges = screen.getAllByText('tag1');
        fireEvent.click(badges[0]);

        expect(screen.getByText('tool1')).toBeInTheDocument();
        expect(screen.getByText('tool3')).toBeInTheDocument();
        expect(screen.queryByText('tool2')).not.toBeInTheDocument();
    });
});
