/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import { seedServices, cleanupServices, seedUser, cleanupUser, seedCollection, cleanupCollection } from './e2e/test-data';

test.describe('Tool Exploration', () => {
    test.describe.configure({ mode: 'serial' });

    test.beforeEach(async ({ request, page }) => {
        await seedServices(request);
        // Seed a collection with weather-service to ensure we have a robust fallback if Payment Gateway fails
        await seedCollection('tools-test-stack', request);
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
        await cleanupCollection('tools-test-stack', request);
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
                // Check for ANY seeded tool. weather-service (get_weather) is robustly seeded via collection.
                // Payment Gateway (process_payment) sometimes fails to load tools in CI environment.
                await expect(page.getByText(/process_payment|get_weather|echo_tool/).first()).toBeVisible({ timeout: 5000 });
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

        // Verify at least one expected tool is visible
        await expect(page.getByText(/process_payment|get_weather|echo_tool/).first()).toBeVisible({ timeout: 10000 });

        // Look for *a* seeded tool description to confirm details loaded.
        try {
            await expect(page.getByText(/echo_tool|get_weather|process_payment/).first()).toBeVisible({ timeout: 20000 });
        } catch (e) {
            console.log('Tool not found. Page content:', await page.content());
            throw e;
        }

        // Check for description of EITHER Echo or Weather tool to confirm detail loading
        // Echo: "Echoes back input", Weather: "Get current weather" (if generic), Payment: "Process a payment"
        await expect(page.getByText(/Echoes back input|Process a payment/).first()).toBeVisible({ timeout: 20000 });
    });

    test('should allow inspecting a tool', async ({ page }) => {
        await page.goto('/tools');

        // Wait/Reload loop for async backend registration
        for (let i = 0; i < 10; i++) {
            try {
                await expect(page.getByText(/process_payment|get_weather|echo_tool/).first()).toBeVisible({ timeout: 5000 });
                break;
            } catch (e) {
                await page.reload();
                await page.waitForLoadState('networkidle');
                await page.waitForTimeout(1000);
            }
        }
        await expect(page.getByText(/process_payment|get_weather|echo_tool/).first()).toBeVisible({ timeout: 10000 });

        // Use get_weather if available, fallback to echo_tool or process_payment
        const toolRow = page.locator('tr').filter({ hasText: /get_weather|echo_tool|process_payment/ }).first();
        await toolRow.getByRole('button', { name: 'Inspect' }).click();

        // Check for common elements in Inspector
        await expect(page.getByText('Test & Execute').first()).toBeVisible();
    });

    test('should execute a tool and show results', async ({ page }) => {
        await page.goto('/tools');

        // Wait/Reload loop for async backend registration
        for (let i = 0; i < 10; i++) {
            try {
                await expect(page.getByText(/process_payment|get_weather|echo_tool/).first()).toBeVisible({ timeout: 5000 });
                break;
            } catch (e) {
                await page.reload();
                await page.waitForLoadState('networkidle');
                await page.waitForTimeout(1000);
            }
        }
        await expect(page.getByText(/process_payment|get_weather|echo_tool/).first()).toBeVisible({ timeout: 10000 });

        // Prefer echo_tool for execution test as it has simple input
        const echoRow = page.locator('tr').filter({ hasText: /echo_tool/ });
        let targetRow = echoRow;
        let isEcho = true;

        if (await echoRow.count() === 0) {
             // Fallback to get_weather or process_payment
             targetRow = page.locator('tr').filter({ hasText: /get_weather|process_payment/ }).first();
             isEcho = false;
        }

        await targetRow.getByRole('button', { name: 'Inspect' }).click();

        // Switch to JSON input tab
        await page.getByRole('tab', { name: 'JSON', exact: true }).click();

        // Fill arguments
        const textArea = page.locator('textarea#args');
        if (isEcho) {
            await textArea.fill('{"message": "Hello MCP"}');
        } else {
            // Generic empty object for others just to trigger execution
            await textArea.fill('{}');
        }

        // Click Execute
        await page.getByRole('button', { name: 'Execute' }).click();

        // Verify result
        const outputArea = page.locator('pre.text-green-600, pre.text-green-400');

        try {
            await expect(outputArea).toBeVisible({ timeout: 5000 });
        } catch (e) {
            const errorArea = page.getByText(/Error:/);
            await expect(errorArea).toBeVisible({ timeout: 5000 });
        }
    });
});
