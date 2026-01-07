/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */


import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi } from 'vitest';
import { ToolInspector } from '@/components/tools/tool-inspector';
import { ToolDefinition } from '@/lib/client';

// Mock dependencies
vi.mock('@/components/ui/sheet', () => ({
  Sheet: ({ children, open }: any) => open ? <div>{children}</div> : null,
  SheetContent: ({ children }: any) => <div>{children}</div>,
  SheetHeader: ({ children }: any) => <div>{children}</div>,
  SheetTitle: ({ children }: any) => <div>{children}</div>,
  SheetDescription: ({ children }: any) => <div>{children}</div>,
}));

vi.mock('@/components/ui/tabs', () => ({
  Tabs: ({ children, defaultValue, onValueChange }: any) => <div>{children}</div>,
  TabsList: ({ children }: any) => <div>{children}</div>,
  TabsTrigger: ({ children, value }: any) => <button>{children}</button>,
  TabsContent: ({ children, value }: any) => <div>{children}</div>,
}));

// Mock fetch
global.fetch = vi.fn();

describe('ToolInspector', () => {
  const mockTool: ToolDefinition = {
    name: 'test_tool',
    description: 'A test tool',
    title: 'Test Tool',
    isStream: false,
    readOnlyHint: false,
    destructiveHint: false,
    idempotentHint: false,
    openWorldHint: false,
    callId: '',
    profiles: [],
    tags: [],
    mergeStrategy: 0,
    serviceId: 'test_service',
    disable: false,
    inputSchema: {
        type: 'object',
        properties: {
            arg1: { type: 'string', description: 'Argument 1' }
        },
        required: ['arg1']
    }
  };

  it('renders nothing when closed', () => {
    const { container } = render(<ToolInspector tool={mockTool} open={false} onOpenChange={() => {}} />);
    expect(container.firstChild).toBeNull();
  });

  it('renders tool details when open', () => {
    render(<ToolInspector tool={mockTool} open={true} onOpenChange={() => {}} />);
    expect(screen.getByText('test_tool')).toBeDefined();
    expect(screen.getByText('A test tool')).toBeDefined();
    expect(screen.getByText('test_service')).toBeDefined();
  });

  it('renders input fields based on schema', () => {
    render(<ToolInspector tool={mockTool} open={true} onOpenChange={() => {}} />);
    expect(screen.getByText('arg1')).toBeDefined();
    expect(screen.getByText('(string)')).toBeDefined();
    // In a real DOM (not mocked Sheet), we would look for the input
  });
});
