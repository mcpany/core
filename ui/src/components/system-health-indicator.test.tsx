import { render, screen, waitFor, fireEvent } from '@testing-library/react';
import { SystemHealthIndicator } from './system-health-indicator';
import { useSystemHealth } from '@/hooks/use-system-health';
import { vi, Mock } from 'vitest';

// Mock the hook
vi.mock('@/hooks/use-system-health');

describe('SystemHealthIndicator', () => {
  const mockUseSystemHealth = useSystemHealth as Mock;

  beforeEach(() => {
    mockUseSystemHealth.mockReset();
  });

  it('renders "Connecting..." when loading', () => {
    mockUseSystemHealth.mockReturnValue({
      report: null,
      loading: true,
      error: null,
      refresh: vi.fn(),
    });

    render(<SystemHealthIndicator />);
    expect(screen.getByText('Connecting...')).toBeInTheDocument();
  });

  it('renders "System Healthy" when status is healthy', () => {
    mockUseSystemHealth.mockReturnValue({
      report: { status: 'healthy', timestamp: '2023-01-01T00:00:00Z', checks: {} },
      loading: false,
      error: null,
      refresh: vi.fn(),
    });

    render(<SystemHealthIndicator />);
    expect(screen.getByText('System Healthy')).toBeInTheDocument();
  });

  it('renders "System Degraded" when status is not healthy', () => {
    mockUseSystemHealth.mockReturnValue({
      report: { status: 'degraded', timestamp: '2023-01-01T00:00:00Z', checks: {} },
      loading: false,
      error: null,
      refresh: vi.fn(),
    });

    render(<SystemHealthIndicator />);
    expect(screen.getByText('System Degraded')).toBeInTheDocument();
  });

  it('renders "Connection Error" when there is an error', () => {
    mockUseSystemHealth.mockReturnValue({
      report: null,
      loading: false,
      error: new Error('Failed to fetch'),
      refresh: vi.fn(),
    });

    render(<SystemHealthIndicator />);
    expect(screen.getByText('Connection Error')).toBeInTheDocument();
  });

  it('opens dialog on click and displays details', () => {
    mockUseSystemHealth.mockReturnValue({
      report: {
        status: 'healthy',
        timestamp: '2023-01-01T00:00:00Z',
        checks: {
          'database': { status: 'ok', latency: '10ms' },
          'auth_service': { status: 'degraded', message: 'Timeout' }
        }
      },
      loading: false,
      error: null,
      refresh: vi.fn(),
    });

    render(<SystemHealthIndicator />);
    fireEvent.click(screen.getByRole('button'));

    expect(screen.getByText('System Status')).toBeInTheDocument();
    expect(screen.getByText('database')).toBeInTheDocument();
    expect(screen.getByText('auth service')).toBeInTheDocument(); // capitalized and underscore replaced
    expect(screen.getByText('Timeout')).toBeInTheDocument();
  });
});
