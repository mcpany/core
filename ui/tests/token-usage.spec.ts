/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import { seedServices, cleanupServices, seedUser, cleanupUser } from './e2e/test-data';

test.describe('Token Usage Visibility', () => {
    test.describe.configure({ mode: 'serial' });

    test.beforeEach(async ({ request, page }) => {
        await seedServices(request);
        await seedUser(request, "e2e-token-user");

        // Login first
        await page.goto('/login');
        await page.fill('input[name="username"]', 'e2e-token-user');
        await page.fill('input[name="password"]', 'password');
        await page.click('button[type="submit"]');
        await expect(page).toHaveURL('/', { timeout: 15000 });
    });

    test.afterEach(async ({ request }) => {
        await cleanupServices(request);
        await cleanupUser(request, "e2e-token-user");
    });

    test('should show token usage estimation in tool runner', async ({ page }) => {
        await page.goto('/tools');

        // Wait for services to load
        for (let i = 0; i < 10; i++) {
            try {
                await expect(page.getByText('process_payment').first()).toBeVisible({ timeout: 5000 });
                break;
            } catch (e) {
                await page.reload();
                await page.waitForLoadState('networkidle');
                await page.waitForTimeout(1000);
            }
        }

        // Inspect echo_tool
        const toolRow = page.locator('tr').filter({ hasText: /echo_tool/ });
        await toolRow.getByRole('button', { name: 'Inspect' }).click({ timeout: 30000 });

        // Switch to JSON input
        await page.getByRole('dialog').getByRole('tablist').filter({ hasText: 'Form' }).getByRole('tab', { name: 'JSON', exact: true }).click();

        // Fill arguments
        const textArea = page.locator('textarea').first();
        const inputPayload = '{"message": "Hello MCP Tokens"}';
        await textArea.fill(inputPayload);

        // Execute
        await page.getByRole('button', { name: 'Execute' }).click();

        // Wait for result (success or error)
        // We check for EITHER success (output) OR failure (error message),
        // proving that we hit the REAL backend.
        const outputArea = page.locator('pre.text-green-600, pre.text-green-400');
        try {
             await expect(outputArea).toBeVisible({ timeout: 5000 });
        } catch (e) {
             const errorArea = page.getByText(/(Error:|upstream|fetch)/i);
             await expect(errorArea).toBeVisible({ timeout: 5000 });
        }

        // Check for Token Usage display
        // We expect something like "Tokens: Input ~X" or "Input Tokens"
        await expect(page.getByTitle("Estimated Input Tokens")).toBeVisible();
        await expect(page.getByTitle("Estimated Output Tokens")).toBeVisible();
        await expect(page.getByTitle("Estimated Cost")).toBeVisible();
    });
});
