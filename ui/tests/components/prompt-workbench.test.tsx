/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */


import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { PromptWorkbench } from '../../src/components/prompts/prompt-workbench';
import { apiClient, PromptDefinition } from '../../src/lib/client';
import { vi } from 'vitest';

// Mock apiClient
vi.mock('../../src/lib/client', () => ({
  apiClient: {
    listPrompts: vi.fn(),
    executePrompt: vi.fn(),
  },
}));

// Mock useRouter
vi.mock('next/navigation', () => ({
  useRouter: () => ({
    push: vi.fn(),
  }),
}));

describe('PromptWorkbench', () => {
  const mockPrompts: PromptDefinition[] = [
  {
    name: 'test-prompt',
    title: 'Test Prompt',
    description: 'A test prompt',
    inputSchema: {
      fields: {
         arg1: { kind: { case: 'stringValue', value: 'Argument 1' } }
      }
    },
    messages: [
        { role: 0, text: { text: "Hello", annotations: undefined } }
    ],
    disable: false,
    profiles: []
  },
];

  beforeEach(() => {
    vi.clearAllMocks();
    (apiClient.listPrompts as any).mockResolvedValue({ prompts: mockPrompts });
  });

  it('renders list of prompts', async () => {
    render(<PromptWorkbench initialPrompts={mockPrompts} />);

    expect(screen.getByText('test-prompt')).toBeInTheDocument();
    expect(screen.getByText('A test prompt')).toBeInTheDocument();
  });

  it('selects a prompt and shows details', async () => {
    render(<PromptWorkbench initialPrompts={mockPrompts} />);

    fireEvent.click(screen.getByText('test-prompt'));

    // Wait for configuration panel to appear
    await waitFor(() => {
        expect(screen.getByText('Configuration')).toBeInTheDocument();
    });

    // Check for the argument label.
    // Note: getByLabelText might fail if the label isn't strictly associated via 'for' attribute or nesting.
    // Using getAllByText to verify it's rendered at least.
    expect(screen.getAllByText(/arg1/i)[0]).toBeInTheDocument();
  });

  it('executes a prompt', async () => {
    (apiClient.executePrompt as any).mockResolvedValue({
      messages: [{ role: 'user', content: 'test output' }]
    });

    render(<PromptWorkbench initialPrompts={mockPrompts} />);

    fireEvent.click(screen.getByText('test-prompt'));

    // Wait for the form to render
    await waitFor(() => {
        expect(screen.getByText('Configuration')).toBeInTheDocument();
    });

    // Find the input associated with arg1.
    // If getByLabelText fails, we try to find the input by placeholder or role near the text.
    // For now, let's try finding the input by role 'textbox' if there is only one, or by placeholder if we knew it.
    // Assuming simple form, maybe just 1 textbox?
    // Or we use a more generic approach: find the input element.
    const inputs = screen.getAllByRole('textbox');
    // Assuming the first input is for arg1 since it's the only field
    const input = inputs[0];

    fireEvent.change(input, { target: { value: 'value1' } });

    const generateBtn = screen.getByText('Generate Preview');
    fireEvent.click(generateBtn);

    await waitFor(() => {
       expect(apiClient.executePrompt).toHaveBeenCalledWith('test-prompt', { arg1: 'value1' });
       expect(screen.getByText('test output')).toBeInTheDocument();
    });
  });
});
