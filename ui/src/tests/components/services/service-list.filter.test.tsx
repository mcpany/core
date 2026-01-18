import { render, screen, fireEvent } from '@testing-library/react';
import { ServiceList } from '@/components/services/service-list';
import { UpstreamServiceConfig } from '@/lib/client';

const mockServices: UpstreamServiceConfig[] = [
  {
    name: 'backend-api',
    id: 'backend-api',
    version: '1.0.0',
    disable: false,
    priority: 0,
    tags: ['production', 'api', 'backend'],
    httpService: { address: 'http://localhost:8080' },
  },
  {
    name: 'frontend-ui',
    id: 'frontend-ui',
    version: '1.0.0',
    disable: false,
    priority: 0,
    tags: ['production', 'ui', 'frontend'],
    httpService: { address: 'http://localhost:3000' },
  },
  {
    name: 'worker-service',
    id: 'worker-service',
    version: '1.0.0',
    disable: false,
    priority: 0,
    tags: ['internal', 'worker'],
    commandLineService: { command: 'go run worker.go' },
  },
];

describe('ServiceList Filter', () => {
  it('filters services by tag (case insensitive)', () => {
    const { getByPlaceholderText } = render(<ServiceList services={mockServices} />);
    const input = getByPlaceholderText('Filter by tag...');

    // Initially all services are shown
    expect(screen.getByText('backend-api')).toBeDefined();
    expect(screen.getByText('frontend-ui')).toBeDefined();
    expect(screen.getByText('worker-service')).toBeDefined();

    // Filter by 'api' (lowercase)
    fireEvent.change(input, { target: { value: 'api' } });
    expect(screen.getByText('backend-api')).toBeDefined();
    expect(screen.queryByText('frontend-ui')).toBeNull();
    expect(screen.queryByText('worker-service')).toBeNull();

    // Filter by 'PRODUCTION' (uppercase)
    fireEvent.change(input, { target: { value: 'PRODUCTION' } });
    expect(screen.getByText('backend-api')).toBeDefined();
    expect(screen.getByText('frontend-ui')).toBeDefined();
    expect(screen.queryByText('worker-service')).toBeNull();

    // Filter by 'WORKER' (mixed case)
    fireEvent.change(input, { target: { value: 'WoRkEr' } });
    expect(screen.queryByText('backend-api')).toBeNull();
    expect(screen.queryByText('frontend-ui')).toBeNull();
    expect(screen.getByText('worker-service')).toBeDefined();

    // Clear filter
    fireEvent.change(input, { target: { value: '' } });
    expect(screen.getByText('backend-api')).toBeDefined();
    expect(screen.getByText('frontend-ui')).toBeDefined();
    expect(screen.getByText('worker-service')).toBeDefined();
  });
});
