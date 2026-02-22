import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { SetupWizard } from './setup-wizard';
import { apiClient } from '@/lib/client';
import { vi } from 'vitest';

// Mock apiClient
vi.mock('@/lib/client', () => ({
  apiClient: {
    registerService: vi.fn(),
  },
}));

// Mock useRouter
vi.mock('next/navigation', () => ({
  useRouter: () => ({
    push: vi.fn(),
  }),
}));

// Mock toast
vi.mock('@/hooks/use-toast', () => ({
  useToast: () => ({
    toast: vi.fn(),
  }),
}));

describe('SetupWizard', () => {
  it('renders welcome step initially', () => {
    render(<SetupWizard />);
    expect(screen.getByText('Welcome to MCP Any')).toBeInTheDocument();
  });

  it('navigates to selection step', () => {
    render(<SetupWizard />);
    fireEvent.click(screen.getByRole('button', { name: /Get Started/i }));
    expect(screen.getByText('Choose Connection Type')).toBeInTheDocument();
  });

  it('selects Local Command and navigates to config', () => {
    render(<SetupWizard />);
    fireEvent.click(screen.getByRole('button', { name: /Get Started/i }));

    // Select Local (it's radio group)
    fireEvent.click(screen.getByText('Local Command'));
    fireEvent.click(screen.getByRole('button', { name: /Continue/i }));

    expect(screen.getByText('Configure Local Service')).toBeInTheDocument();
    expect(screen.getByLabelText('Command')).toBeInTheDocument();
  });

  it('submits configuration and calls API', async () => {
    (apiClient.registerService as any).mockResolvedValue({});

    render(<SetupWizard />);
    fireEvent.click(screen.getByRole('button', { name: /Get Started/i }));
    fireEvent.click(screen.getByText('Local Command'));
    fireEvent.click(screen.getByRole('button', { name: /Continue/i }));

    fireEvent.change(screen.getByLabelText('Service Name'), { target: { value: 'My Service' } });
    fireEvent.change(screen.getByLabelText('Command'), { target: { value: 'echo test' } });

    fireEvent.click(screen.getByRole('button', { name: /Connect Service/i }));

    await waitFor(() => {
      expect(apiClient.registerService).toHaveBeenCalledWith(expect.objectContaining({
        name: 'My Service',
        commandLineService: expect.objectContaining({
          command: 'echo test'
        })
      }));
    });

    expect(screen.getByText('Setup Complete!')).toBeInTheDocument();
  });
});
