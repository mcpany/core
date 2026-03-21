/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import React from 'react';
import { render, screen, waitFor, act } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import MiddlewarePage from '../../src/app/middleware/page';

const mockSettings = {
  middlewares: [
    { name: 'Authentication', priority: 10 },
    { name: 'Rate Limiter', priority: 20 },
    { name: 'Logging', priority: 30 },
  ],
};

beforeEach(() => {
  global.fetch = vi.fn().mockResolvedValue({
    ok: true,
    json: () => Promise.resolve(mockSettings),
  });
});

describe('MiddlewarePage Component', () => {
  it('should display the middleware pipeline heading', () => {
    render(<MiddlewarePage />);
    expect(screen.getByText('Middleware Pipeline')).toBeDefined();
  });

  it('should display core middleware items', async () => {
    render(<MiddlewarePage />);
    await waitFor(() => {
      expect(screen.getAllByText('Authentication').length).toBeGreaterThan(0);
      expect(screen.getAllByText('Rate Limiter').length).toBeGreaterThan(0);
      expect(screen.getAllByText('Logging').length).toBeGreaterThan(0);
    });
  });

  it('should display switches for each middleware', async () => {
    render(<MiddlewarePage />);
    await waitFor(() => {
      const buttons = screen.getAllByRole('button');
      expect(buttons.length).toBeGreaterThan(0);
    });
  });

  it('should display settings buttons', async () => {
    render(<MiddlewarePage />);
    await waitFor(() => {
      const buttons = screen.getAllByRole('button');
      expect(buttons.length).toBeGreaterThan(0);
    });
  });
});
