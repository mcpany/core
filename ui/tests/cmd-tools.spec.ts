/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import { seedUser, cleanupUser } from './e2e/test-data';

test.describe('Command Line Tools', () => {
    test.describe.configure({ mode: 'serial' });

    test.beforeEach(async ({ request, page }) => {
        await seedUser(request, "e2e-cmd-admin");

        // Login first
        await page.goto('/login');
        await page.fill('input[name="username"]', 'e2e-cmd-admin');
        await page.fill('input[name="password"]', 'password');
        await page.click('button[type="submit"]');
        await expect(page).toHaveURL('/', { timeout: 15000 });
    });

    test.afterEach(async ({ request }) => {
        await cleanupUser(request, "e2e-cmd-admin");
    });

    test('should create a CMD service and add a tool', async ({ page }) => {
        // 1. Create CMD Service
        await page.goto('/upstream-services');
        await page.getByRole('button', { name: 'Add Service' }).click();

        await page.waitForTimeout(2000); // Wait for sheet animation
        await expect(page.getByRole('dialog')).toBeVisible();

        // Select "Custom Service" template
        await page.getByText('Custom Service').click();

        await page.waitForSelector('input#name', { state: 'visible', timeout: 30000 });
        await page.fill('input#name', 'cmd-test-service');

        // Select CMD Service Type
        // The default might be HTTP, so we need to switch
        await page.locator('button[role="combobox"]').first().click();
        await page.getByRole('option', { name: 'Command Line' }).click();

        await page.getByRole('tab', { name: 'Connection' }).click();
        await page.fill('input#command', 'echo');

        // Save
        await page.click('button:has-text("Save Changes")');
        await expect(page.getByText('Service Created')).toBeVisible();

        // 2. Add a Tool
        const row = page.locator('tr').filter({ hasText: 'cmd-test-service' });
        await row.getByRole('button', { name: 'Edit' }).click();
        await expect(page.getByRole('dialog')).toBeVisible();

        // Go to Tools tab
        await page.getByRole('tab', { name: 'Tools' }).click();

        // Add Tool
        await page.getByRole('button', { name: 'Add Tool' }).click();

        // Configure Tool
        await page.fill('input#tool-name', 'echo_hello');
        await page.fill('input#tool-description', 'Echoes hello');

        // Add Argument
        await page.click('button:has-text("Add Argument")');
        await page.locator('input[placeholder^="Argument"]').first().fill('hello');

        // Add Parameter
        await page.click('button:has-text("Add Parameter")');
        await page.locator('input[id^="param-name"]').fill('name');

        // Add another argument using parameter
        await page.click('button:has-text("Add Argument")');
        await page.locator('input[placeholder^="Argument"]').nth(1).fill('{{name}}');

        // Close Tool Sheet (Press Escape)
        await page.keyboard.press('Escape');

        // Now Save the Service
        await page.click('button:has-text("Save Changes")');
        await expect(page.getByText('Service Updated')).toBeVisible();

        // 3. Verify Persistence
        await page.reload();

        const row2 = page.locator('tr').filter({ hasText: 'cmd-test-service' });
        await row2.getByRole('button', { name: 'Edit' }).click();
        await page.click('button[role="tab"]:has-text("Tools")');

        await expect(page.getByText('echo_hello')).toBeVisible();
        // The summary in list might vary in format, but checking for tool name is good start.
        await expect(page.getByText('echo_hello')).toBeVisible();
    });
});
