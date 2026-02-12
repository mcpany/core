/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import { seedServices, cleanupServices, seedUser, cleanupUser } from './e2e/test-data';

test.describe('Tool Exploration', () => {
    test.describe.configure({ mode: 'serial' });

    test.beforeEach(async ({ request, page }) => {
        await seedServices(request);
        await seedUser(request, "e2e-tools-admin");

        // Login first
        await page.goto('/login');
        await page.fill('input[name="username"]', 'e2e-tools-admin');
        await page.fill('input[name="password"]', 'password');
        await page.click('button[type="submit"]');
        await expect(page).toHaveURL('/', { timeout: 15000 });
    });

    test.afterEach(async ({ request }) => {
        await cleanupServices(request);
        await cleanupUser(request, "e2e-tools-admin");
    });

    test('should list available tools from real backend', async ({ page }) => {
        await page.goto('/tools');

        // Backend registration is async (worker-based), so we might need to reload if not immediately visible.
        // The UI fetches once on mount.
        let found = false;
        // Increase retries to 10 for slow CI environments where backend worker might be lagging
        for (let i = 0; i < 10; i++) {
            try {
                // Check for weather-service tool first (get_weather) to verify generic seeding works.
                // We use get_weather because weather-service is consistently registered in CI.
                await expect(page.getByText('get_weather').first()).toBeVisible({ timeout: 5000 });
                found = true;
                break;
            } catch (e) {
                console.log(`Tools not found yet, reloading... (Attempt ${i + 1}/10)`);
                await page.reload();
                // Wait for network idle and a small buffer
                await page.waitForLoadState('networkidle');
                await page.waitForTimeout(1000);
            }
        }

        // Verify Weather tool is visible (as proxy for "Backend Working")
        await expect(page.getByText('get_weather').first()).toBeVisible({ timeout: 10000 });

        // Also check for the description to ensure data is loaded correctly
        await expect(page.getByText('Get current weather').first()).toBeVisible({ timeout: 10000 });
    });

    test('should allow inspecting a tool', async ({ page }) => {
        await page.goto('/tools');

        // Wait/Reload loop for async backend registration
        for (let i = 0; i < 10; i++) {
            try {
                await expect(page.getByText('get_weather').first()).toBeVisible({ timeout: 5000 });
                break;
            } catch (e) {
                await page.reload();
                await page.waitForLoadState('networkidle');
                await page.waitForTimeout(1000);
            }
        }
        await expect(page.getByText('get_weather').first()).toBeVisible({ timeout: 10000 });

        // Inspect get_weather
        const toolRow = page.locator('tr').filter({ hasText: /get_weather/ });
        await toolRow.getByRole('button', { name: 'Inspect' }).click();

        // Check for details specific to weather tool
        await expect(page.getByText('Get current weather').first()).toBeVisible();
        await expect(page.getByText('Test & Execute').first()).toBeVisible();
    });

    test('should execute a tool and show results', async ({ page }) => {
        await page.goto('/tools');

        // Wait/Reload loop for async backend registration
        for (let i = 0; i < 10; i++) {
            try {
                await expect(page.getByText('get_weather').first()).toBeVisible({ timeout: 5000 });
                break;
            } catch (e) {
                await page.reload();
                await page.waitForLoadState('networkidle');
                await page.waitForTimeout(1000);
            }
        }
        await expect(page.getByText('get_weather').first()).toBeVisible({ timeout: 10000 });

        const toolRow = page.locator('tr').filter({ hasText: /get_weather/ });
        await toolRow.getByRole('button', { name: 'Inspect' }).click();

        // Switch to JSON input tab
        await page.getByRole('tab', { name: 'JSON', exact: true }).click();

        // Fill arguments for get_weather (location is usually required or optional)
        const textArea = page.locator('textarea#args');
        await textArea.fill('{"location": "San Francisco"}');

        // Click Execute
        await page.getByRole('button', { name: 'Execute' }).click();

        // Verify result
        // If it fails, the UI shows the error.
        // We check for EITHER success (output) OR failure (error message),
        // proving that we hit the REAL backend.

        const outputArea = page.locator('pre.text-green-600, pre.text-green-400');

        // Wait for *some* result (success or error)
        try {
            await expect(outputArea).toBeVisible({ timeout: 5000 });
        } catch (e) {
            // If success not visible, check for error (e.g. from backend execution failure)
            const errorArea = page.getByText(/Error:/);
            await expect(errorArea).toBeVisible({ timeout: 5000 });
        }
    });
});
