/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { render, screen } from '@testing-library/react';
import { ConfigStatusBanner } from '../../components/config-status-banner';
import { useDoctor } from '../../hooks/useDoctor';

import { vi, describe, it, expect, Mock } from 'vitest';

// Mock the hook
vi.mock('../../hooks/useDoctor');

describe('ConfigStatusBanner', () => {
  it('renders nothing when report is null', () => {
    (useDoctor as Mock).mockReturnValue({ report: null });
    const { container } = render(<ConfigStatusBanner />);
    expect(container).toBeEmptyDOMElement();
  });

  it('renders nothing when configuration check is ok', () => {
    (useDoctor as Mock).mockReturnValue({
      report: {
        status: 'healthy',
        checks: {
          configuration: { status: 'ok' },
        },
      },
    });
    const { container } = render(<ConfigStatusBanner />);
    expect(container).toBeEmptyDOMElement();
  });

  it('renders alert when configuration check is degraded', () => {
    (useDoctor as Mock).mockReturnValue({
      report: {
        status: 'degraded',
        checks: {
          configuration: { status: 'degraded', message: 'Syntax error in config.yaml' },
        },
      },
    });
    render(<ConfigStatusBanner />);

    expect(screen.getByText('Configuration Error')).toBeInTheDocument();
    expect(screen.getByText(/Syntax error in config.yaml/)).toBeInTheDocument();
  });
});
