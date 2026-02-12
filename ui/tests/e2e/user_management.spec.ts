/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import { seedUser, cleanupUser } from './test-data';

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

        // Verify Sheet opens
        await expect(page.locator('h2:has-text("Add User")')).toBeVisible();

        // Fill username
        await page.fill('input[name="id"]', 'test-api-user');

        // Select Role (Viewer by default, but let's test selecting it)
        // Click the select trigger
        await page.click('button[role="combobox"]');
        // Select the Viewer option to close the dropdown
        await page.click('div[role="option"]:has-text("Viewer")');

        // Select Auth Method API Key
        await page.click('button[role="tab"]:has-text("API Key")');

        // Click Generate Key
        await page.click('button:has-text("Generate Key")');

        // Expect key to be displayed
        const keyInput = page.locator('input[readonly]');
        await expect(keyInput).toBeVisible();
        const key = await keyInput.inputValue();
        expect(key).toContain('mcp_sk_');

        // Verify Warning Alert
        await expect(page.locator('text=Save this key now!')).toBeVisible();

        // Save
        await page.click('button:has-text("Save User")');

        // Expect success toast
        await expect(page.locator("text=User Created").first()).toBeVisible();

        // Verify user created in list
        // Use a more specific locator to avoid matching the toast message
        const row = page.locator('tr', { hasText: 'test-api-user' });
        await expect(row).toBeVisible();

        // Verify Auth Method badge/text
        await expect(row).toContainText('API Key');
        await expect(row).toContainText('Service Account'); // Subtext
    });
});
