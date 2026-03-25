/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { render, screen } from '../../../tests/test-utils';
import userEvent from '@testing-library/user-event';
import { ServiceList } from '@/components/services/service-list';
import { UpstreamServiceConfig } from '@/lib/client';
import { ServiceHealthProvider } from "@/contexts/service-health-context";
import { TooltipProvider } from "@/components/ui/tooltip";
import { vi } from 'vitest';

// Mock ResizeObserver
global.ResizeObserver = class ResizeObserver {
  observe() {}
  unobserve() {}
  disconnect() {}
};

// Mock PointerEvent which is often needed for Radix primitives
class MockPointerEvent extends Event {
  button: number;
  ctrlKey: boolean;
  pointerType: string;

  constructor(type: string, props: PointerEventInit) {
    super(type, props);
    this.button = props.button || 0;
    this.ctrlKey = props.ctrlKey || false;
    this.pointerType = props.pointerType || 'mouse';
  }
}
window.PointerEvent = MockPointerEvent as unknown as typeof PointerEvent;
window.HTMLElement.prototype.scrollIntoView = vi.fn();
window.HTMLElement.prototype.releasePointerCapture = vi.fn();
window.HTMLElement.prototype.hasPointerCapture = vi.fn();

describe('ServiceList', () => {
    const mockService: UpstreamServiceConfig = {
        id: 'test-id',
        name: 'test-service',
        version: '1.0.0',
        disable: false,
        httpService: { address: 'http://localhost' },
        priority: 0,
        loadBalancingStrategy: 0,
    } as unknown as UpstreamServiceConfig;

    it('renders dropdown menu with duplicate and export actions', async () => {
        const user = userEvent.setup();
        const onEdit = vi.fn();
        const onDuplicate = vi.fn();
        const onExport = vi.fn();
        const onDelete = vi.fn();

        render(
            <TooltipProvider>
                <ServiceHealthProvider>
                    <ServiceList
                        services={[mockService]}
                        onEdit={onEdit}
                        onDuplicate={onDuplicate}
                        onExport={onExport}
                        onDelete={onDelete}
                    />
                </ServiceHealthProvider>
            </TooltipProvider>
        );

        // Find the "MoreHorizontal" button (trigger)
        const trigger = screen.getByRole('button', { name: /open menu/i });
        await user.click(trigger);

        // Check if items are visible (awaiting because of portal/animation)
        expect(await screen.findByText('Edit')).toBeInTheDocument();
        expect(screen.getByText('Duplicate')).toBeInTheDocument();
        expect(screen.getByText('Export')).toBeInTheDocument();
        expect(screen.getByText('Delete')).toBeInTheDocument();

        // Click Duplicate
        await user.click(screen.getByText('Duplicate'));
        expect(onDuplicate).toHaveBeenCalledWith(mockService);
    });

    it('renders without actions if not provided', async () => {
        const user = userEvent.setup();
        render(
            <TooltipProvider>
                <ServiceHealthProvider>
                    <ServiceList
                        services={[mockService]}
                    />
                </ServiceHealthProvider>
            </TooltipProvider>
        );

        expect(screen.getByText('test-service')).toBeInTheDocument();

        const trigger = screen.getByRole('button', { name: /open menu/i });
        await user.click(trigger);

        // Menu should open but items should not be there
        // Note: DropdownMenuContent usually renders empty if no children?
        // Or in our code we wrap items in checks.

        // Wait for potential animation/render
        await new Promise(r => setTimeout(r, 100));

        expect(screen.queryByText('Edit')).not.toBeInTheDocument();
        expect(screen.queryByText('Duplicate')).not.toBeInTheDocument();
        expect(screen.queryByText('Export')).not.toBeInTheDocument();
        expect(screen.queryByText('Delete')).not.toBeInTheDocument();
    });

    it('renders bulk actions when services are selected', async () => {
        const user = userEvent.setup();
        const onBulkToggle = vi.fn();
        const onBulkDelete = vi.fn();

        render(
            <TooltipProvider>
                <ServiceHealthProvider>
                    <ServiceList
                        services={[mockService, { ...mockService, id: 'test-2', name: 'service-2' }]}
                        onBulkToggle={onBulkToggle}
                        onBulkDelete={onBulkDelete}
                    />
                </ServiceHealthProvider>
            </TooltipProvider>
        );

        // Check select all
        const selectAllCheckbox = screen.getByLabelText('Select all');
        await user.click(selectAllCheckbox);

        // Bulk actions should be visible
        expect(screen.getByText('2 selected')).toBeInTheDocument();
        expect(screen.getByText('Enable')).toBeInTheDocument();
        expect(screen.getByText('Disable')).toBeInTheDocument();
        expect(screen.getByText('Delete')).toBeInTheDocument();

        // Click Disable
        await user.click(screen.getByText('Disable'));
        expect(onBulkToggle).toHaveBeenCalledWith(['test-service', 'service-2'], false);

        // Click Delete
        await user.click(selectAllCheckbox); // Reselect all
        await user.click(screen.getByText('Delete'));
        expect(onBulkDelete).toHaveBeenCalledWith(['test-service', 'service-2']);
    });
});
