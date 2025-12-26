import { render, screen } from '@testing-library/react';
import { NetworkGraphClient } from '../src/components/network/network-graph-client';

// Mock React Flow since it uses ResizeObserver which is not polyfilled in JSDOM by default
jest.mock('@xyflow/react', () => ({
  ReactFlow: ({ children }: any) => <div data-testid="react-flow">{children}</div>,
  MiniMap: () => <div data-testid="minimap" />,
  Controls: () => <div data-testid="controls" />,
  Background: () => <div data-testid="background" />,
  useNodesState: (initial: any) => [initial, jest.fn(), jest.fn()],
  useEdgesState: (initial: any) => [initial, jest.fn(), jest.fn()],
  addEdge: jest.fn(),
  Position: { Top: 'top', Bottom: 'bottom' },
  Handle: () => <div />,
}));

describe('NetworkGraphClient', () => {
  it('renders the graph container', () => {
    render(<NetworkGraphClient />);
    expect(screen.getByText('Network Graph')).toBeInTheDocument();
    expect(screen.getByText('Visualize connections between your MCP host, servers, and tools.')).toBeInTheDocument();
    expect(screen.getByTestId('react-flow')).toBeInTheDocument();
  });

  it('renders the stats panel title', () => {
    render(<NetworkGraphClient />);
    expect(screen.getByText('Live Topology')).toBeInTheDocument();
  });
});
