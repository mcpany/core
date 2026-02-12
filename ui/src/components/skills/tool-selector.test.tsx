// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

import { render, screen, waitFor, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { ToolSelector } from '@/components/skills/tool-selector';
import { apiClient } from '@/lib/client';

// Mock dependencies
vi.mock('@/lib/client', () => ({
  apiClient: {
    listTools: vi.fn(),
  },
}));

vi.mock('sonner', () => ({
  toast: {
    error: vi.fn(),
  },
}));

describe('ToolSelector', () => {
  const mockTools = [
    { name: 'tool_a', description: 'Description A', serviceId: 'service_1' },
    { name: 'tool_b', description: 'Description B', serviceId: 'service_2' },
  ];

  beforeEach(() => {
    vi.clearAllMocks();
    (apiClient.listTools as any).mockResolvedValue({ tools: mockTools });
  });

  it('renders and fetches tools on mount', async () => {
    render(<ToolSelector selected={[]} onChange={() => {}} />);

    const trigger = screen.getByRole('combobox');
    expect(trigger).toBeDefined();

    // Open it
    fireEvent.click(trigger);

    await waitFor(() => {
        expect(apiClient.listTools).toHaveBeenCalled();
    });

    await waitFor(() => {
        // We might have badges if selected, but here selected is empty.
        // So getByText should find the list items.
        expect(screen.getByText('tool_a')).toBeDefined();
        expect(screen.getByText('tool_b')).toBeDefined();
    });
  });

  it('displays selected tools in trigger', () => {
    render(<ToolSelector selected={['tool_a']} onChange={() => {}} />);

    // The trigger is the button with role combobox
    const trigger = screen.getByRole('combobox');
    expect(trigger.textContent).toContain('tool_a');
    expect(trigger.textContent).not.toContain('tool_b');
  });

  it('calls onChange when selecting a tool', async () => {
    const onChange = vi.fn();
    render(<ToolSelector selected={[]} onChange={onChange} />);

    fireEvent.click(screen.getByRole('combobox'));

    // Wait for options to appear
    await waitFor(() => screen.getByText('tool_a'));

    // Click the option. Since nothing is selected, getByText should be unique (only in list)
    fireEvent.click(screen.getByText('tool_a'));

    expect(onChange).toHaveBeenCalledWith(['tool_a']);
  });

  it('calls onChange when deselecting a tool', async () => {
    const onChange = vi.fn();
    render(<ToolSelector selected={['tool_a']} onChange={onChange} />);

    // Trigger has "tool_a" (Badge)
    // Dropdown has "tool_a" (Option)

    fireEvent.click(screen.getByRole('combobox'));

    await waitFor(() => {
        // Wait for list to render.
        // We need to find the element in the list.
        // We can look for the description to ensure we are in the list item context?
        // Or getAllByText
        const elements = screen.getAllByText('tool_a');
        expect(elements.length).toBeGreaterThan(0);
    });

    // The option in the dropdown usually has the description nearby or we can pick the last one.
    const elements = screen.getAllByText('tool_a');
    // The last one is likely the one in the portal/popover content which is rendered last.
    fireEvent.click(elements[elements.length - 1]);

    expect(onChange).toHaveBeenCalledWith([]);
  });
});
