import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { GlobalSearch } from '../src/components/global-search';

// Mock dependencies
vi.mock('next/navigation', () => ({
  useRouter: () => ({
    push: vi.fn(),
  }),
}));

vi.mock('next-themes', () => ({
  useTheme: () => ({
    setTheme: vi.fn(),
  }),
}));

// Mock Dialog components since they rely on Radix UI which might need complex DOM setup
// or we can test the interaction logic.
// However, since we are using `cmdk` which renders into the DOM, we should be able to query it.
// The `Dialog` from shadcn usually renders in a Portal. `jsdom` supports portals.

// But there is a catch: Radix Dialog might not open immediately or might need Pointer events that JSDOM handles poorly.
// Let's try to mock the Dialog internals if standard rendering fails, but first try standard.

describe('GlobalSearch', () => {
  it('renders the search button', () => {
    render(<GlobalSearch />);
    const buttons = screen.getAllByText('Search...');
    expect(buttons.length).toBeGreaterThan(0);
    expect(buttons[0]).toBeInTheDocument();
  });

  // Note: Testing the actual opening of the dialog and interaction with `cmdk` in JSDOM
  // can be tricky due to focus management and Radix UI portals.
  // For this unit test, we primarily verify the component renders.
  // The complex interactions are better covered by the E2E tests we already have.
});
