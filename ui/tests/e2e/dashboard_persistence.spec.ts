/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Dashboard Persistence', () => {
    test.describe.configure({ mode: 'serial' });

    test('should load dashboard layout from backend', async ({ page, request }) => {
        // 1. Seed preferences with a specific layout (only System Health)
        // Note: The widget type must match WIDGET_DEFINITIONS in registry
        const customLayout = [
            {
                instanceId: "test-widget-1",
                type: "system-health", // Matches type in registry
                title: "Test System Health",
                size: "full",
                hidden: false
            }
        ];

        const seedRes = await request.post('/api/v1/user/preferences', {
            data: {
                // user_id defaults to "default" if not handled by auth middleware,
                // or the test runner is authenticated as someone.
                // admin.proto defines request body as { user_id, preferences } but if using grpc-gateway
                // mapped to body="*", we send:
                user_id: "default",
                preferences: {
                    "dashboard_layout": JSON.stringify(customLayout)
                }
            }
        });
        expect(seedRes.ok()).toBeTruthy();

        // 2. Load the dashboard
        await page.goto('/');

        // 3. Verify "Test System Health" is visible
        await expect(page.getByText('Test System Health')).toBeVisible();

        // 4. Verify other widgets are NOT visible (e.g. "Recent Activity" which is default)
        await expect(page.getByText('Recent Activity')).toBeHidden();
    });

    test('should persist layout changes to backend', async ({ page, request }) => {
        // 1. Start with empty layout
        await request.post('/api/v1/user/preferences', {
            data: {
                user_id: "default",
                preferences: {
                    "dashboard_layout": "[]"
                }
            }
        });

        await page.goto('/');

        // 2. Add a widget via UI
        await page.getByRole('button', { name: 'Add Widget' }).click();
        // Click "Add" on the first available widget (e.g. System Health)
        await page.locator('button').filter({ hasText: 'Add' }).first().click();

        // Wait for debounce save (1s)
        await page.waitForTimeout(2000);

        // 3. Verify backend has updated preferences
        const fetchRes = await request.get('/api/v1/user/preferences?user_id=default');
        expect(fetchRes.ok()).toBeTruthy();
        const data = await fetchRes.json();
        const layout = JSON.parse(data.preferences.dashboard_layout);

        expect(layout.length).toBeGreaterThan(0);
        // Expect at least one widget
    });
});
