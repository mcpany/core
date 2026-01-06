
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { PromptWorkbench } from '../../src/components/prompts/prompt-workbench';
import { apiClient } from '../../src/lib/client';
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
  const mockPrompts = [
    {
      name: 'test-prompt',
      description: 'A test prompt',
      arguments: [
        { name: 'arg1', description: 'Argument 1', required: true },
      ],
      enabled: true,
      serviceName: 'test-service',
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

    expect(screen.getByText('Configuration')).toBeInTheDocument();
    expect(screen.getByLabelText(/arg1/i)).toBeInTheDocument();
  });

  it('executes a prompt', async () => {
    (apiClient.executePrompt as any).mockResolvedValue({
      messages: [{ role: 'user', content: 'test output' }]
    });

    render(<PromptWorkbench initialPrompts={mockPrompts} />);

    fireEvent.click(screen.getByText('test-prompt'));

    const input = screen.getByLabelText(/arg1/i);
    fireEvent.change(input, { target: { value: 'value1' } });

    const generateBtn = screen.getByText('Generate Preview');
    fireEvent.click(generateBtn);

    await waitFor(() => {
       expect(apiClient.executePrompt).toHaveBeenCalledWith('test-prompt', { arg1: 'value1' });
       expect(screen.getByText('test output')).toBeInTheDocument();
    });
  });
});
