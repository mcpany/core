/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import { seedServices, cleanupServices, seedUser, cleanupUser } from './test-data';

test.describe('Onboarding Wizard', () => {
    test.describe.configure({ mode: 'serial' });

    test.beforeEach(async ({ request, page }) => {
        // Ensure clean state
        await cleanupServices(request);
        await seedUser(request, "onboarding-user");

        // Login
        await page.goto('/login');
        await page.fill('input[name="username"]', 'onboarding-user');
        await page.fill('input[name="password"]', 'password');
        await page.click('button[type="submit"]');
        await expect(page).toHaveURL('/', { timeout: 15000 });
    });

    test.afterEach(async ({ request }) => {
        await cleanupServices(request);
        await cleanupUser(request, "onboarding-user");
    });

    test('Fresh install shows Step 1', async ({ page }) => {
        // Assert Step 1 content
        await expect(page.locator('text=Welcome to MCP Any')).toBeVisible();
        await expect(page.locator('text=1. Services').first()).toHaveClass(/text-primary/); // active
        await expect(page.locator('text=Browse Marketplace')).toBeVisible();

        // Ensure Step 2 content is NOT visible yet (we are in Step 1)
        await expect(page.locator('text=Claude Desktop')).not.toBeVisible();
    });

    test('After adding service shows Step 2', async ({ request, page }) => {
        // Seed a service
        await seedServices(request);
        await page.reload();

        // Assert Step 2 content
        await expect(page.locator('text=Great! You have active services')).toBeVisible();
        await expect(page.locator('text=2. Connect Client').first()).toHaveClass(/text-primary/);
        await expect(page.locator('text=Claude Desktop')).toBeVisible();

        // Check Config
        await expect(page.locator('pre')).toContainText('"mcpServers"');
    });
});
