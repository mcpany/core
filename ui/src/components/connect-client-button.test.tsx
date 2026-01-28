/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { render, screen, fireEvent } from '@testing-library/react';
import { ConnectClientButton } from './connect-client-button';
import React from 'react';
import { vi } from 'vitest';

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

  it('allows API key input', () => {
    render(<ConnectClientButton />);
    fireEvent.click(screen.getByText('Connect'));
    const input = screen.getByPlaceholderText('Optional (if configured)');
    fireEvent.change(input, { target: { value: 'my-secret-key' } });
    expect(input).toHaveValue('my-secret-key');

    // Check if JSON updated (approximate check)
    // The JsonView renders a pre tag.
    // We expect the text to contain the api key in the URL.
    // Since JsonView renders JSON string, we can look for the string.
    // "http://localhost/sse?api_key=my-secret-key" (localhost is default in JSDOM)
    // Actually window.location.origin is "http://localhost" in JSDOM.

    // Simpler: search for the text in the document
    expect(screen.getByText(/api_key=my-secret-key/)).toBeInTheDocument();
  });
});
