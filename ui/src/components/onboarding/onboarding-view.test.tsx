/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { OnboardingView } from './onboarding-view';
import { apiClient } from '@/lib/client';
import { vi } from 'vitest';

// Mock apiClient
vi.mock('@/lib/client', () => ({
  apiClient: {
    validateService: vi.fn(),
    registerService: vi.fn(),
  },
}));

// Mock framer-motion
vi.mock('framer-motion', () => ({
  motion: {
    div: ({ children, ...props }: any) => <div {...props}>{children}</div>,
  },
  AnimatePresence: ({ children }: any) => <>{children}</>,
}));

// Mock toast
vi.mock('@/hooks/use-toast', () => ({
    useToast: () => ({ toast: vi.fn() })
}));

describe('OnboardingView', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders the welcome message', () => {
    render(<OnboardingView onComplete={vi.fn()} />);
    expect(screen.getByText(/Welcome to MCP Any/i)).toBeInTheDocument();
    expect(screen.getByText(/Connect your first server to get started/i)).toBeInTheDocument();
  });

  it('renders selection cards', () => {
    render(<OnboardingView onComplete={vi.fn()} />);
    expect(screen.getByText(/Quick Start/i)).toBeInTheDocument();
    expect(screen.getByText(/Local Command/i)).toBeInTheDocument();
    expect(screen.getByText(/Remote Server/i)).toBeInTheDocument();
  });

  it('navigates to configuration when Quick Start is selected', async () => {
    render(<OnboardingView onComplete={vi.fn()} />);
    fireEvent.click(screen.getByText(/Quick Start/i));
    await waitFor(() => {
        expect(screen.getByText(/Configure Demo Server/i)).toBeInTheDocument();
    });
  });

  it('navigates to configuration when Local Command is selected', async () => {
    render(<OnboardingView onComplete={vi.fn()} />);
    fireEvent.click(screen.getByText(/Local Command/i));
    await waitFor(() => {
        expect(screen.getByText(/Configure Local Command/i)).toBeInTheDocument();
    });
  });

  it('handles connection flow successfully', async () => {
    const onComplete = vi.fn();
    (apiClient.validateService as any).mockResolvedValue({ valid: true });
    (apiClient.registerService as any).mockResolvedValue({});

    render(<OnboardingView onComplete={onComplete} />);

    // Select Quick Start
    fireEvent.click(screen.getByText(/Quick Start/i));

    // Click Connect (using role is safer)
    const connectBtn = screen.getByRole('button', { name: /Connect/i });
    fireEvent.click(connectBtn);

    // Wait for Success
    await waitFor(() => {
        expect(screen.getByText(/Connected Successfully!/i)).toBeInTheDocument();
    });

    // Click Go to Dashboard
    fireEvent.click(screen.getByText(/Go to Dashboard/i));
    expect(onComplete).toHaveBeenCalled();
  });

  it('handles connection failure', async () => {
    (apiClient.validateService as any).mockResolvedValue({ valid: false, message: "Network Error" });

    render(<OnboardingView onComplete={vi.fn()} />);

    // Select Quick Start
    fireEvent.click(screen.getByText(/Quick Start/i));

    // Click Connect
    const connectBtn = screen.getByRole('button', { name: /Connect/i });
    fireEvent.click(connectBtn);

    // Should stay on config or show error toast
    // Logic: setStep("configuration") on error.
    // So we check if "Configure Demo Server" is visible.
    await waitFor(() => {
        expect(screen.getByText(/Configure Demo Server/i)).toBeInTheDocument();
    });

    // Check if validateService was called
    expect(apiClient.validateService).toHaveBeenCalled();
  });
});
