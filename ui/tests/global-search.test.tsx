import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { GlobalSearch } from '../src/components/global-search';
import { apiClient } from '../src/lib/client';
import { useRouter } from 'next/navigation';

// Mock the useRouter hook
jest.mock('next/navigation', () => ({
  useRouter: jest.fn(),
}));

// Mock the apiClient
jest.mock('../src/lib/client', () => ({
  apiClient: {
    listServices: jest.fn(),
  },
}));

describe('GlobalSearch', () => {
  const mockPush = jest.fn();

  beforeEach(() => {
    (useRouter as jest.Mock).mockReturnValue({
      push: mockPush,
    });
    (apiClient.listServices as jest.Mock).mockResolvedValue({
      services: [
        { name: 'service-1', id: '1' },
        { name: 'service-2', id: '2' },
      ],
    });
  });

  afterEach(() => {
    jest.clearAllMocks();
  });

  it('renders correctly but hidden initially', () => {
    render(<GlobalSearch />);
    // The dialog content shouldn't be visible initially
    expect(screen.queryByPlaceholderText('Type a command or search...')).not.toBeInTheDocument();
  });

  it('opens when Ctrl+K is pressed', async () => {
    render(<GlobalSearch />);

    fireEvent.keyDown(document, { key: 'k', ctrlKey: true });

    await waitFor(() => {
        expect(screen.getByPlaceholderText('Type a command or search...')).toBeInTheDocument();
    });
  });

  it('fetches services when opened', async () => {
    render(<GlobalSearch />);

    fireEvent.keyDown(document, { key: 'k', ctrlKey: true });

    await waitFor(() => {
      expect(apiClient.listServices).toHaveBeenCalled();
    });
  });

  it('navigates to dashboard when selected', async () => {
    render(<GlobalSearch />);
    fireEvent.keyDown(document, { key: 'k', ctrlKey: true });

    await waitFor(() => {
        expect(screen.getByText('Dashboard')).toBeInTheDocument();
    });

    fireEvent.click(screen.getByText('Dashboard'));

    expect(mockPush).toHaveBeenCalledWith('/');
  });
});
