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

        // Open the Sheet
        await page.click('button:has-text("Add User")');
        await expect(page.getByRole('dialog')).toBeVisible(); // Sheet is a dialog role
        await expect(page.getByRole('heading', { name: 'Add User' })).toBeVisible(); // Sheet Title

        // Fill Form
        await page.fill('input[name="id"]', 'test-api-user');

        // Select Role (Select component is tricky in Playwright, usually hidden input)
        // We need to click the trigger
        await page.click('button[role="combobox"]'); // Select trigger
        await page.click('div[role="option"]:has-text("Viewer")'); // Select option

        // Switch to API Key Tab
        await page.click('button[role="tab"]:has-text("API Key")');

        // Expect Generate Button
        const generateBtn = page.locator('button:has-text("Generate")');
        await expect(generateBtn).toBeVisible();

        // Click Generate
        await generateBtn.click();

        // Expect key to be displayed
        const keyInput = page.locator('input[readonly]');
        await expect(keyInput).toBeVisible();
        const key = await keyInput.inputValue();
        expect(key).toContain('mcp_sk_');

        // Verify Copy Button appears
        const copyBtn = page.locator('button:has-text("Copy Key")');
        await expect(copyBtn).toBeVisible();

        // Save
        await page.click('button:has-text("Save User")');

        // Verify Sheet closed
        await expect(page.getByRole('dialog')).toBeHidden();

        // Verify user created in list
        // Row should contain username and "API Key" badge/text
        const row = page.locator('tr:has-text("test-api-user")');
        await expect(row).toBeVisible();
        await expect(row).toContainText('API Key');
        await expect(row).toContainText('viewer'); // Role badge
    });

    test('should validate form inputs', async ({ page }) => {
        await page.goto('/users');
        await page.click('button:has-text("Add User")');

        // Submit empty form
        await page.click('button:has-text("Save User")');

        // Expect validation errors
        await expect(page.getByText('Username must be at least 3 characters')).toBeVisible();

        // Password is required for new user (default tab)
        await expect(page.getByText('Password is required for new users')).toBeVisible();
    });
});
