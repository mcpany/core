/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import { seedServices, cleanupServices, seedUser, cleanupUser, seedProfile, cleanupProfile } from './e2e/test-data';

test.describe('User Profile Access Control', () => {
    test.describe.configure({ mode: 'serial' });

    test.beforeEach(async ({ request }) => {
        // 1. Seed services
        await seedServices(request);

        // 2. Create a "Restricted" profile that allows ONLY "User Service" (svc_02)
        await seedProfile(request, "Restricted", ["svc_02"]);

        // 3. Create a User assigned to this Profile
        await seedUser(request, "restricted_user", "Restricted");
    });

    test.afterEach(async ({ request }) => {
        await cleanupUser(request, "restricted_user");
        await cleanupProfile(request, "Restricted");
        await cleanupServices(request);
    });

    // FIXME: This test requires running upstream services (svc_02).
    // In local dev environment without docker-compose, these services are not reachable.
    test.fixme('Restricted user sees only allowed services', async ({ page }) => {
        // Login as restricted user
        await page.goto('/login');
        await page.waitForLoadState('networkidle');

        await page.fill('input[name="username"]', 'restricted_user');
        await page.fill('input[name="password"]', 'password'); // Matches test-data hash
        await page.click('button[type="submit"]');

        // Wait for redirect
        await expect(page).toHaveURL('/', { timeout: 15000 });

        // Go to Tools page
        await page.goto('/tools');
        await page.waitForLoadState('networkidle');

        // Verify "User Service" tools are visible (e.g. "get_user")
        await expect(page.locator('text=get_user')).toBeVisible();

        // Verify "Payment Gateway" tools are NOT visible (e.g. "process_payment")
        // Note: seedServices adds "Payment Gateway" (svc_01) with "process_payment"
        await expect(page.locator('text=process_payment')).not.toBeVisible();
    });

    // FIXME: This test requires running upstream services.
    test.fixme('Admin user sees all services', async ({ page, request }) => {
        // Create an admin user without profile restrictions
        await seedUser(request, "full_admin");

        // Login
        await page.goto('/login');
        await page.fill('input[name="username"]', 'full_admin');
        await page.fill('input[name="password"]', 'password');
        await page.click('button[type="submit"]');
        await expect(page).toHaveURL('/', { timeout: 15000 });

        // Tools page
        await page.goto('/tools');

        // Verify ALL tools visible
        await expect(page.locator('text=get_user')).toBeVisible();
        await expect(page.locator('text=process_payment')).toBeVisible();

        await cleanupUser(request, "full_admin");
    });
});
