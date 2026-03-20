/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Resource Explorer', () => {

  test.beforeEach(async ({ request }) => {
    // We clear out global state to avoid clutter
    await request.post('/api/v1/debug/seed', {
        data: {
            upstream_services: [
                {
                    id: "local_fs",
                    name: "Local Filesystem",
                    filesystem_service: {
                        root_paths: {
                            "/app": "."
                        }
                    }
                }
            ],
            credentials: [],
            secrets: [],
            profiles: [],
            users: []
        },
        headers: { 'X-API-Key': process.env.MCPANY_API_KEY || 'test-token' }
    });
  });

  test('should load resources and allow selection', async ({ page }) => {

    // Navigate to the resources page
    await page.goto('/resources');

    // Use first() because the URI display might also contain the text
    await expect(page.getByText('package.json').first()).toBeVisible({ timeout: 15000 });
    await expect(page.getByText('package.json').first()).toBeVisible();

    // Verify search functionality
    const searchInput = page.getByPlaceholder('Search resources...');
    await searchInput.fill('tsconfig.json');
    await expect(page.getByText('tsconfig.json').first()).toBeVisible();
    await expect(page.getByText('package.json')).not.toBeVisible();

    // Clear search
    await searchInput.fill('');
    await expect(page.getByText('package.json').first()).toBeVisible();

    // Select a resource
    await page.getByText('package.json').first().click();

    // Verify preview loads
    // Use first() because the list item also shows the URI
    await expect(page.getByText('file:///app/package.json').first()).toBeVisible(); // URI header

    // Check if content area is visible (looking for syntax highlighter or code)
    await expect(page.locator('pre').first()).toBeVisible();

    // Verify toolbar buttons
    await expect(page.getByTitle('List View')).toBeVisible();
    await expect(page.getByTitle('Grid View')).toBeVisible();

    // Switch to Grid view
    await page.getByTitle('Grid View').click();

    // Verify grid item exists
    // In grid view, items are cards. We check for text again but layout changes.
    await expect(page.getByText('package.json').first()).toBeVisible();
  });
});
