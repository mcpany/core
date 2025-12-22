
import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { GlobalSearch } from '../components/global-search';
import { apiClient } from '../lib/client';

// Mock dependencies
jest.mock('next/navigation', () => ({
  useRouter: () => ({
    push: jest.fn(),
  }),
}));

jest.mock('../lib/client', () => ({
  apiClient: {
    listServices: jest.fn(),
  },
}));

// Mock the Command component parts because they use Radix primitives which can be tricky in JSDOM
// However, since we installed the files in ui/src/components/ui/command.tsx, they might work if environment is right.
// But usually for unit testing logic, shallow mocks are safer.
// Actually, let's try to test the integration first. If it fails, we'll mock.

describe('GlobalSearch', () => {
  beforeEach(() => {
    (apiClient.listServices as jest.Mock).mockResolvedValue({
      services: [
        { name: 'test-service', url: 'http://localhost:3000' }
      ]
    });
  });

  it('renders nothing initially', () => {
    render(<GlobalSearch />);
    // Dialog should be closed by default
    const dialog = screen.queryByRole('dialog');
    expect(dialog).not.toBeInTheDocument();
  });

  // Note: Radix Dialog uses portals, and might be hard to test open state without user interaction simulation
  // and proper setup.
});
