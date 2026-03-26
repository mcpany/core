/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */


import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import { StackEditor } from "../../src/components/stacks/stack-editor";
import { apiClient } from "@/lib/client";

// Mock the ConfigEditor component because Monaco is hard to test in JSDOM
vi.mock('../../src/components/stacks/config-editor', () => ({
  ConfigEditor: ({ value, onChange }: { value: string, onChange: (v: string) => void }) => (
    <textarea
      data-testid="config-editor-mock"
      role="textbox"
      value={value}
      onChange={(e) => onChange(e.target.value)}
    />
  ),
}));

// Mock StackVisualizer and ServicePalette to avoid complex dependencies
vi.mock('../../src/components/stacks/stack-visualizer', () => ({
  StackVisualizer: () => <div data-testid="stack-visualizer" />,
}));

vi.mock('../../src/components/stacks/service-palette', () => ({
  ServicePalette: ({ onTemplateSelect }: any) => (
    <div data-testid="service-palette">Service Palette</div>
  ),
}));

// Mock the apiClient
vi.mock('@/lib/client', () => ({
  apiClient: {
    getCollection: vi.fn(),
    saveCollection: vi.fn(),
  },
}));


describe('StackEditor', () => {
  const mockStackId = 'test-stack';
  const mockConfig = 'version: "1.0"\nservices:\n  test:\n    image: test/image';
  const mockCollection = {
    name: mockStackId,
    services: [
      { name: 'test', image: 'test/image' }
    ]
  };

  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(apiClient.getCollection).mockResolvedValue(mockCollection);
    vi.mocked(apiClient.saveCollection).mockResolvedValue({});
  });

  it('renders correctly and loads config', async () => {
    render(<StackEditor stackId={mockStackId} />);

    await waitFor(() => {
      expect(screen.getByText('config.yaml')).toBeDefined();
    });
    await waitFor(() => {
      expect(apiClient.getCollection).toHaveBeenCalledWith(mockStackId);
      expect(screen.getByTestId('config-editor-mock')).toBeDefined();
    });
  });

  it('validates valid YAML', async () => {
    render(<StackEditor stackId={mockStackId} />);

    await waitFor(() => {
      expect(screen.getByText('Valid YAML')).toBeDefined();
    });
  });

  it('validates valid YAML', async () => {
    render(<StackEditor stackId={mockStackId} />);

    await waitFor(() => {
      expect(screen.getByText('Valid YAML')).toBeDefined();
    });
  });
});
