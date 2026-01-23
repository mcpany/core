/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { ServiceTemplateSelector } from '@/components/services/service-template-selector';
import { SERVICE_TEMPLATES } from '@/lib/templates';
import { vi } from 'vitest';

describe('ServiceTemplateSelector', () => {
    it('renders all categories', () => {
        render(<ServiceTemplateSelector onSelect={vi.fn()} />);

        // "All" filter button
        expect(screen.getByRole('button', { name: 'All' })).toBeInTheDocument();

        // Check for a few known categories as filter buttons
        expect(screen.getByRole('button', { name: 'Database' })).toBeInTheDocument();
        expect(screen.getByRole('button', { name: 'Web' })).toBeInTheDocument();
    });

    it('filters templates by category', async () => {
        const user = userEvent.setup();
        render(<ServiceTemplateSelector onSelect={vi.fn()} />);

        // Initially "PostgreSQL" (Database) and "Brave Search" (Web) should be visible
        expect(screen.getByText('PostgreSQL')).toBeInTheDocument();
        expect(screen.getByText('Brave Search')).toBeInTheDocument();

        // Click "Database" category filter button
        await user.click(screen.getByRole('button', { name: 'Database' }));

        // "PostgreSQL" should still be visible
        expect(screen.getByText('PostgreSQL')).toBeInTheDocument();

        // "Brave Search" should be gone (Web category)
        expect(screen.queryByText('Brave Search')).not.toBeInTheDocument();
    });

    it('filters templates by search query', async () => {
        const user = userEvent.setup();
        render(<ServiceTemplateSelector onSelect={vi.fn()} />);

        const searchInput = screen.getByPlaceholderText('Search templates...');
        await user.type(searchInput, 'Postgre');

        expect(screen.getByText('PostgreSQL')).toBeInTheDocument();
        expect(screen.queryByText('Brave Search')).not.toBeInTheDocument();
    });

    it('calls onSelect when a template is clicked', async () => {
        const user = userEvent.setup();
        const onSelect = vi.fn();
        render(<ServiceTemplateSelector onSelect={onSelect} />);

        // Click on the PostgreSQL card title to be specific and avoid clicking badges or other elements if layout changes
        await user.click(screen.getByText('PostgreSQL'));

        const postgresTemplate = SERVICE_TEMPLATES.find(t => t.name === 'PostgreSQL');
        expect(onSelect).toHaveBeenCalledWith(postgresTemplate);
    });
});
