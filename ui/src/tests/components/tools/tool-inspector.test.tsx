/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { render, screen } from '@testing-library/react';
import { describe, it, expect, vi } from 'vitest';
import { ToolInspector } from '@/components/tools/tool-inspector';
import { ToolDefinition } from '@/lib/client';
import { TooltipProvider } from "@/components/ui/tooltip";

// Mock dependencies

// Mock Dialog as it uses Portals which can be tricky in tests
vi.mock('@/components/ui/dialog', () => ({
  Dialog: ({ children, open }: { children: React.ReactNode; open: boolean }) => open ? <div>{children}</div> : null,
  DialogContent: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
  DialogHeader: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
  DialogTitle: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
  DialogDescription: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
  DialogFooter: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
}));

vi.mock('@/components/ui/tabs', () => ({
  Tabs: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
  TabsList: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
  TabsTrigger: ({ children }: { children: React.ReactNode }) => <button>{children}</button>,
  TabsContent: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
}));

// Mock ResizeObserver for Recharts
global.ResizeObserver = class ResizeObserver {
  observe() {}
  unobserve() {}
  disconnect() {}
};

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
    render(
      <TooltipProvider>
        <ToolInspector tool={mockTool} open={true} onOpenChange={() => {}} />
      </TooltipProvider>
    );
    expect(screen.getByText('test_tool')).toBeDefined();
    expect(screen.getByText('A test tool')).toBeDefined();
    expect(screen.getByText('test_service')).toBeDefined();
  });

  it('renders visual editor tabs', () => {
    render(
      <TooltipProvider>
        <ToolInspector tool={mockTool} open={true} onOpenChange={() => {}} />
      </TooltipProvider>
    );
    // Check for main tabs
    expect(screen.getByText('Test & Execute')).toBeDefined();

    // Check for ToolArgumentsEditor tabs (Form/JSON/Schema)
    // Note: Since we mock Tabs, we expect to see the Trigger buttons.
    expect(screen.getByText('Form')).toBeDefined();
    expect(screen.getByText('JSON')).toBeDefined();
  });
});
