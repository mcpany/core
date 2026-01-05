/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Network Graph Feature', () => {
    test('renders the network graph with core and service nodes', async ({ page }) => {
        // Go to Network page
        await page.goto('/network');

        // Check for the main container
        const graphContainer = page.locator('.react-flow');
        await expect(graphContainer).toBeVisible({ timeout: 10000 });

        // Wait for nodes to appear
        // We look for nodes by their data-id or text content
        // The core node should be labeled "MCP Gateway"
        await expect(page.getByText('MCP Gateway')).toBeVisible();

        // Check for some service nodes
        await expect(page.getByText('weather-service')).toBeVisible();
        await expect(page.getByText('postgres-db')).toBeVisible();

        // Check for the controls card
        // There are two "Network Graph" texts: one in the sidebar nav, one in the card title
        // We want to check the one in the card title which is inside a h3 or div inside the main area
        // Or simply check if there are 2 of them, which confirms the page loaded
        await expect(page.getByText('Network Graph').first()).toBeVisible();

        // Interact with a node (Click to open details)
        await page.getByText('weather-service').click();

        // Check if the side sheet opens
        await expect(page.getByRole('heading', { name: 'weather-service' })).toBeVisible();
        await expect(page.getByText('Operational Status')).toBeVisible();
    });

    test('filters work correctly', async ({ page }) => {
        await page.goto('/network');

        // Initially system nodes (Gateway) should be visible
        await expect(page.getByText('MCP Gateway')).toBeVisible();

        // Uncheck "Show System"
        await page.getByLabel('Show System (Core/Middleware)').uncheck();

        // Gateway should disappear
        await expect(page.getByText('MCP Gateway')).not.toBeVisible();

        // Re-check
        await page.getByLabel('Show System (Core/Middleware)').check();
        await expect(page.getByText('MCP Gateway')).toBeVisible();
    });
});
