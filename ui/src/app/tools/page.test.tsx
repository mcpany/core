/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */


import React from 'react';
import { render, screen, waitFor, fireEvent, within } from '@testing-library/react';
import ToolsPage from './page';
import { apiClient } from '@/lib/client';
import { vi, Mock } from 'vitest';

// Mock apiClient
vi.mock('@/lib/client', () => ({
    apiClient: {
        listTools: vi.fn(),
        listServices: vi.fn(),
        setToolStatus: vi.fn(),
        getToolUsage: vi.fn().mockResolvedValue([]),
    },
}));

// Mock usePinnedTools
vi.mock('@/hooks/use-pinned-tools', () => ({
    usePinnedTools: () => ({
        isPinned: () => false,
        togglePin: vi.fn(),
        isLoaded: true,
    }),
}));

// Mock Select component to avoid Radix UI interaction issues in JSDOM
// eslint-disable-next-line @typescript-eslint/no-unused-vars
vi.mock('@/components/ui/select', () => ({
    Select: ({ value, onValueChange, children }: { value: string; onValueChange: (val: string) => void; children: React.ReactNode }) => (
        <div data-testid={`select-mock-${value}`}>
            <select
                value={value}
                onChange={(e) => onValueChange(e.target.value)}
                data-testid="select-native"
            >
                 <option value="all">All Services</option>
                 <option value="service1-id">Service One</option>
                 <option value="service2-id">Service Two</option>
                 <option value="none">No Grouping</option>
                 <option value="service">Group by Service</option>
                 <option value="category">Group by Category</option>
            </select>
            {children}
        </div>
    ),
    SelectContent: ({ children }: { children: React.ReactNode }) => <>{children}</>,
    SelectItem: ({ value, children }: { value: string; children: React.ReactNode }) => <option value={value}>{children}</option>,
    SelectTrigger: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
    SelectValue: () => null,
}));

// ResizeObserver mock (needed for some UI components)
global.ResizeObserver = class ResizeObserver {
    observe() {}
    unobserve() {}
    disconnect() {}
};
Element.prototype.scrollIntoView = vi.fn();
Element.prototype.setPointerCapture = () => {};
Element.prototype.releasePointerCapture = () => {};
Element.prototype.hasPointerCapture = () => false;


describe('ToolsPage', () => {
    const mockTools = [
        { name: 'tool1', description: 'Tool 1', serviceId: 'service1-id', disable: false },
        { name: 'tool2', description: 'Tool 2', serviceId: 'service2-id', disable: false },
        { name: 'tool3', description: 'Tool 3', serviceId: 'service1-id', disable: true },
        { name: 'special-tool', description: 'A very special tool', serviceId: 'service1-id', disable: false },
    ];

    const mockServices = [
        { id: 'service1-id', name: 'Service One' },
        { id: 'service2-id', name: 'Service Two' },
    ];

    beforeEach(() => {
        vi.clearAllMocks();
        (apiClient.listTools as Mock).mockResolvedValue({ tools: mockTools });
        (apiClient.listServices as Mock).mockResolvedValue(mockServices);
    });

    it('renders tools and services', async () => {
        render(<ToolsPage />);

        await waitFor(() => {
            expect(screen.getByText('tool1')).toBeInTheDocument();
            expect(screen.getByText('tool2')).toBeInTheDocument();
            expect(screen.getByText('tool3')).toBeInTheDocument();
            expect(screen.getByText('Est. Context')).toBeInTheDocument();
        });
    });

    it('filters tools by service', async () => {
        render(<ToolsPage />);

        await waitFor(() => {
            expect(screen.getByText('tool1')).toBeInTheDocument();
        });

        expect(apiClient.listServices).toHaveBeenCalled();

        // Select 'Service One' using the mock native select. We need to distinguish between the two selects.
        // The service filter select initializes with "all".
        const select = screen.getAllByTestId('select-native').find(
            (el) => (el as HTMLSelectElement).value === 'all'
        );

        if (!select) throw new Error("Service filter select not found");

        fireEvent.change(select, { target: { value: 'service1-id' } });

        await waitFor(() => {
            // tool1 and tool3 should be visible (Service One)
            const table = screen.getByRole('table');
            expect(within(table).getByText('tool1')).toBeInTheDocument();
            expect(within(table).getByText('tool3')).toBeInTheDocument();
            // tool2 should not be visible (Service Two)
            expect(within(table).queryByText('tool2')).not.toBeInTheDocument();
        });
    });

    it('filters tools by search query', async () => {
        render(<ToolsPage />);

        await waitFor(() => {
            expect(screen.getByText('tool1')).toBeInTheDocument();
            expect(screen.getByText('special-tool')).toBeInTheDocument();
        });

        const searchInput = screen.getByPlaceholderText('Search tools...');
        fireEvent.change(searchInput, { target: { value: 'special' } });

        await waitFor(() => {
            const table = screen.getByRole('table');
            expect(within(table).getByText('special-tool')).toBeInTheDocument();
            expect(within(table).queryByText('tool1')).not.toBeInTheDocument();
            expect(within(table).queryByText('tool2')).not.toBeInTheDocument();
        });
    });

    it('filters tools by description search', async () => {
        render(<ToolsPage />);

        await waitFor(() => {
            expect(screen.getByText('tool1')).toBeInTheDocument();
        });

        const searchInput = screen.getByPlaceholderText('Search tools...');
        fireEvent.change(searchInput, { target: { value: 'very special' } });

        await waitFor(() => {
            const table = screen.getByRole('table');
            expect(within(table).getByText('special-tool')).toBeInTheDocument();
            expect(within(table).queryByText('tool1')).not.toBeInTheDocument();
        });
    });

    it('groups tools by category', async () => {
        const mockToolsWithTags = [
            { name: 'tool1', description: 'Tool 1', serviceId: 'service1-id', disable: false, tags: ['search'] },
            { name: 'tool2', description: 'Tool 2', serviceId: 'service2-id', disable: false, tags: ['database'] },
            { name: 'tool3', description: 'Tool 3', serviceId: 'service1-id', disable: false, tags: ['search'] },
        ];
        (apiClient.listTools as Mock).mockResolvedValue({ tools: mockToolsWithTags });

        render(<ToolsPage />);

        await waitFor(() => {
            expect(screen.getByText('tool1')).toBeInTheDocument();
        });

        // Find the group by select which initializes with "none"
        const selects = screen.getAllByTestId('select-native');
        const groupBySelect = selects.find(
            (el) => (el as HTMLSelectElement).value === 'none'
        );

        if (!groupBySelect) throw new Error("Group By select not found");

        fireEvent.change(groupBySelect, { target: { value: 'category' } });

        await waitFor(() => {
            // Check if group headers exist
            expect(screen.getByText('search')).toBeInTheDocument();
            expect(screen.getByText('database')).toBeInTheDocument();
        });
    });
});
