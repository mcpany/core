/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import { seedUser, cleanupUser, seedCollection, cleanupCollection } from './test-data';

test.describe('Dashboard Skeletons', () => {
    test.describe.configure({ mode: 'serial' });

    test.beforeEach(async ({ request, page }) => {
        await seedCollection('mcpany-system', request);
        await seedUser(request, "e2e-admin-dashboard");
        // Login
        await page.goto('/login');
        await page.waitForLoadState('networkidle');
        await page.fill('input[name="username"]', "e2e-admin-dashboard");
        await page.fill('input[name="password"]', 'password');
        await Promise.all([
          page.waitForURL('/', { timeout: 30000 }),
          page.click('button[type="submit"]', { force: true })
        ]);
        await expect(page).toHaveURL('/', { timeout: 15000 });
    });

    test.afterEach(async ({ request }) => {
        await cleanupCollection('mcpany-system', request);
    });

    test('should display skeleton loaders while fetching dashboard data', async ({ page }) => {
        // Intercept network requests to artificially delay dashboard API responses,
        // allowing us to assert the skeletons are visible.
        await page.route('/api/v1/dashboard/metrics', async route => {
            // Delay by 1.5 seconds
            await new Promise(resolve => setTimeout(resolve, 1500));
            await route.continue();
        });

        await page.route('/api/v1/system/status', async route => {
            await new Promise(resolve => setTimeout(resolve, 1500));
            await route.continue();
        });

        await page.route('/api/v1/dashboard/health', async route => {
            await new Promise(resolve => setTimeout(resolve, 1500));
            await route.continue();
        });

        // 1. Load the dashboard
        await page.goto('/');

        // 2. Ensure Add Widget is visible, so we can ensure the widgets we want to check are present.
        const addWidgetButton = page.getByTestId('add-widget-trigger').first();
        if (await addWidgetButton.isVisible()) {
            // Wait for it to be actionable
            await addWidgetButton.waitFor({ state: 'visible' });
        }

        // 3. Verify Skeletons are visible
        // Wait for the animate-pulse class on at least one element to confirm skeletons are rendering.
        const skeletonLocator = page.locator('.animate-pulse');
        await expect(skeletonLocator.first()).toBeVisible({ timeout: 5000 });

        // Assert we see multiple skeletons
        const skeletonCount = await skeletonLocator.count();
        expect(skeletonCount).toBeGreaterThan(5);

        // 4. Wait for real data to load
        // Once the route delays finish, real data should replace skeletons.
        await expect(skeletonLocator).toHaveCount(0, { timeout: 15000 });

        // Verify a real element appeared (e.g., Total Requests string from MetricsOverview)
        const totalRequestsHeader = page.getByText(/^Total Requests$/).first();
        await expect(totalRequestsHeader).toBeVisible({ timeout: 10000 });
    });
});
