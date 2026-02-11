/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import { seedUser, cleanupUser } from '../e2e/test-data';

test.describe('User Management', () => {
    test.beforeEach(async ({ request, page }) => {
        // cleanup potential leftovers
        await cleanupUser(request, "test-api-user");

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
        await expect(page.locator('div[role="dialog"]')).toBeVisible();
        await expect(page.locator('h2:has-text("Add User")')).toBeVisible();

        // Fill username
        await page.fill('input[name="id"]', 'test-api-user');

        // Select Role (it's a select trigger now)
        await page.click('button[role="combobox"]');
        await page.click('div[role="option"]:has-text("Viewer")');

        // Select API Key Tab
        await page.click('button[role="tab"]:has-text("API Key")');

        // Expect Generate Button
        await expect(page.locator('button:has-text("Generate")')).toBeVisible();

        // Click Generate
        await page.click('button:has-text("Generate")');

        // Expect key to be displayed
        const keyInput = page.locator('input[readonly]');
        await expect(keyInput).toBeVisible();
        const key = await keyInput.inputValue();
        expect(key).toContain('mcp_sk_');

        // Expect Copy button
        await expect(page.locator('button:has-text("Copy to Clipboard")')).toBeVisible();

        // Save (Button text is "Create User")
        await page.click('button:has-text("Create User")');

        // Expect Sheet to close
        await expect(page.locator('div[role="dialog"]')).toBeHidden();

        // Verify user created in the list
        // Row should contain username and "API Key" text
        const row = page.locator('tr:has-text("test-api-user")');
        await expect(row).toBeVisible();
        await expect(row).toContainText('API Key');
        await expect(row).toContainText('viewer');

        // Take screenshot for verification
        await page.screenshot({ path: 'verification.png', fullPage: true });
    });
});
