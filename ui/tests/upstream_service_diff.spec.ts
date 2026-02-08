/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import { login } from './e2e/auth-helper';
import { seedUser, cleanupUser } from './e2e/test-data';

test.beforeEach(async ({ page, request }) => {
    await seedUser(request, "e2e-admin");
    await login(page);
});

test.afterEach(async ({ request }) => {
    await cleanupUser(request, "e2e-admin");
});

test('Upstream Service Editor Diff Tab', async ({ page }) => {
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
