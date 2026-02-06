/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi } from 'vitest';
import { InspectorTable } from './inspector-table';
import { Trace } from '@/types/trace';

// Mock TraceDetail
vi.mock('@/components/traces/trace-detail', () => ({
  TraceDetail: () => <div data-testid="trace-detail">Trace Detail Content</div>
}));

// Mock react-virtuoso
vi.mock("react-virtuoso", () => ({
  TableVirtuoso: ({ data, itemContent, fixedHeaderContent, components }: any) => {
      const Table = components?.Table || 'table';
      const TableBody = components?.TableBody || 'tbody';
      const TableRow = components?.TableRow || 'tr';

      return (
          <Table>
             <thead>
                {fixedHeaderContent && fixedHeaderContent()}
             </thead>
             <TableBody>
                {data.map((item: any, index: number) => {
                    // For the mock, we need to handle the TableRow component if it's passed
                    // We render the itemContent inside it
                    const content = itemContent(index, item);

                    // In the real implementation, we passed item to TableRow.
                    // We simulate that here.
                    return (
                        <TableRow key={item.id || index} item={item} onClick={() => {}}>
                            {content}
                        </TableRow>
                    );
                })}
             </TableBody>
          </Table>
      )
  }
}));

const mockTrace: Trace = {
  id: 'test-trace-1',
  rootSpan: {
    id: 'span-1',
    name: 'test-span',
    type: 'tool',
    startTime: Date.now(),
    endTime: Date.now() + 100,
    status: 'success',
  },
  timestamp: new Date().toISOString(),
  totalDuration: 100,
  status: 'success',
  trigger: 'user',
};

describe('InspectorTable', () => {
  it('renders traces correctly', async () => {
    render(<InspectorTable traces={[mockTrace]} />);

    expect(await screen.findByText('test-span')).toBeInTheDocument();
    expect(screen.getByText('test-trace-1')).toBeInTheDocument();
    expect(screen.getByText('100ms')).toBeInTheDocument();
  });

  it('renders empty state correctly', () => {
    render(<InspectorTable traces={[]} />);
    expect(screen.getByText('No traces found.')).toBeInTheDocument();
  });

  it('renders loading state correctly', () => {
    render(<InspectorTable traces={[]} loading={true} />);
    expect(screen.getByText('Loading traces...')).toBeInTheDocument();
  });

  it('opens details on click', async () => {
    render(<InspectorTable traces={[mockTrace]} />);

    // Wait for the row to appear (due to dynamic import)
    const rowText = await screen.findByText('test-span');
    const row = rowText.closest('tr');

    expect(row).not.toBeNull();
    if (row) {
        fireEvent.click(row);
    }
  });
});
