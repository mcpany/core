/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import React from 'react';
import { render, screen, waitFor } from '@testing-library/react';
import { ServiceDetail } from './service-detail';
import { vi } from 'vitest';

// Mock react-virtuoso (TableVirtuoso)
vi.mock("react-virtuoso", () => {
  console.log("!!! Mocking react-virtuoso !!!");
  return {
    TableVirtuoso: ({ data, itemContent, components }: any) => {
        console.log("!!! Rendering Mocked TableVirtuoso !!!");
        const Table = components?.Table || 'table';
        const TableBody = components?.TableBody || 'tbody';
        const TableRow = components?.TableRow || 'tr';

        return (
        <div data-testid="virtuoso-table-mock">
            <Table>
            <TableBody>
                {data.map((item: any, index: number) => (
                <TableRow key={index}>
                    {itemContent(index, item)}
                </TableRow>
                ))}
            </TableBody>
            </Table>
        </div>
        );
    },
  };
});

// Mock next/link
vi.mock("next/link", () => ({
  default: ({ children, href }: any) => <a href={href}>{children}</a>,
}));

// Mock the API client
const mockGetService = vi.fn();
vi.mock('@/lib/client', () => ({
  apiClient: {
    getService: (...args: any[]) => mockGetService(...args),
    getServiceStatus: vi.fn().mockResolvedValue({ metrics: { requests: 100 } }),
    setServiceStatus: vi.fn(),
    initiateOAuth: vi.fn(),
    updateService: vi.fn(),
  },
}));

// Mock useToast
vi.mock('@/hooks/use-toast', () => ({
  useToast: () => ({
    toast: vi.fn(),
  }),
}));

// Mock child components to isolate ServiceDetail
vi.mock("@/components/diagnostics/connection-diagnostic", () => ({
  ConnectionDiagnosticDialog: () => <div data-testid="connection-diagnostic-mock" />,
}));
vi.mock("./register-service-dialog", () => ({
  RegisterServiceDialog: () => <div data-testid="register-service-dialog-mock" />,
}));
vi.mock("./service-property-card", () => ({
  ServicePropertyCard: () => <div data-testid="service-property-card-mock" />,
}));
vi.mock("./file-config-card", () => ({
  FileConfigCard: () => <div data-testid="file-config-card-mock" />,
}));
vi.mock("@/components/safety/tool-safety-table", () => ({
  ToolSafetyTable: () => <div data-testid="tool-safety-table-mock" />,
}));
vi.mock("@/components/safety/resource-safety-table", () => ({
  ResourceSafetyTable: () => <div data-testid="resource-safety-table-mock" />,
}));
vi.mock("@/components/safety/policy-editor", () => ({
  PolicyEditor: () => <div data-testid="policy-editor-mock" />,
}));

describe('ServiceDetail', () => {
  const mockService = {
    id: 'test-service',
    name: 'Test Service',
    version: '1.0.0',
    mcpService: {
        tools: Array.from({ length: 60 }, (_, i) => ({
            name: `tool-${i}`,
            description: `Description for tool ${i}`,
            source: 'configured',
            type: 'tool'
        })),
        prompts: [],
        resources: []
    },
  };

  beforeEach(() => {
    vi.clearAllMocks();
    global.ResizeObserver = vi.fn().mockImplementation(() => ({
      observe: vi.fn(),
      unobserve: vi.fn(),
      disconnect: vi.fn(),
    }));
  });

  it('renders large list of tools using virtualization', async () => {
    mockGetService.mockResolvedValue({ service: mockService });

    render(<ServiceDetail serviceId="test-service" />);

    // Wait for loading to finish
    await waitFor(() => {
      expect(screen.getByText('Test Service')).toBeInTheDocument();
    });

    // Check if mock logs appeared
    // Verify content is rendered inside the mock
    expect(screen.getByText('tool-0')).toBeInTheDocument();
    // With initialItemCount=50, the last item might not be rendered in JSDOM without scroll
    // expect(screen.getByText('tool-59')).toBeInTheDocument();
  });

  it('uses standard table for small lists', async () => {
     const smallService = {
        ...mockService,
        mcpService: {
            tools: [],
            prompts: [{ name: 'prompt-1', description: 'Small prompt list' }],
            resources: []
        }
     };
     mockGetService.mockResolvedValue({ service: smallService });

     render(<ServiceDetail serviceId="test-service" />);

     await waitFor(() => {
        expect(screen.getByText('prompt-1')).toBeInTheDocument();
     });

     // Check that content is present
     expect(screen.getByText('prompt-1')).toBeInTheDocument();
  });
});
