import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi } from 'vitest';
import { AgentChainTracer } from './agent-chain-tracer';
import * as useTracesModule from '@/hooks/use-traces';
import React from 'react';

// Mock the hook
vi.mock('@/hooks/use-traces', () => ({
  useTraces: vi.fn()
}));

describe('AgentChainTracer', () => {
  it('renders correctly with traces', () => {
    vi.mocked(useTracesModule.useTraces).mockReturnValue({
      traces: [
        {
          id: 'trace-1',
          status: 'success',
          timestamp: new Date().toISOString(),
          totalDuration: 100,
          rootSpan: {
            name: 'test-action',
            serviceName: 'test-agent',
            input: 'test-input',
            errorMessage: null
          }
        } as any
      ],
      loading: false,
      isConnected: true,
      isPaused: false,
      setIsPaused: vi.fn(),
      clearTraces: vi.fn(),
      refresh: vi.fn()
    });

    render(<AgentChainTracer />);
    expect(screen.getByText('Agent Chain Tracer (A2A)')).toBeDefined();
    expect(screen.getByText('test-agent')).toBeDefined();
    expect(screen.getByText('test-action')).toBeDefined();
    expect(screen.getByText('100ms')).toBeDefined();
  });

  it('renders expanded details when clicked', () => {
    vi.mocked(useTracesModule.useTraces).mockReturnValue({
      traces: [
        {
          id: 'trace-1',
          status: 'error',
          timestamp: new Date().toISOString(),
          totalDuration: 150,
          rootSpan: {
            name: 'fail-action',
            serviceName: 'fail-agent',
            input: 'fail-input',
            errorMessage: 'something went wrong'
          }
        } as any
      ],
      loading: false,
      isConnected: true,
      isPaused: false,
      setIsPaused: vi.fn(),
      clearTraces: vi.fn(),
      refresh: vi.fn()
    });

    render(<AgentChainTracer />);

    // Click to expand
    fireEvent.click(screen.getByText('fail-agent'));

    // Assert details are shown
    expect(screen.getByText('something went wrong')).toBeDefined();
    expect(screen.getByText('Pending Consensus')).toBeDefined(); // error -> speculative
  });
});
