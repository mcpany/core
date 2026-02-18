/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('User Management', () => {
    test.beforeEach(async ({ request, page }) => {
        // Reset state
        await request.post('/api/v1/debug/reset', { headers: { 'X-API-Key': 'test-token' } });

        // Seed admin user
        const seedData = {
            users: [{
                id: "e2e-admin-users",
                authentication: {
                    basicAuth: {
                        username: "e2e-admin-users",
                        passwordHash: "$2a$12$KPRtQETm7XKJP/L6FjYYxuCFpTK/oRs7v9U6hWx9XFnWy6UuDqK/a" // "password"
                    }
                },
                roles: ["admin"]
            }]
        };
        await request.post('/api/v1/debug/seed', {
            headers: { 'X-API-Key': 'test-token' },
            data: seedData
        });

        await page.goto('/login');
        await page.fill('input[name="username"]', 'e2e-admin-users');
        await page.fill('input[name="password"]', 'password');
        await page.click('button[type="submit"]');
        await expect(page).toHaveURL('/', { timeout: 15000 });
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
        await expect(page.locator('div[role="dialog"]')).toBeHidden();

        // Verify user created in list
        // Row should contain "test-api-user"
        const row = page.locator('tr:has-text("test-api-user")');
        await expect(row).toBeVisible();

        // Row should indicate API Key auth
        await expect(row.locator('text=API Key')).toBeVisible();

        // Row should have Viewer role (default)
        await expect(row.locator('text=viewer')).toBeVisible();
    });
});
