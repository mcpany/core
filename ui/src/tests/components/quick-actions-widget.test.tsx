/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import React from 'react';
import { render, screen } from '@testing-library/react';
import { QuickActionsWidget } from '../../components/dashboard/quick-actions-widget';
import { describe, it, expect } from 'vitest';

describe('QuickActionsWidget', () => {
    it('should render the widget title', () => {
        render(<QuickActionsWidget />);
        expect(screen.getByText('Quick Actions')).toBeInTheDocument();
    });

    it('should render all action links', () => {
        render(<QuickActionsWidget />);

        expect(screen.getByText('Add Service')).toBeInTheDocument();
        expect(screen.getByText('Playground')).toBeInTheDocument();
        expect(screen.getByText('Manage Secrets')).toBeInTheDocument();
        expect(screen.getByText('Network Map')).toBeInTheDocument();
    });

    it('should have correct hrefs for links', () => {
        render(<QuickActionsWidget />);

        const addServiceLink = screen.getByText('Add Service').closest('a');
        expect(addServiceLink).toHaveAttribute('href', '/services');

        const playgroundLink = screen.getByText('Playground').closest('a');
        expect(playgroundLink).toHaveAttribute('href', '/playground');

        const secretsLink = screen.getByText('Manage Secrets').closest('a');
        expect(secretsLink).toHaveAttribute('href', '/secrets');

        const networkLink = screen.getByText('Network Map').closest('a');
        expect(networkLink).toHaveAttribute('href', '/network');
    });
});
