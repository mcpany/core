import { render, screen, fireEvent, waitFor, within } from '@testing-library/react';
import ToolsPage from './page';
import { apiClient } from '@/lib/client';
import { vi, describe, it, expect, beforeEach } from 'vitest';

// Mock the client
vi.mock('@/lib/client', () => ({
  apiClient: {
    listTools: vi.fn(),
    setToolStatus: vi.fn(),
  },
}));

// Mock ToolInspector to avoid issues with its internal dependencies or rendering
vi.mock('@/components/tools/tool-inspector', () => ({
  ToolInspector: () => <div data-testid="tool-inspector" />,
}));

describe('ToolsPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    window.localStorage.clear();
  });

  it('renders tools, pins a tool, and filters', async () => {
    const mockTools = [
      { name: 'toolA', description: 'Description A', serviceId: 'service1', disable: false },
      { name: 'toolB', description: 'Description B', serviceId: 'service1', disable: false },
    ];

    (apiClient.listTools as any).mockResolvedValue({ tools: mockTools });

    render(<ToolsPage />);

    // Wait for tools to load
    await waitFor(() => {
      expect(screen.getByText('toolA')).toBeInTheDocument();
      expect(screen.getByText('toolB')).toBeInTheDocument();
    });

    // Verify initial order (alphabetical) - implicit by list order
    // But testing specific DOM order is better
    const rows = screen.getAllByRole('row');
    // Row 0 is header. Row 1 is toolA, Row 2 is toolB
    expect(within(rows[1]).getByText('toolA')).toBeInTheDocument();
    expect(within(rows[2]).getByText('toolB')).toBeInTheDocument();

    // Pin toolB
    const pinBtnB = screen.getByRole('button', { name: 'Pin toolB' });
    fireEvent.click(pinBtnB);

    // Verify toolB is now first
    // Note: React re-render happens.
    await waitFor(() => {
        const newRows = screen.getAllByRole('row');
        expect(within(newRows[1]).getByText('toolB')).toBeInTheDocument();
        expect(within(newRows[2]).getByText('toolA')).toBeInTheDocument();
    });

    // Toggle "Show Pinned Only"
    const showPinnedSwitch = screen.getByLabelText('Show Pinned Only');
    fireEvent.click(showPinnedSwitch);

    // Verify only toolB is visible
    await waitFor(() => {
        expect(screen.getByText('toolB')).toBeInTheDocument();
        expect(screen.queryByText('toolA')).not.toBeInTheDocument();
    });
  });
});
