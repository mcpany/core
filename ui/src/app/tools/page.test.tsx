/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { render, screen, fireEvent, waitFor, within } from '@testing-library/react';
import ToolsPage from './page';
import { apiClient } from '@/lib/client';
import { vi, describe, it, expect, beforeEach } from 'vitest';
import userEvent from '@testing-library/user-event';

// Mock the client
vi.mock('@/lib/client', () => ({
  apiClient: {
    listTools: vi.fn(),
    setToolStatus: vi.fn(),
  },
}));

// Mock ToolInspector to avoid issues with its internal dependencies or rendering
vi.mock('@/components/tools/tool-inspector', () => ({
  ToolInspector: () => <div data-testid="tool-inspector" />,
}));

// Mock ResizeObserver for Select component
global.ResizeObserver = class ResizeObserver {
  observe() {}
  unobserve() {}
  disconnect() {}
};

// Mock pointer capture methods
window.HTMLElement.prototype.setPointerCapture = vi.fn();
window.HTMLElement.prototype.releasePointerCapture = vi.fn();
window.HTMLElement.prototype.hasPointerCapture = vi.fn();


describe('ToolsPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    window.localStorage.clear();
  });

  it('renders tools, pins a tool, and filters by pin', async () => {
    const mockTools = [
      { name: 'toolA', description: 'Description A', serviceId: 'service1', disable: false },
      { name: 'toolB', description: 'Description B', serviceId: 'service1', disable: false },
    ];

    (apiClient.listTools as any).mockResolvedValue({ tools: mockTools });

    render(<ToolsPage />);

    // Wait for tools to load
    await waitFor(() => {
      expect(screen.getByText('toolA')).toBeInTheDocument();
      expect(screen.getByText('toolB')).toBeInTheDocument();
    });

    // Verify initial order (alphabetical) - implicit by list order
    // But testing specific DOM order is better
    const rows = screen.getAllByRole('row');
    // Row 0 is header. Row 1 is toolA, Row 2 is toolB
    expect(within(rows[1]).getByText('toolA')).toBeInTheDocument();
    expect(within(rows[2]).getByText('toolB')).toBeInTheDocument();

    // Pin toolB
    const pinBtnB = screen.getByRole('button', { name: 'Pin toolB' });
    fireEvent.click(pinBtnB);

    // Verify toolB is now first
    await waitFor(() => {
        const newRows = screen.getAllByRole('row');
        expect(within(newRows[1]).getByText('toolB')).toBeInTheDocument();
        expect(within(newRows[2]).getByText('toolA')).toBeInTheDocument();
    });

    // Toggle "Show Pinned Only"
    const showPinnedSwitch = screen.getByLabelText('Show Pinned Only');
    fireEvent.click(showPinnedSwitch);

    // Verify only toolB is visible
    await waitFor(() => {
        expect(screen.getByText('toolB')).toBeInTheDocument();
        expect(screen.queryByText('toolA')).not.toBeInTheDocument();
    });
  });

  it('filters tools by service', async () => {
    const user = userEvent.setup();
    const mockTools = [
        { name: 'github-tool', description: 'GH', serviceId: 'github', disable: false },
        { name: 'postgres-tool', description: 'DB', serviceId: 'postgres', disable: false },
    ];

    (apiClient.listTools as any).mockResolvedValue({ tools: mockTools });

    render(<ToolsPage />);

    // Wait for tools to load
    await waitFor(() => {
        expect(screen.getByText('github-tool')).toBeInTheDocument();
        expect(screen.getByText('postgres-tool')).toBeInTheDocument();
    });

    // Open Service Filter
    // The Select component uses a trigger with role 'combobox' (default for radix select)
    const trigger = screen.getByRole('combobox');
    await user.click(trigger);

    // Select 'github'
    // Radix UI Select renders options in a portal, so we query by role 'option'
    const githubOption = await screen.findByRole('option', { name: 'github' });
    await user.click(githubOption);

    // Verify filtering
    await waitFor(() => {
        expect(screen.getByText('github-tool')).toBeInTheDocument();
        expect(screen.queryByText('postgres-tool')).not.toBeInTheDocument();
    });

    // Select 'all' again
    await user.click(trigger);
    const allOption = await screen.findByRole('option', { name: 'All Services' });
    await user.click(allOption);

    // Verify all back
    await waitFor(() => {
        expect(screen.getByText('github-tool')).toBeInTheDocument();
        expect(screen.getByText('postgres-tool')).toBeInTheDocument();
    });
  });
});
