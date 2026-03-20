import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { ServiceWizard } from './service-wizard';
import { apiClient } from '@/lib/client';
import { TooltipProvider } from '@/components/ui/tooltip';
import { ToastProvider } from '@/components/ui/toast';

// Mock the client to prevent actual network calls
vi.mock('@/lib/client', async () => {
    const actual = await vi.importActual('@/lib/client');
    return {
        ...actual,
        apiClient: {
            validateService: vi.fn(),
            registerService: vi.fn(),
        }
    };
});

describe('ServiceWizard', () => {
    beforeEach(() => {
        vi.clearAllMocks();
    });

    const renderWizard = (props = {}) => {
        const defaultProps = {
            open: true,
            onOpenChange: vi.fn(),
            onSuccess: vi.fn(),
            ...props,
        };

        // Wrap in required context providers
        return render(
            <TooltipProvider>
                <ToastProvider>
                    <ServiceWizard {...defaultProps} />
                </ToastProvider>
            </TooltipProvider>
        );
    };

    it('renders the first step (Template Selection)', () => {
        renderWizard();
        expect(screen.getByText('Choose a Template')).toBeInTheDocument();
        expect(screen.getByText('Start from Blank Service')).toBeInTheDocument();
    });

    it('progresses to step 2 when selecting "Blank Service"', async () => {
        renderWizard();
        const startBlankBtn = screen.getByText('Start from Blank Service');
        fireEvent.click(startBlankBtn);

        // Should now be on step 2
        expect(screen.getByText('Service Details')).toBeInTheDocument();
        expect(screen.getByLabelText(/Service Name/i)).toBeInTheDocument();
    });

    it('prevents progressing from step 2 if name is missing', async () => {
        renderWizard();
        // Go to step 2
        fireEvent.click(screen.getByText('Start from Blank Service'));

        // Next button should be disabled because Name is required
        const nextBtn = screen.getByRole('button', { name: /Next/i });
        expect(nextBtn).toBeDisabled();

        // Fill name
        const nameInput = screen.getByLabelText(/Service Name/i);
        fireEvent.change(nameInput, { target: { value: 'Test API' } });

        // Next button should be enabled now
        expect(nextBtn).not.toBeDisabled();
    });

    it('allows testing and saving on the final step', async () => {
        // Setup mock responses
        (apiClient.validateService as any).mockResolvedValue({ valid: true });
        (apiClient.registerService as any).mockResolvedValue({ success: true });

        const onSuccess = vi.fn();
        const onOpenChange = vi.fn();

        renderWizard({ onSuccess, onOpenChange });

        // Step 1 -> 2
        fireEvent.click(screen.getByText('Start from Blank Service'));

        // Fill out Step 2
        fireEvent.change(screen.getByLabelText(/Service Name/i), { target: { value: 'Test API' } });
        fireEvent.click(screen.getByRole('button', { name: /Next/i }));

        // Step 3 (Auth) -> 4
        fireEvent.click(screen.getByRole('button', { name: /Next/i }));

        // Step 4 (Review)
        expect(screen.getByText('Review & Create')).toBeInTheDocument();
        expect(screen.getByText('Test API')).toBeInTheDocument();

        // Test connection
        const testBtn = screen.getByRole('button', { name: /Test Connection/i });
        fireEvent.click(testBtn);
        expect(apiClient.validateService).toHaveBeenCalled();

        // Await the test resolution (using a small timeout since state updates are async)
        await new Promise(resolve => setTimeout(resolve, 0));

        // Save
        const createBtn = screen.getByRole('button', { name: /Create Service/i });
        fireEvent.click(createBtn);

        expect(apiClient.registerService).toHaveBeenCalled();

        await new Promise(resolve => setTimeout(resolve, 0));
        expect(onSuccess).toHaveBeenCalled();
        expect(onOpenChange).toHaveBeenCalledWith(false);
    });
});
