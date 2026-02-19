/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import { seedUser, cleanupUser } from '../e2e/test-data';

test.describe('User Management', () => {
    test.beforeEach(async ({ request, page }) => {
        await cleanupUser(request, "test-api-user").catch(() => { });
        await seedUser(request, "e2e-admin-users");
        await page.goto('/login');
        await page.fill('input[name="username"]', 'e2e-admin-users');
        await page.fill('input[name="password"]', 'password');
        await page.click('button[type="submit"]', { force: true });
        await page.waitForURL('/', { timeout: 30000 });
        await expect(page).toHaveURL('/', { timeout: 15000 });
    });

    test.afterEach(async ({ request }) => {
        await cleanupUser(request, "e2e-admin-users").catch(() => { });
        await cleanupUser(request, "test-api-user").catch(() => { });
    });

    test('should allow creating a user with API Key', async ({ page }) => {
        await page.goto('/users');

        // Wait for list to load
        await expect(page.locator('h2:has-text("Users")')).toBeVisible();

        // Click Add User
        await page.click('button:has-text("Add User")');

        // Expect Sheet to open
        await expect(page.locator('div[role="dialog"]')).toBeVisible();
        await expect(page.locator('h2:has-text("Add New User")')).toBeVisible();

        // Fill username
        await page.fill('input[name="id"]', 'test-api-user');

        // Select API Key Tab
        await page.click('button[role="tab"]:has-text("API Key")');

        // 4. API Key Configuration
        // Ensure input is visible
        await expect(page.locator('input[placeholder="e.g. Production API Key"]')).toBeVisible({ timeout: 10000 });

        // Type key name
        await page.fill('input[placeholder="e.g. Production API Key"]', 'test-api-key');

        // Wait for and click Generate
        const generateButton = page.getByRole('button', { name: 'Generate' });
        await expect(generateButton).toBeVisible();
        await generateButton.click();

        // Wait for key to be generated (it should show the warning and code block)
        await expect(page.getByText('Warning: This key will only be shown once')).toBeVisible({ timeout: 10000 });
        const codeBlock = page.locator('pre, div').filter({ hasText: 'mcp_sk_' }).first();
        await expect(codeBlock).toBeVisible({ timeout: 10000 });

        // Save
        // Ensure the button is enabled and then click it
        const saveButton = page.getByRole('button', { name: 'Save Changes' });
        await expect(saveButton).toBeEnabled();
        await saveButton.click();

        // Verify Sheet closed
        await expect(page.locator('div[role="dialog"]')).toBeHidden({ timeout: 10000 });

        // Verify user created in list
        // Use a more specific row locator
        const row = page.locator('tr').filter({ hasText: 'test-api-user' });
        await expect(row).toBeVisible({ timeout: 15000 });

        // Row should indicate API Key auth
        await expect(row.getByText('API Key')).toBeVisible();

        // Row should have Viewer role (default)
        await expect(row.getByText('viewer', { exact: false })).toBeVisible();
    });
});
