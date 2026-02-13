// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { vi, describe, it, expect } from 'vitest';
import SkillWizard from '@/components/skills/skill-wizard';
import { apiClient } from '@/lib/client';
import { SkillService } from '@/lib/skill-service';

// Mock Dependencies
vi.mock('next/navigation', () => ({
  useRouter: () => ({ push: vi.fn() }),
  useParams: () => ({ name: undefined }),
}));

vi.mock('@/lib/client', () => ({
  apiClient: {
    listTools: vi.fn(),
  },
}));

vi.mock('@/lib/skill-service', () => ({
  SkillService: {
    create: vi.fn(),
    get: vi.fn(),
    update: vi.fn(),
    uploadAsset: vi.fn(),
  },
}));

// Mock MarkdownEditor as it relies on react-syntax-highlighter which might be heavy or need mocks
vi.mock('@/components/markdown-editor', () => ({
  MarkdownEditor: ({ value, onChange, placeholder }: any) => (
    <textarea
      data-testid="markdown-editor"
      value={value}
      onChange={(e) => onChange(e.target.value)}
      placeholder={placeholder}
    />
  ),
}));

// Mock MultiSelect
vi.mock('@/components/ui/multi-select', () => ({
  MultiSelect: ({ options, selected, onChange }: any) => (
    <div data-testid="multi-select">
      {options.map((opt: any) => (
        <button
          key={opt.value}
          onClick={() => {
             const newSelected = selected.includes(opt.value)
                ? selected.filter((s: string) => s !== opt.value)
                : [...selected, opt.value];
             onChange(newSelected);
          }}
          data-selected={selected.includes(opt.value)}
        >
          {opt.label}
        </button>
      ))}
    </div>
  ),
}));

describe('SkillWizard', () => {
  it('loads tools on mount', async () => {
    (apiClient.listTools as any).mockResolvedValue({
      tools: [
        { name: 'tool1', description: 'desc1' },
        { name: 'tool2', description: 'desc2' },
      ],
    });

    render(<SkillWizard />);

    await waitFor(() => {
      expect(apiClient.listTools).toHaveBeenCalled();
    });

    // Check if tools are rendered in our mock MultiSelect
    expect(screen.getByText('tool1')).toBeInTheDocument();
    expect(screen.getByText('tool2')).toBeInTheDocument();
  });

  it('allows filling metadata and selecting tools', async () => {
    (apiClient.listTools as any).mockResolvedValue({
        tools: [{ name: 'tool1' }],
    });

    render(<SkillWizard />);

    // Metadata Step
    fireEvent.change(screen.getByLabelText('Skill Name (ID)'), { target: { value: 'my-skill' } });
    fireEvent.change(screen.getByLabelText('Description'), { target: { value: 'Test Description' } });

    await waitFor(() => expect(screen.getByText('tool1')).toBeInTheDocument());
    fireEvent.click(screen.getByText('tool1'));

    fireEvent.click(screen.getByText('Next'));

    // Instructions Step
    await waitFor(() => expect(screen.getByText('Instructions')).toBeInTheDocument());
    const editor = screen.getByTestId('markdown-editor');
    fireEvent.change(editor, { target: { value: 'Do this.' } });

    fireEvent.click(screen.getByText('Next'));

    // Assets Step
    await waitFor(() => expect(screen.getByText('Existing Assets')).toBeInTheDocument());

    // Save
    fireEvent.click(screen.getByText('Create Skill'));

    await waitFor(() => {
        expect(SkillService.create).toHaveBeenCalledWith({
            name: 'my-skill',
            description: 'Test Description',
            allowedTools: ['tool1'],
            instructions: 'Do this.',
            assets: [],
        });
    });
  });
});
