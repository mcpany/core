import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { describe, it, expect, vi } from 'vitest';
import { InteropTester } from '../../../app/universal-agent-bus/interop-tester';
import { apiClient } from '../../../lib/client';

// Mock apiClient and useToast
vi.mock('../../../lib/client', () => ({
  apiClient: {
    postInteropTask: vi.fn(),
  },
}));

vi.mock('@/hooks/use-toast', () => ({
  useToast: () => ({
    toast: vi.fn(),
  }),
}));

describe('InteropTester', () => {
  it('renders correctly', () => {
    render(<InteropTester />);
    expect(screen.getByText('Interop Task Tester')).toBeInTheDocument();
  });

  it('handles invalid JSON payload', async () => {
    render(<InteropTester />);
    const payloadInput = screen.getByLabelText('Payload (JSON)');
    fireEvent.change(payloadInput, { target: { value: '{invalid}' } });

    const sendButton = screen.getByRole('button', { name: 'Send Task' });
    fireEvent.click(sendButton);

    expect(apiClient.postInteropTask).not.toHaveBeenCalled();
  });

  it('submits a task and displays result', async () => {
    (apiClient.postInteropTask as any).mockResolvedValue({ status: 'success', output: 'test output' });

    render(<InteropTester />);

    const sendButton = screen.getByRole('button', { name: 'Send Task' });
    fireEvent.click(sendButton);

    await waitFor(() => {
      expect(apiClient.postInteropTask).toHaveBeenCalledWith(expect.objectContaining({
        framework: 'CrewAI',
        intent: 'task_delegation',
        payload: { role: 'data_analyst' },
      }));
    });

    await waitFor(() => {
      expect(screen.getByTestId('interop-result')).toBeInTheDocument();
      expect(screen.getByTestId('interop-result')).toHaveTextContent('success');
    });
  });

  it('handles API errors', async () => {
    (apiClient.postInteropTask as any).mockRejectedValue(new Error('API error'));

    render(<InteropTester />);

    const sendButton = screen.getByRole('button', { name: 'Send Task' });
    fireEvent.click(sendButton);

    await waitFor(() => {
      expect(apiClient.postInteropTask).toHaveBeenCalled();
    });

    // Result should not be displayed if error occurs
    expect(screen.queryByTestId('interop-result')).not.toBeInTheDocument();
  });
});
