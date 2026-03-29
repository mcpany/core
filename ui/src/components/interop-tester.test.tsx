import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { InteropTester } from './interop-tester';
import * as client from '../lib/client';

vi.mock('../lib/client', () => ({
  fetchWithAuth: vi.fn()
}));

describe('InteropTester', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders correctly with default values', () => {
    render(<InteropTester />);
    expect(screen.getByText('Interop Tester')).toBeInTheDocument();

    const frameworkInput = screen.getByLabelText('Framework') as HTMLInputElement;
    expect(frameworkInput.value).toBe('CrewAI');

    const intentInput = screen.getByLabelText('Intent') as HTMLInputElement;
    expect(intentInput.value).toBe('task_delegation');

    const payloadInput = screen.getByLabelText('Payload Role') as HTMLInputElement;
    expect(payloadInput.value).toBe('data_analyst');
  });

  it('updates input values when typed into', () => {
    render(<InteropTester />);

    const frameworkInput = screen.getByLabelText('Framework') as HTMLInputElement;
    fireEvent.change(frameworkInput, { target: { value: 'OpenClaw' } });
    expect(frameworkInput.value).toBe('OpenClaw');

    const intentInput = screen.getByLabelText('Intent') as HTMLInputElement;
    fireEvent.change(intentInput, { target: { value: 'adaptive_reasoning' } });
    expect(intentInput.value).toBe('adaptive_reasoning');

    const payloadInput = screen.getByLabelText('Payload Role') as HTMLInputElement;
    fireEvent.change(payloadInput, { target: { value: 'tester' } });
    expect(payloadInput.value).toBe('tester');
  });

  it('submits form successfully and displays result', async () => {
    (client.fetchWithAuth as any).mockResolvedValue({
      json: vi.fn().mockResolvedValue({ status: 'success', output: 'ok' })
    });

    render(<InteropTester />);

    const submitBtn = screen.getByRole('button', { name: 'Submit Task' });
    fireEvent.click(submitBtn);

    await waitFor(() => {
      expect(client.fetchWithAuth).toHaveBeenCalledWith('/api/v1/interop/task', expect.objectContaining({
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          framework: 'CrewAI',
          intent: 'task_delegation',
          payload: { role: 'data_analyst' },
        })
      }));
    });

    await waitFor(() => {
      expect(screen.getByText('Result:')).toBeInTheDocument();
      expect(screen.getByText(/"status": "success"/)).toBeInTheDocument();
    });
  });

  it('handles fetch errors gracefully', async () => {
    (client.fetchWithAuth as any).mockRejectedValue(new Error('Network error'));

    render(<InteropTester />);

    const submitBtn = screen.getByRole('button', { name: 'Submit Task' });
    fireEvent.click(submitBtn);

    await waitFor(() => {
      expect(screen.getByText('Result:')).toBeInTheDocument();
      expect(screen.getByText(/Error submitting task: Error: Network error/)).toBeInTheDocument();
    });
  });
});
