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

        // Wait for sheet to open
        await expect(page.locator('h2:has-text("Add New User")')).toBeVisible();

        // Fill username
        await page.fill('input[name="id"]', 'test-api-user');

        // Select Role (Shadcn Select)
        await page.click('button[role="combobox"]'); // Click the trigger
        await page.click('div[role="option"]:has-text("Viewer")'); // Select Viewer

        // Select API Key Auth Method (Tabs)
        await page.click('button[role="tab"]:has-text("API Key")');

        // Generate Key
        await page.click('button:has-text("Generate Key")');

        // Verify Key is shown
        // Look for the code block-like element or the Copy button appearing
        await expect(page.locator('button:has-text("Copy Key")')).toBeVisible();
        const keyDisplay = page.locator('p.font-mono');
        await expect(keyDisplay).toBeVisible();
        const key = await keyDisplay.textContent();
        expect(key).toContain('mcp_sk_');

        // Save
        await page.click('button:has-text("Save User")');

        // Wait for success toast
        await expect(page.locator('text=User Created')).toBeVisible();

        // Verify user created in list
        // UserList uses a table. Look for the row.
        await expect(page.locator('tr:has-text("test-api-user")')).toBeVisible();
        // Verify Auth Method Badge/Icon
        await expect(page.locator('tr:has-text("test-api-user") >> text=API Key')).toBeVisible();
        // Verify Role Badge
        await expect(page.locator('tr:has-text("test-api-user") >> text=viewer')).toBeVisible();
    });
});
