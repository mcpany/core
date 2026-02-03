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
        await page.click('button:has-text("Add User")');

        // Expect Tabs for Authentication Method
        await expect(page.locator('button[role="tab"]:has-text("API Key")')).toBeVisible();

        // Click API Key tab
        await page.click('button[role="tab"]:has-text("API Key")');

        // Expect Generate Button
        await expect(page.locator('button:has-text("Generate")')).toBeVisible();

        // Fill username
        await page.fill('input[name="id"]', 'test-api-user');
        await page.fill('input[name="role"]', 'viewer');

        // Click Generate
        await page.click('button:has-text("Generate")');

        // Expect key to be displayed
        const keyInput = page.locator('input[readonly]');
        await expect(keyInput).toBeVisible();
        const key = await keyInput.inputValue();
        expect(key).toContain('mcp_sk_');

        // Save
        await page.click('button:has-text("Save changes")');

        // Verify user created
        await expect(page.locator('text=test-api-user')).toBeVisible();
        await expect(page.locator('tr:has-text("test-api-user") >> text=API Key')).toBeVisible();

        // Screenshot

    });
});
