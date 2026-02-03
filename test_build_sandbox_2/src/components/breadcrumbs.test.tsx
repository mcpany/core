/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import React from 'react';
import { render, screen, waitFor } from '@testing-library/react';
import { Breadcrumbs } from './breadcrumbs';
import userEvent from '@testing-library/user-event';
import { vi } from 'vitest';

// ResizeObserver mock
global.ResizeObserver = class ResizeObserver {
    observe() {}
    unobserve() {}
    disconnect() {}
};

// Mock pointer capture methods
Element.prototype.setPointerCapture = () => {};
Element.prototype.releasePointerCapture = () => {};
Element.prototype.hasPointerCapture = () => false;
window.HTMLElement.prototype.scrollIntoView = vi.fn();

describe('Breadcrumbs', () => {
    it('renders basic breadcrumbs', () => {
        const items = [
            { label: 'Services', href: '/services' },
            { label: 'MyService', href: '/services/my-service' },
        ];
        render(<Breadcrumbs items={items} />);

        expect(screen.getByText('Services')).toBeInTheDocument();
        expect(screen.getByText('MyService')).toBeInTheDocument();
    });

    it('renders siblings dropdown trigger when siblings are present', () => {
        const items = [
            {
                label: 'Services',
                href: '/services',
            },
            {
                label: 'MyService',
                href: '/services/my-service',
                siblings: [
                    { label: 'Service A', href: '/services/service-a' },
                    { label: 'Service B', href: '/services/service-b' },
                ],
            },
        ];
        render(<Breadcrumbs items={items} />);

        // Look for the trigger (More options)
        const trigger = screen.getByText('More options');
        expect(trigger).toBeInTheDocument();
    });

    it('opens dropdown and shows siblings on click', async () => {
        const user = userEvent.setup();
        const items = [
            {
                label: 'Services',
                href: '/services',
            },
            {
                label: 'MyService',
                href: '/services/my-service',
                siblings: [
                    { label: 'Service A', href: '/services/service-a' },
                    { label: 'Service B', href: '/services/service-b' },
                ],
            },
        ];
        render(<Breadcrumbs items={items} />);

        const trigger = screen.getByText('More options');
        await user.click(trigger);

        await waitFor(() => {
            expect(screen.getByText('Service A')).toBeInTheDocument();
            expect(screen.getByText('Service B')).toBeInTheDocument();
        });
    });
});
