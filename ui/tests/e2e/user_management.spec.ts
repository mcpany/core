/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import { seedUser, cleanupUser } from '../e2e/test-data';

test.describe('User Management', () => {
    test.beforeEach(async ({ request, page }) => {
        await seedUser(request, "e2e-admin-users");
        await page.goto('/login');
        await page.fill('input[name="username"]', 'e2e-admin-users');
        await page.fill('input[name="password"]', 'password');
        await page.click('button[type="submit"]');
        await expect(page).toHaveURL('/', { timeout: 15000 });
    });

    test.afterEach(async ({ request }) => {
        await cleanupUser(request, "e2e-admin-users");
        await cleanupUser(request, "test-api-user");
    });

    test('should allow creating a user with API Key', async ({ page }) => {
        await page.goto('/users');

        // Open the Add User Sheet
        await page.getByRole('button', { name: 'Add User' }).click();
        await expect(page.getByRole('heading', { name: 'Add New User' })).toBeVisible();

        // Fill username
        await page.getByLabel('Username').fill('test-api-user');

        // Select Role (Viewer)
        // The Select trigger usually has the placeholder or current value
        await page.getByRole('combobox').click();
        await page.getByRole('option', { name: 'Viewer' }).click();

        // Click API Key tab
        await page.getByRole('tab', { name: 'API Key' }).click();

        // Click Generate
        await page.getByRole('button', { name: 'Generate New' }).click();

        // Expect key to be displayed (it's inside a div with border)
        await expect(page.locator('text=mcp_sk_')).toBeVisible();

        // Save
        await page.getByRole('button', { name: 'Create User' }).click();

        // Verify user created in the list
        // The list might take a moment to refresh, but our optimistic update or reload happens fast
        // We use first() because the avatar cell also has the username as alt text
        await expect(page.getByRole('cell', { name: 'test-api-user' }).first()).toBeVisible();
        // Check for API Key badge/icon text
        await expect(page.getByRole('cell', { name: 'API Key' }).first()).toBeVisible();
    });
});
