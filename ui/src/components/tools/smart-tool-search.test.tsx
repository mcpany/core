/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */


import { render, screen, fireEvent } from '@testing-library/react';
import { SmartToolSearch } from './smart-tool-search';
import { vi, describe, it, expect } from 'vitest';
import { ToolDefinition } from '@proto/config/v1/tool';

// Mock the hook
vi.mock('@/hooks/use-recent-tools', () => ({
  useRecentTools: () => ({
    recentTools: ['tool-1'],
    addRecent: vi.fn(),
    removeRecent: vi.fn(),
    clearRecent: vi.fn(),
    isLoaded: true,
  }),
}));

const mockTools: ToolDefinition[] = [
  { name: 'tool-1', description: 'Description 1', serviceId: 'service-1' },
  { name: 'tool-2', description: 'Description 2', serviceId: 'service-1' },
  { name: 'tool-3', description: 'Description 3', serviceId: 'service-2' },
] as ToolDefinition[];

describe('SmartToolSearch', () => {
  it('renders and shows recent tools correctly', () => {
    const setSearchQuery = vi.fn();
    const onToolSelect = vi.fn();

    render(
      <SmartToolSearch
        tools={mockTools}
        searchQuery=""
        setSearchQuery={setSearchQuery}
        onToolSelect={onToolSelect}
      />
    );

    const input = screen.getByPlaceholderText('Search tools...');
    fireEvent.focus(input);

    // Should show recent tool section
    expect(screen.getByText('Recent')).toBeInTheDocument();

    // Should show the recent tool 'tool-1'
    // There might be multiple elements with text 'tool-1' (in recent and all tools),
    // so we check if at least one exists.
    const tool1Elements = screen.getAllByText('tool-1');
    expect(tool1Elements.length).toBeGreaterThan(0);

    // Should show 'tool-2' in All Tools
    expect(screen.getByText('tool-2')).toBeInTheDocument();
  });
});
