/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { render, screen, waitFor, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi } from 'vitest';
import { AuditLogViewer } from '@/components/audit/audit-log-viewer';

// Mock dependencies
global.fetch = vi.fn();

vi.mock('@/lib/client', () => ({
  apiClient: {
    listAuditLogs: vi.fn().mockResolvedValue({
      entries: [
        {
          timestamp: '2023-01-01T12:00:00Z',
          toolName: 'test-tool',
          userId: 'user-1',
          duration: '100ms',
          durationMs: 100,
          error: null,
          arguments: JSON.stringify({ arg1: 'value1' }),
          result: { success: true, data: [1, 2, 3] },
        }
      ]
    }),
    exportAuditLogs: vi.fn()
  }
}));

describe('AuditLogViewer', () => {
  it('renders and fetches logs successfully', async () => {
    render(<AuditLogViewer />);

    // Wait for the table to load the entries
    await waitFor(() => {
      expect(screen.getByText('test-tool')).toBeInTheDocument();
    });

    // Verify columns exist
    expect(screen.getByText('Timestamp')).toBeInTheDocument();
    expect(screen.getByText('Tool')).toBeInTheDocument();
    expect(screen.getByText('User')).toBeInTheDocument();
  });

  it('opens details dialog when clicking View', async () => {
    render(<AuditLogViewer />);

    // Wait for data
    await waitFor(() => {
      expect(screen.getByText('test-tool')).toBeInTheDocument();
    });

    // Click "View" button for the first row
    const viewButtons = screen.getAllByRole('button', { name: /View/i });
    fireEvent.click(viewButtons[0]);

    // Check if dialog is opened and contains details
    await waitFor(() => {
      expect(screen.getByText('Audit Log Detail')).toBeInTheDocument();
      expect(screen.getByText('Arguments')).toBeInTheDocument();
      expect(screen.getByText('Result')).toBeInTheDocument();
    });
  });
});
