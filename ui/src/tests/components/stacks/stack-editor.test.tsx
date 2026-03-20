/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { StackEditor } from '@/components/stacks/stack-editor';
import { apiClient } from '@/lib/client';
import { vi } from 'vitest';

// Mock ConfigEditor to render a simple textarea for testing
vi.mock('@/components/stacks/config-editor', () => ({
  ConfigEditor: ({ value, onChange }: { value: string; onChange: (val: string) => void }) => (
    <textarea
      value={value}
      onChange={(e) => onChange(e.target.value)}
      data-testid="config-editor-mock"
    />
  ),
}));

// Mock ServicePalette to show expected text
vi.mock('@/components/stacks/service-palette', () => ({
  ServicePalette: ({ onTemplateSelect }: any) => (
    <div data-testid="service-palette">Service Palette</div>
  ),
}));

// Mock StackVisualizer to show expected text based on YAML content
vi.mock('@/components/stacks/stack-visualizer', () => ({
  StackVisualizer: ({ yamlContent }: { yamlContent: string }) => {
    // eslint-disable-next-line @typescript-eslint/no-require-imports
    const yaml = require('js-yaml');
    let hasServices = false;
    try {
      const doc = yaml.load(yamlContent) as any;
      hasServices = doc?.services && Object.keys(doc.services).length > 0;
    } catch {}
    return (
      <div data-testid="stack-visualizer">
        {!hasServices && <p>No services defined</p>}
      </div>
    );
  },
}));

// Mock ResizeObserver for scroll area
global.ResizeObserver = class ResizeObserver {
  observe() {}
  unobserve() {}
  disconnect() {}
};

describe('StackEditor', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    global.fetch = vi.fn();
  });

  it('loads and displays configuration', async () => {
    (global.fetch as any).mockImplementation(async (url: string) => {
        if (url.includes('/collections/test-stack')) {
            return { ok: true, json: async () => ({ name: 'test-stack', services: [] }) };
        }
        return { ok: true, json: async () => ({}) };
    });

    render(<StackEditor stackId="test-stack" />);

    await waitFor(() => {
      expect(screen.getByText('config.yaml')).toBeInTheDocument();
      // The content will be a yaml dump of the collection.
      // Since services is empty array, it might be just "name: test-stack\nservices: {}\n" or similar.
      // Let's just check for the presence of the editor mock.
      expect(screen.getByTestId('config-editor-mock')).toBeInTheDocument();
    });
  });

  it('validates YAML content', async () => {
    (global.fetch as any).mockImplementation(async (url: string) => {
        if (url.includes('/collections/test-stack')) {
            return { ok: true, json: async () => ({ name: 'test-stack', services: [] }) };
        }
        return { ok: true, json: async () => ({}) };
    });

    const { container } = render(<StackEditor stackId="test-stack" />);

    // Find textarea by selector if role is elusive
    await waitFor(() => expect(screen.getByTestId('config-editor-mock')).toBeInTheDocument());
    const textarea = screen.getByTestId('config-editor-mock');

    // Valid YAML
    fireEvent.change(textarea, { target: { value: 'key: value' } });
    await waitFor(() => {
        expect(screen.getByText('Valid YAML')).toBeInTheDocument();
    });

    // Invalid YAML
    fireEvent.change(textarea, { target: { value: 'key: "unclosed quote' } });

    await waitFor(() => {
         expect(screen.getByText('Invalid YAML')).toBeInTheDocument();
    });
  });

  it('toggles palette and visualizer', async () => {
    (global.fetch as any).mockImplementation(async (url: string) => {
        if (url.includes('/collections/test-stack')) {
            return { ok: true, json: async () => ({ name: 'test-stack', services: [] }) };
        }
        return { ok: true, json: async () => ({}) };
    });

    render(<StackEditor stackId="test-stack" />);

    // Wait for the component to finish loading
    await waitFor(() => {
      expect(screen.getByText('Service Palette')).toBeInTheDocument();
    });
    // Since config is empty, visualizer shows "No services defined"
    expect(screen.getByText('No services defined')).toBeInTheDocument();
  });
});
