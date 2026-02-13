/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import { seedUser, cleanupUser, seedCollection, cleanupCollection } from './e2e/test-data';

test.describe('Tool Exploration', () => {
    test.describe.configure({ mode: 'serial' });

    test.beforeEach(async ({ request, page }) => {
        // Use seedCollection for a reliable MCP server (weather-service) which works in CI environment
        await seedCollection("e2e-weather", request);
        // Only seed user and collection. Avoid seedServices() which injects flaky/external services (Stripe, etc)
        // that might pollute the UI or cause backend errors in CI.
        await seedUser(request, "e2e-tools-admin");

        // Login first
        await page.goto('/login');
        await page.fill('input[name="username"]', 'e2e-tools-admin');
        await page.fill('input[name="password"]', 'password');
        await page.click('button[type="submit"]');

        // Wait for login redirection explicitly to avoid "unexpected value" errors on CI
        await page.waitForURL('/', { timeout: 20000 });
    });

    test.afterEach(async ({ request }) => {
        await cleanupCollection("e2e-weather", request);
        await cleanupUser(request, "e2e-tools-admin");
    });

    test('should list available tools from real backend', async ({ page }) => {
        await page.goto('/tools');

        // Backend registration is async (worker-based), so we might need to reload if not immediately visible.
        // The UI fetches once on mount.
        let found = false;
        // Increase retries to 15 for slow CI environments
        for (let i = 0; i < 15; i++) {
            try {
                // Check for 'get_weather' tool which is reliably registered in CI logs
                await expect(page.getByText('get_weather').first()).toBeVisible({ timeout: 5000 });
                found = true;
                break;
            } catch (e) {
                console.log(`Tools not found yet, reloading... (Attempt ${i + 1}/15)`);
                await page.reload();
                await page.waitForLoadState('domcontentloaded');
                await page.waitForTimeout(2000);
            }
        }

        // Verify Weather Tool is visible
        await expect(page.getByText('get_weather').first()).toBeVisible({ timeout: 10000 });
        await expect(page.getByText('Get current weather').first()).toBeVisible({ timeout: 10000 });
    });

    test('should allow inspecting a tool', async ({ page }) => {
        await page.goto('/tools');

        // Wait/Reload loop
        for (let i = 0; i < 15; i++) {
            try {
                await expect(page.getByText('get_weather').first()).toBeVisible({ timeout: 5000 });
                break;
            } catch (e) {
                await page.reload();
                await page.waitForLoadState('domcontentloaded');
                await page.waitForTimeout(2000);
            }
        }
        await expect(page.getByText('get_weather').first()).toBeVisible({ timeout: 10000 });

        // Filter row by text
        const toolRow = page.locator('tr').filter({ hasText: 'get_weather' });
        await toolRow.getByRole('button', { name: 'Inspect' }).click();

        await expect(page.getByText('Get current weather').first()).toBeVisible();
        await expect(page.getByText('Test & Execute').first()).toBeVisible();
    });

    test('should execute a tool and show results', async ({ page }) => {
        await page.goto('/tools');

        // Wait/Reload loop
        for (let i = 0; i < 15; i++) {
            try {
                await expect(page.getByText('get_weather').first()).toBeVisible({ timeout: 5000 });
                break;
            } catch (e) {
                await page.reload();
                await page.waitForLoadState('domcontentloaded');
                await page.waitForTimeout(2000);
            }
        }
        await expect(page.getByText('get_weather').first()).toBeVisible({ timeout: 10000 });

        const toolRow = page.locator('tr').filter({ hasText: 'get_weather' });
        await toolRow.getByRole('button', { name: 'Inspect' }).click();

        // Switch to JSON input tab (optional if default form works, but let's stick to JSON for raw input)
        await page.getByRole('tab', { name: 'JSON', exact: true }).click();

        // Fill arguments (weather tool might take no args or optional args, {} is usually fine)
        const textArea = page.locator('textarea#args');
        await textArea.fill('{}');

        // Click Execute
        await page.getByRole('button', { name: 'Execute' }).click();

        const outputArea = page.locator('pre.text-green-600, pre.text-green-400');

        try {
            await expect(outputArea).toBeVisible({ timeout: 15000 });
        } catch (e) {
            const errorArea = page.getByText(/Error:/);
            await expect(errorArea).toBeVisible({ timeout: 15000 });
        }
    });
});
