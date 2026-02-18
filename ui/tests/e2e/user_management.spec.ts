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

        // Wait for list to load
        await expect(page.locator('h2:has-text("Users")')).toBeVisible();

        // Click Add User
        await page.click('button:has-text("Add User")');

        // Expect Sheet to open
        await expect(page.locator('div[role="dialog"]')).toBeVisible({ timeout: 60000 });
        await expect(page.locator('h2:has-text("Add New User")')).toBeVisible({ timeout: 60000 });

        // Fill username
        await page.fill('input[name="id"]', 'test-api-user');

        // Select API Key Tab
        await page.click('button[role="tab"]:has-text("API Key")');

        // Expect Generate Button
        await expect(page.locator('button:has-text("Generate Key")')).toBeVisible();

        // Click Generate
        await page.click('button:has-text("Generate Key")');

        // Expect key to be displayed in the code block
        // The code block has class "bg-muted" and contains "mcp_sk_"
        const codeBlock = page.locator('div.bg-muted:has-text("mcp_sk_")');
        await expect(codeBlock).toBeVisible();
        const key = await codeBlock.textContent();
        expect(key).toContain('mcp_sk_');

        // Expect warning
        await expect(page.locator('text=Warning: This key will only be shown once')).toBeVisible();

        // Save
        await page.click('button:has-text("Save Changes")');

        // Verify Sheet closed
        await expect(page.locator('div[role="dialog"]')).toBeHidden({ timeout: 60000 });

        // Verify user created in list
        // Row should contain "test-api-user"
        const row = page.locator('tr:has-text("test-api-user")');
        await expect(row).toBeVisible({ timeout: 60000 });

        // Row should indicate API Key auth
        await expect(row.locator('text=API Key')).toBeVisible();

        // Row should have Viewer role (default)
        await expect(row.locator('text=viewer')).toBeVisible();
    });
});
