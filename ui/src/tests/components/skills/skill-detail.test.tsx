// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

import { render, screen, waitFor } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import SkillDetail from '@/components/skills/skill-detail';
import { SkillService } from '@/lib/skill-service';
import { apiClient } from '@/lib/client';

// Mock Next.js hooks
vi.mock('next/navigation', () => ({
  useParams: () => ({ name: 'test-skill' }),
  useRouter: () => ({ push: vi.fn() }),
}));

// Mock Link
vi.mock('next/link', () => ({
  default: ({ children }: { children: React.ReactNode }) => <a>{children}</a>,
}));

// Mock Client Services
vi.mock('@/lib/skill-service', () => ({
  SkillService: {
    get: vi.fn(),
  },
}));

vi.mock('@/lib/client', () => ({
  apiClient: {
    listTools: vi.fn(),
  },
}));

// Mock UI Components to simplify testing logic/structure
vi.mock('@/components/ui/tabs', () => ({
  Tabs: ({ children, defaultValue }: { children: React.ReactNode, defaultValue: string }) => <div data-testid="tabs" data-default={defaultValue}>{children}</div>,
  TabsList: ({ children }: { children: React.ReactNode }) => <div role="tablist">{children}</div>,
  TabsTrigger: ({ children, value }: { children: React.ReactNode, value: string }) => <button role="tab" data-value={value}>{children}</button>,
  TabsContent: ({ children, value }: { children: React.ReactNode, value: string }) => <div role="tabpanel" data-value={value}>{children}</div>,
}));

describe('SkillDetail', () => {
  const mockSkill = {
    name: 'test-skill',
    description: 'A test skill description',
    instructions: 'Do this and that.',
    allowedTools: ['tool_a', 'tool_b'],
    assets: [],
  };

  const mockTools = [
    { name: 'tool_a', description: 'Tool A', inputSchema: {} },
    // tool_b is missing
    { name: 'tool_c', description: 'Tool C', inputSchema: {} },
  ];

  beforeEach(() => {
    vi.clearAllMocks();
    (SkillService.get as any).mockResolvedValue(mockSkill);
    (apiClient.listTools as any).mockResolvedValue({ tools: mockTools });
  });

  it('renders skill details and simulation tab', async () => {
    render(<SkillDetail />);

    // Check loading state
    expect(screen.getByText('Loading skill...')).toBeDefined();

    // Wait for data
    await waitFor(() => {
      expect(screen.getByText('test-skill')).toBeDefined();
    });

    expect(screen.getByText('A test skill description')).toBeDefined();

    // Check Tabs
    expect(screen.getByRole('tab', { name: /Overview/i })).toBeDefined();
    expect(screen.getByRole('tab', { name: /Simulation & Workbench/i })).toBeDefined();
  });

  it('identifies missing tools in simulation tab', async () => {
    render(<SkillDetail />);

    await waitFor(() => {
        expect(screen.getByText('test-skill')).toBeDefined();
    });

    // In our mocked Tabs, all content is rendered but we can check if the Alert is present
    // We expect "Missing Tools Detected" because 'tool_b' is in allowedTools but not in mockTools
    expect(screen.getByText('Missing Tools Detected')).toBeDefined();
    expect(screen.getAllByText(/tool_b/).length).toBeGreaterThan(0);
  });

  it('renders system context preview', async () => {
    render(<SkillDetail />);

    await waitFor(() => {
        expect(screen.getByText('System Context Preview')).toBeDefined();
    });

    // Check for instructions
    expect(screen.getByText('Do this and that.')).toBeDefined();

    // Check for resolved tool (Tool A)
    expect(screen.getByText('Tool A')).toBeDefined();
  });
});
