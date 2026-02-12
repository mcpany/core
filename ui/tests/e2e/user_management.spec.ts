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

        // Expect Sheet to open
        await expect(page.locator('h2:has-text("Add New User")')).toBeVisible();

        // Expect Tabs for Authentication Method
        await expect(page.locator('button[role="tab"]:has-text("API Key")')).toBeVisible();

        // Fill username
        await page.fill('input[name="id"]', 'test-api-user');

        // Select Role (Dropdown)
        await page.click('button[role="combobox"]'); // Shadcn select trigger
        await page.click('div[role="option"]:has-text("Viewer")');

        // Click API Key tab
        await page.click('button[role="tab"]:has-text("API Key")');

        // Expect Generate Button
        await expect(page.locator('button:has-text("Generate")')).toBeVisible();

        // Click Generate
        await page.click('button:has-text("Generate")');

        // Expect key to be displayed
        // We look for input[readonly] inside the sheet
        const keyInput = page.locator('input[readonly][value^="mcp_sk_"]');
        await expect(keyInput).toBeVisible();
        const key = await keyInput.inputValue();
        expect(key).toContain('mcp_sk_');

        // Save
        await page.click('button:has-text("Create User")');

        // Verify user created in the list
        // The list might need a moment to refresh, but our optimistic UI or fast backend should handle it.
        await expect(page.locator('text=test-api-user')).toBeVisible();

        // Check for API Key icon/text in the row
        // The row should contain "API Key"
        const row = page.locator('tr:has-text("test-api-user")');
        await expect(row).toContainText('API Key');
        await expect(row).toContainText('Viewer');

        // Screenshot
        await page.screenshot({ path: 'verification/users_page.png', fullPage: true });
    });
});
