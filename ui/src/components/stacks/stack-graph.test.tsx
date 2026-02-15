import React from 'react';
import { render, screen, waitFor } from '@testing-library/react';
import { StackGraph } from './stack-graph';
import { vi, describe, it, expect } from 'vitest';

// Mock ReactFlow
vi.mock('@xyflow/react', () => {
    return {
        ReactFlow: () => <div>ReactFlow Mock</div>,
        Controls: () => <div>Controls Mock</div>,
        Background: () => <div>Background Mock</div>,
        BackgroundVariant: { Dots: 'dots' },
        useNodesState: (initial: any) => React.useState(initial),
        useEdgesState: (initial: any) => React.useState(initial),
    };
});

// Mock ResizeObserver
global.ResizeObserver = vi.fn().mockImplementation(() => ({
    observe: vi.fn(),
    unobserve: vi.fn(),
    disconnect: vi.fn(),
}));

describe('StackGraph', () => {
    it('shows loading state initially', () => {
        render(<StackGraph yamlContent="" />);
        expect(screen.getByText('Processing stack...')).toBeInTheDocument();
    });

    it('renders graph after processing', async () => {
        const yaml = `
services:
  web:
    image: nginx
        `;
        render(<StackGraph yamlContent={yaml} />);

        expect(screen.getByText('Processing stack...')).toBeInTheDocument();

        await waitFor(() => {
            expect(screen.queryByText('Processing stack...')).not.toBeInTheDocument();
        });

        expect(screen.getByText('ReactFlow Mock')).toBeInTheDocument();
    });

    it('handles yaml error', async () => {
        const yaml = `
services:
  web:
    image: nginx
  invalid_yaml: [
        `; // Invalid YAML
        render(<StackGraph yamlContent={yaml} />);

        await waitFor(() => {
            expect(screen.getByText('YAML Syntax Error')).toBeInTheDocument();
        });
    });
});
