import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { BulkServiceImport } from './bulk-service-import';
import { apiClient } from '@/lib/client';

// Mock dependencies
vi.mock('@/lib/client', () => ({
  apiClient: {
    registerService: vi.fn(),
  },
}));

vi.mock('@/hooks/use-toast', () => ({
  useToast: () => ({
    toast: vi.fn(),
  }),
}));

// Mock ResizeObserver for ScrollArea
global.ResizeObserver = class ResizeObserver {
  observe() {}
  unobserve() {}
  disconnect() {}
};

describe('BulkServiceImport Component', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('should enable "Preview Import" button when URL is provided', () => {
    render(<BulkServiceImport onImportSuccess={() => {}} onCancel={() => {}} />);

    const urlInput = screen.getByPlaceholderText('https://api.example.com/openapi.json');
    const previewButton = screen.getByText('Preview Import');

    // Initially disabled
    expect(previewButton).toBeDisabled();

    // Enter URL
    fireEvent.change(urlInput, { target: { value: 'https://example.com/api.json' } });

    // Should be enabled now
    expect(previewButton).not.toBeDisabled();
  });

  it('should enable "Preview Import" button when JSON is provided', () => {
    render(<BulkServiceImport onImportSuccess={() => {}} onCancel={() => {}} />);

    const jsonInput = screen.getByPlaceholderText('[{"name": "service1", "httpService": {"address": "http://..."}}, ...]');
    const previewButton = screen.getByText('Preview Import');

    fireEvent.change(jsonInput, { target: { value: '[]' } });
    expect(previewButton).not.toBeDisabled();
  });

  it('should show preview when valid JSON is entered', async () => {
    render(<BulkServiceImport onImportSuccess={() => {}} onCancel={() => {}} />);

    const jsonInput = screen.getByPlaceholderText('[{"name": "service1", "httpService": {"address": "http://..."}}, ...]');
    const previewButton = screen.getByText('Preview Import');

    const testData = [{ name: 'test-service', httpService: { address: 'http://localhost' } }];
    fireEvent.change(jsonInput, { target: { value: JSON.stringify(testData) } });

    fireEvent.click(previewButton);

    await waitFor(() => {
      expect(screen.getByText(/Found/)).toBeInTheDocument();
      expect(screen.getByText('test-service')).toBeInTheDocument();
      expect(screen.getByText('HTTP')).toBeInTheDocument();
    });

    // Check for Confirm Import button
    expect(screen.getByText('Confirm Import')).toBeInTheDocument();
  });

  it('should call registerService and onImportSuccess when Confirm Import is clicked', async () => {
    const onImportSuccess = vi.fn();
    render(<BulkServiceImport onImportSuccess={onImportSuccess} onCancel={() => {}} />);

    const jsonInput = screen.getByPlaceholderText('[{"name": "service1", "httpService": {"address": "http://..."}}, ...]');

    const testData = [{ name: 'test-service', httpService: { address: 'http://localhost' } }];
    fireEvent.change(jsonInput, { target: { value: JSON.stringify(testData) } });

    // Go to preview
    fireEvent.click(screen.getByText('Preview Import'));

    await waitFor(() => {
      expect(screen.getByText('test-service')).toBeInTheDocument();
    });

    // Click Confirm
    fireEvent.click(screen.getByText('Confirm Import'));

    await waitFor(() => {
      expect(apiClient.registerService).toHaveBeenCalledWith(expect.objectContaining({
        name: 'test-service'
      }));
    });

    // onImportSuccess is called after a timeout, so we wait a bit more
    await waitFor(() => {
        expect(onImportSuccess).toHaveBeenCalled();
    }, { timeout: 1500 });
  });
});
