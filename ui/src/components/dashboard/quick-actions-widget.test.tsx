/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import React from 'react';
import { render, screen } from '@testing-library/react';
import { QuickActionsWidget } from './quick-actions-widget';
import { vi } from 'vitest';

// Mock RegisterServiceDialog to just render the trigger button
vi.mock('@/components/register-service-dialog', () => ({
  RegisterServiceDialog: ({ trigger }: { trigger: React.ReactNode }) => <div data-testid="register-dialog-trigger">{trigger}</div>,
}));

describe('QuickActionsWidget', () => {
    it('renders all quick action buttons', () => {
        render(<QuickActionsWidget />);

        expect(screen.getByText('Quick Actions')).toBeInTheDocument();
        expect(screen.getByText('Add Service')).toBeInTheDocument();
        expect(screen.getByText('Marketplace')).toBeInTheDocument();
        expect(screen.getByText('Playground')).toBeInTheDocument();
        expect(screen.getByText('All Services')).toBeInTheDocument();
        expect(screen.getByText('System Logs')).toBeInTheDocument();
        expect(screen.getByText('Secrets')).toBeInTheDocument();
    });

    it('has correct links', () => {
        render(<QuickActionsWidget />);

        // Helper to check link href
        const checkLink = (text: string, href: string) => {
            const link = screen.getByText(text).closest('a');
            expect(link).toHaveAttribute('href', href);
        };

        checkLink('Marketplace', '/marketplace');
        checkLink('Playground', '/playground');
        checkLink('All Services', '/services');
        checkLink('System Logs', '/logs');
        checkLink('Secrets', '/secrets');
    });
});
