/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import { seedUser, cleanupUser } from './test-data';

test.describe('Dashboard Widget Persistence', () => {
    test.beforeAll(async ({ request }) => {
        await seedUser(request);
    });

    test.afterAll(async ({ request }) => {
        await cleanupUser(request);
    });

    test('should persist hidden widgets after reload', async ({ page, request }) => {
        // First, ensure the preferences are clean/empty
        await request.post('/api/v1/user/preferences', {
            data: { "dashboard-layout": "" },
            headers: { 'X-API-Key': 'test-token' }
        });

        await page.goto('/', { waitUntil: 'networkidle' });

        // Open layout settings and hide a widget
        await page.getByRole('button', { name: 'Layout' }).click();

        // Find a specific widget checkbox and uncheck it
        const checkbox = page.locator('label', { hasText: 'Service Health' }).locator('..').locator('input[type="checkbox"]');
        await checkbox.uncheck();

        // Wait for the debounce save to trigger
        await page.waitForTimeout(1500);

        // Reload the page
        await page.reload({ waitUntil: 'networkidle' });

        // Open layout settings again
        await page.getByRole('button', { name: 'Layout' }).click();

        // Verify the checkbox remains unchecked
        const reloadedCheckbox = page.locator('label', { hasText: 'Service Health' }).locator('..').locator('input[type="checkbox"]');
        expect(await reloadedCheckbox.isChecked()).toBe(false);

        // Also check the backend state explicitly via API
        const prefsRes = await request.get('/api/v1/user/preferences', {
            headers: { 'X-API-Key': 'test-token' }
        });
        const prefs = await prefsRes.json();

        expect(prefs['dashboard-layout']).toBeDefined();
        const layout = JSON.parse(prefs['dashboard-layout']);
        const hiddenWidget = layout.find((w: { title: string, hidden: boolean }) => w.title === 'Service Health');
        expect(hiddenWidget).toBeDefined();
        expect(hiddenWidget.hidden).toBe(true);
    });
});
