
import React from 'react';
import { render, screen, waitFor, fireEvent } from '@testing-library/react';
import ToolsPage from './page';
import { apiClient } from '@/lib/client';
import { vi } from 'vitest';

// Mock apiClient
vi.mock('@/lib/client', () => ({
    apiClient: {
        listTools: vi.fn(),
        listServices: vi.fn(),
        setToolStatus: vi.fn(),
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
    Select: ({ value, onValueChange, _children }: any) => (
        <div data-testid="select-mock">
            <select
                value={value}
                onChange={(e) => onValueChange(e.target.value)}
                data-testid="select-native"
            >
                {/* We need to render children to find SelectItem, but SelectItem structure is complex in Radix.
                    We will just render options manually based on children if we could, but children are ReactNodes.
                    Instead, we can just expose a way to trigger change.
                 */}
                 <option value="all">All Services</option>
                 <option value="service1-id">Service One</option>
                 <option value="service2-id">Service Two</option>
            </select>
        </div>
    ),
    SelectContent: ({ children }: any) => <>{children}</>,
    SelectItem: ({ value, children }: any) => <option value={value}>{children}</option>,
    SelectTrigger: ({ children }: any) => <div>{children}</div>,
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
    ];

    const mockServices = [
        { id: 'service1-id', name: 'Service One' },
        { id: 'service2-id', name: 'Service Two' },
    ];

    beforeEach(() => {
        vi.clearAllMocks();
        (apiClient.listTools as any).mockResolvedValue({ tools: mockTools });
        (apiClient.listServices as any).mockResolvedValue(mockServices);
    });

    it('renders tools and services', async () => {
        render(<ToolsPage />);

        await waitFor(() => {
            expect(screen.getByText('tool1')).toBeInTheDocument();
            expect(screen.getByText('tool2')).toBeInTheDocument();
            expect(screen.getByText('tool3')).toBeInTheDocument();
        });
    });

    it('filters tools by service', async () => {
        render(<ToolsPage />);

        await waitFor(() => {
            expect(screen.getByText('tool1')).toBeInTheDocument();
        });

        expect(apiClient.listServices).toHaveBeenCalled();

        // Select 'Service One' using the mock native select
        const select = screen.getByTestId('select-native');
        fireEvent.change(select, { target: { value: 'service1-id' } });

        await waitFor(() => {
            // tool1 and tool3 should be visible (Service One)
            expect(screen.getByText('tool1')).toBeInTheDocument();
            expect(screen.getByText('tool3')).toBeInTheDocument();
            // tool2 should not be visible (Service Two)
            expect(screen.queryByText('tool2')).not.toBeInTheDocument();
        });
    });
});
