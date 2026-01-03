
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { StackEditor } from '../src/components/stacks/stack-editor';
import { apiClient } from '../src/lib/client';

// Mock the apiClient
vi.mock('../src/lib/client', () => ({
  apiClient: {
    getStackConfig: vi.fn(),
    saveStackConfig: vi.fn(),
  },
}));

describe('StackEditor', () => {
  const mockStackId = 'test-stack';
  const mockConfig = 'version: "1.0"\nservices:\n  test:\n    image: test/image';

  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(apiClient.getStackConfig).mockResolvedValue(mockConfig);
    vi.mocked(apiClient.saveStackConfig).mockResolvedValue({});
  });

  it('renders correctly and loads config', async () => {
    render(<StackEditor stackId={mockStackId} />);

    expect(screen.getByText('config.yaml')).toBeDefined();
    await waitFor(() => {
        const textarea = screen.getByRole('textbox');
        expect(textarea.textContent).toBe(mockConfig);
    });
  });

  it('validates valid YAML', async () => {
    render(<StackEditor stackId={mockStackId} />);

    await waitFor(() => {
         expect(screen.getByText('Valid YAML')).toBeDefined();
    });
  });

  it('detects invalid YAML', async () => {
    render(<StackEditor stackId={mockStackId} />);

    const textarea = await screen.findByRole('textbox');
    fireEvent.change(textarea, { target: { value: 'invalid: yaml: :' } });

    await waitFor(() => {
        expect(screen.getByText('Invalid YAML')).toBeDefined();
    });
  });

  it('saves changes when valid', async () => {
    render(<StackEditor stackId={mockStackId} />);

    const textarea = await screen.findByRole('textbox');
    const newConfig = 'version: "2.0"';
    fireEvent.change(textarea, { target: { value: newConfig } });

    const saveButton = screen.getByText('Save Changes');
    fireEvent.click(saveButton);

    await waitFor(() => {
      expect(apiClient.saveStackConfig).toHaveBeenCalledWith(mockStackId, newConfig);
    });
  });
});
