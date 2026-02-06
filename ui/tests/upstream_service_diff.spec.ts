/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test('Upstream Service Editor Diff Tab', async ({ page }) => {
    // Mock API
    await page.route('/api/v1/services', async route => {
        await route.fulfill({ json: { services: [] } });
    });

    // Navigate to upstream services
    await page.goto('/upstream-services');

    // Click Add Service
    await page.getByRole('button', { name: 'Add Service' }).click();

    // Select Custom Service
    await page.getByText('Custom Service').click();

    // Verify General tab
    await expect(page.getByRole('tab', { name: 'General' })).toBeVisible();

    // Change name
    await page.getByLabel('Service Name').fill('Diff Test Service');

    // Switch to Changes tab
    await page.getByRole('tab', { name: 'Changes' }).click();

    // Verify Diff Viewer is present
    await expect(page.getByText('Configuration Changes')).toBeVisible();

    // Verify Diff Editor content exists
    await expect(page.locator('.monaco-diff-editor')).toBeVisible();
});
