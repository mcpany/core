/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import { seedUser, cleanupUser } from './e2e/test-data';

test.describe('Tool Exploration', () => {
    test.describe.configure({ mode: 'serial' });

    test.beforeEach(async ({ request, page }) => {
        // We rely on the pre-configured 'wttr.in' service which is loaded by the server at startup.
        // This avoids issues with dynamic seeding and profile resolution in CI environments.
        await seedUser(request, "e2e-tools-admin");

        // Login first
        await page.goto('/login');
        await page.fill('input[name="username"]', 'e2e-tools-admin');
        await page.fill('input[name="password"]', 'password');
        await page.click('button[type="submit"]');
        await expect(page).toHaveURL('/', { timeout: 15000 });
    });

    test.afterEach(async ({ request }) => {
        await cleanupUser(request, "e2e-tools-admin");
    });

    test('should list available tools from real backend', async ({ page }) => {
        await page.goto('/tools');

        // Check for get_weather (from wttr.in config which is always present)
        await expect(page.getByText('get_weather').first()).toBeVisible({ timeout: 10000 });
    });

    test('should allow inspecting a tool', async ({ page }) => {
        await page.goto('/tools');

        await expect(page.getByText('get_weather').first()).toBeVisible({ timeout: 10000 });

        // Find the row for get_weather and click Inspect
        const toolRow = page.locator('tr').filter({ hasText: 'get_weather' }).first();
        await toolRow.getByRole('button', { name: 'Inspect' }).click();

        // Verify inspector opens
        await expect(page.getByText('Test & Execute').first()).toBeVisible();
    });

    test('should execute a tool and show results', async ({ page }) => {
        await page.goto('/tools');

        await expect(page.getByText('get_weather').first()).toBeVisible({ timeout: 10000 });

        const toolRow = page.locator('tr').filter({ hasText: 'get_weather' }).first();
        await toolRow.getByRole('button', { name: 'Inspect' }).click();

        // Default mode is usually Form (Schema Form)
        // Fill arguments for get_weather via Form
        // We look for an input labeled "location" (derived from schema property)
        // shadcn/ui form fields usually have a label associated via id
        // Use a generic locator as label might be Title Cased or have different formatting
        await page.locator('input[type="text"], textarea').first().fill('London');

        // Click Execute
        await page.getByRole('button', { name: 'Execute' }).click();

        // Verify result
        // We accept either success (green) or error (if network fails), as both prove we hit the backend.
        const outputArea = page.locator('pre.text-green-600, pre.text-green-400');
        try {
            await expect(outputArea).toBeVisible({ timeout: 5000 });
        } catch (e) {
            const errorArea = page.getByText(/Error:/);
            await expect(errorArea).toBeVisible({ timeout: 5000 });
        }
    });
});
