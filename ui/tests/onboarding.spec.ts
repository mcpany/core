/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import { cleanupServices, seedUser, cleanupUser } from './e2e/test-data';

test.describe('Onboarding Flow', () => {
    test.beforeEach(async ({ request, page }) => {
        // Ensure NO services exist
        await cleanupServices(request);
        await seedUser(request, "e2e-admin");

        // Login
        await page.goto('/login');
        await page.fill('input[name="username"]', 'e2e-admin');
        await page.fill('input[name="password"]', 'password');
        await page.click('button[type="submit"]');
        await expect(page).toHaveURL('/', { timeout: 15000 });
    });

    test.afterEach(async ({ request }) => {
        await cleanupServices(request);
        await cleanupUser(request, "e2e-admin");
    });

    test('Shows Onboarding Wizard when no services exist', async ({ page }) => {
        await expect(page.locator('text=Welcome to MCP Any')).toBeVisible();
        await expect(page.locator('text=Quick Start (Demo)')).toBeVisible();
    });

    test('Quick Start registers services and redirects to Dashboard', async ({ page }) => {
        await expect(page.locator('text=Welcome to MCP Any')).toBeVisible();

        await page.click('button:has-text("Quick Start (Demo)")');

        // Should see success toast
        await expect(page.locator('text=System Ready')).toBeVisible();

        // Should eventually show Dashboard content (e.g. Metrics Overview)
        // Or at least NOT show Onboarding anymore.
        // DashboardGrid usually shows "Metrics Overview" widget title.
        await expect(page.locator('text=Welcome to MCP Any')).not.toBeVisible({ timeout: 10000 });

        // Check for dashboard elements
        await expect(page.locator('text=Metrics Overview')).toBeVisible();
    });
});
