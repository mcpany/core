/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import { seedServices, seedUser, cleanupServices, cleanupUser } from './test-data';

test.describe('Tool Replay Feature', () => {

    test.beforeEach(async ({ request, page }) => {
        // Seed data
        await seedServices(request);
        await seedUser(request);

        // Login
        await page.goto('/login');
        await page.fill('input[name="username"]', 'admin');
        await page.fill('input[name="password"]', 'password');
        await page.click('button[type="submit"]');
        await expect(page).toHaveURL('/');
    });

    test.afterEach(async ({ request }) => {
        await cleanupServices(request);
        await cleanupUser(request);
    });

    test('should replay a tool execution from history', async ({ page }) => {
        // 1. Navigate to tools
        await page.goto('/tools');
        await expect(page.locator('h2')).toContainText('Tools');

        // 2. Find and inspect a tool (e.g. calculator from seedServices)
        // seedServices adds "Math" service with "calculator" tool
        // Assuming "calculator" is visible.
        await expect(page.locator('text=calculator')).toBeVisible();

        // Click inspect button for the row containing "calculator"
        const row = page.locator('tr', { hasText: 'calculator' });
        await row.getByRole('button', { name: 'Inspect' }).click();

        // 3. Verify Inspector is open
        await expect(page.getByRole('dialog')).toBeVisible();
        await expect(page.getByRole('dialog')).toContainText('calculator');

        // 4. Enter arguments and execute
        // By default we are in "Test & Execute" tab.
        // Enter JSON args
        const args = JSON.stringify({ a: 10, b: 20 });
        await page.locator('textarea#args').fill(args);

        // Click Execute
        await page.getByRole('button', { name: 'Execute' }).click();

        // 5. Wait for execution to finish (it might fail or succeed, but we expect output or error)
        // Since the backend is dummy, it will likely return an error or timeout.
        // We wait for output to appear.
        await expect(page.locator('text=Result')).toBeVisible({ timeout: 10000 });
        // Or wait for the error message
        // await expect(page.locator('text=Error')).toBeVisible();

        // 6. Switch to "Performance & Analytics" tab
        await page.getByRole('tab', { name: 'Performance & Analytics' }).click();

        // 7. Verify the execution appears in "Recent Timeline"
        // We look for the arguments we sent
        // The list item displays arguments truncated, but should contain "10" or "20"
        // Wait a bit for metrics to refresh (the component has a setTimeout(fetchMetrics, 500))
        await page.waitForTimeout(1000);
        await page.getByRole('button', { name: 'Refresh' }).click();

        // Check if the log entry is there. It should have the args.
        // We look for the text of arguments.
        await expect(page.locator('div').filter({ hasText: '"a": 10' }).first()).toBeVisible();

        // 8. Click "Replay" button
        // The replay button is an icon button with title "Replay Execution"
        await page.getByRole('button', { name: 'Replay Execution' }).first().click();

        // 9. Verify that view switches back to "Test & Execute"
        // The active tab trigger should be "Test & Execute"
        await expect(page.getByRole('tab', { name: 'Test & Execute' })).toHaveAttribute('data-state', 'active');

        // 10. Verify arguments are populated
        // The textarea should contain the arguments
        await expect(page.locator('textarea#args')).toHaveValue(expect.stringContaining('"a": 10'));
    });
});
