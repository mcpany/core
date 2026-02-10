/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { render, screen, fireEvent } from '@testing-library/react';
import { ConnectClientButton } from './connect-client-button';
import React from 'react';
import { vi } from 'vitest';

// Mock JsonView to simplify testing of data prop propagation
vi.mock('@/components/ui/json-view', () => ({
  JsonView: ({ data }: { data: any }) => <div data-testid="json-view">{JSON.stringify(data)}</div>,
}));

// Mock ResizeObserver which is used by some UI components but not in JSDOM
global.ResizeObserver = class ResizeObserver {
  observe() {}
  unobserve() {}
  disconnect() {}
};

// Mock clipboard
Object.assign(navigator, {
  clipboard: {
    writeText: vi.fn(),
  },
});

// Mock matchMedia
window.matchMedia = window.matchMedia || function() {
    return {
        matches: false,
        addListener: function() {},
        removeListener: function() {}
    };
};

describe('ConnectClientButton', () => {
  it('renders the connect button', () => {
    render(<ConnectClientButton />);
    // "Connect" text might be hidden on mobile but we are testing in JSDOM default (usually desktop size or generic)
    // The button has "Connect" text.
    const button = screen.getByText('Connect');
    expect(button).toBeInTheDocument();
  });

  it('opens dialog when clicked', () => {
    render(<ConnectClientButton />);
    const button = screen.getByText('Connect');
    fireEvent.click(button);
    expect(screen.getByText('Connect to MCP Any')).toBeInTheDocument();
    // Default tab is Claude
    expect(screen.getByText('Claude Desktop Configuration')).toBeInTheDocument();
  });

  it('allows API key input', async () => {
    render(<ConnectClientButton />);
    fireEvent.click(screen.getByText('Connect'));
    const input = screen.getByPlaceholderText('Optional (if configured)');
    fireEvent.change(input, { target: { value: 'my-secret-key' } });
    expect(input).toHaveValue('my-secret-key');

    // Check if the mocked JsonView contains the API key in the generated URL
    // We expect the URL http://localhost(:port)/sse?api_key=my-secret-key to be present in the JSON data
    expect(screen.getByTestId('json-view')).toHaveTextContent(/http:\/\/localhost(:\d+)?\/sse\?api_key=my-secret-key/);
  });
});
