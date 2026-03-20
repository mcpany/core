/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import { seedUser, cleanupUser, seedCollection, cleanupCollection } from './test-data';

test.describe('Tools Analytics', () => {
    test.describe.configure({ mode: 'serial' });

    test.beforeEach(async ({ request, page }) => {
        await seedCollection('mcpany-system', request);
        await seedUser(request, "e2e-admin-tools");
        // Login
        await page.goto('/login');
        await page.waitForLoadState('networkidle');
        await page.fill('input[name="username"]', "e2e-admin-tools");
        await page.fill('input[name="password"]', 'password');
        await Promise.all([
            page.waitForURL('/', { timeout: 30000 }),
            page.click('button[type="submit"]', { force: true })
        ]);
        await expect(page).toHaveURL('/', { timeout: 15000 });
    });

    test.afterEach(async ({ request }) => {
        await cleanupCollection('mcpany-system', request);
        // await cleanupUser(request, "e2e-admin-tools");
    });

    test('should display seeded tool usage metrics on the tools page', async ({ page, request }) => {
        // 1. Seed usage data into the backend
        const stats = [
            {
                name: "calculator",
                serviceId: "mcpany-system",
                totalCalls: 1500,
                successRate: 85.0
            }
        ];

        const seedRes = await request.post('/api/v1/debug/seed_tool_usage', {
            data: stats,
            headers: {
                'Content-Type': 'application/json'
            }
        });
        expect(seedRes.ok()).toBeTruthy();

        // 2. Load the tools page
        await page.goto('/tools');
        await expect(page.getByRole('heading', { name: 'Tools' })).toBeVisible();

        // Wait for the tool row
        const row = page.locator('tr').filter({ hasText: 'calculator' }).first();
        await expect(row).toBeVisible({ timeout: 15000 });

        // 3. Verify metrics are displayed
        await expect(row.getByText('1,500')).toBeVisible();
        await expect(row.getByText('85.0%')).toBeVisible();
    });
});
