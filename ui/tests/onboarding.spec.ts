/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import { seedServices, cleanupServices, seedUser, cleanupUser } from './e2e/test-data';

test.describe('Onboarding Flow', () => {
    test.describe.configure({ mode: 'serial' });

    test.beforeEach(async ({ request, page }) => {
        await cleanupServices(request); // Ensure 0 services to trigger wizard
        await seedUser(request, "onboarding-user");

        await page.goto('/login');
        await page.fill('input[name="username"]', 'onboarding-user');
        await page.fill('input[name="password"]', 'password');
        await page.click('button[type="submit"]');
        await expect(page).toHaveURL('/', { timeout: 15000 });
    });

    test.afterEach(async ({ request }) => {
        await cleanupServices(request);
        // Also cleanup the service created in the test if generic cleanup doesn't catch it
        // cleanupServices removes specific IDs.
        // My Echo Service might have a generated ID or name-based ID.
        // We can try to delete by name "My Echo Service".
        try {
            const context = await request.newContext();
            await context.delete('/api/v1/services/My Echo Service', {
                headers: { 'X-API-Key': 'test-token' }
            });
        } catch (e) {
            // ignore
        }
        await cleanupUser(request, "onboarding-user");
    });

    test('shows welcome wizard when no services exist', async ({ page }) => {
        await expect(page.getByText('Welcome to MCP Any')).toBeVisible({ timeout: 10000 });
        await expect(page.getByRole('button', { name: 'Get Started' })).toBeVisible();
    });

    test('shows dashboard when services exist', async ({ page, request }) => {
        // Seed services manually to trigger Dashboard state
        await seedServices(request);
        await page.reload();

        await expect(page.getByRole('heading', { name: 'Dashboard' })).toBeVisible({ timeout: 10000 });
        await expect(page.getByText('Welcome to MCP Any')).not.toBeVisible();
    });

    test('wizard progression and real registration', async ({ page }) => {
        // Start wizard
        await expect(page.getByText('Welcome to MCP Any')).toBeVisible({ timeout: 10000 });
        await page.getByRole('button', { name: 'Get Started' }).click();

        // Register Step
        await expect(page.getByText('Step 1: Register a Service')).toBeVisible();
        await page.getByRole('button', { name: 'Register New Service' }).click();

        // Dialog
        await expect(page.getByRole('dialog')).toBeVisible();

        // Select "Custom Service" template to proceed to form
        await page.getByText('Custom Service').click();

        // Wait for form to appear
        await expect(page.getByLabel('Service Name')).toBeVisible();
        await page.getByLabel('Service Name').fill('My Echo Service');

        // Select Command Line type
        // Use css selector for the select trigger to be safe
        await page.locator('button[role="combobox"]').click();
        await page.getByRole('option', { name: 'Command Line' }).click();

        // Use full path to echo to be safe in minimal containers
        await page.getByLabel('Command').fill('/bin/echo');

        // Register Service
        await page.getByRole('button', { name: 'Register Service' }).click();

        // Wait for EITHER success toast OR transition to Step 2.
        // The toast might be transient or missed by Playwright, but reaching Step 2 confirms success.
        // If registration failed, we would see "Registration Failed" or stay on Step 1.
        await Promise.any([
            expect(page.getByText('Service Registered')).toBeVisible({ timeout: 10000 }),
            expect(page.getByText('Step 2: Connect AI Client')).toBeVisible({ timeout: 30000 })
        ]).catch(async (e) => {
            // If both fail, log diagnostics
            console.log('Registration verification failed.');
            if (await page.getByText('Registration Failed').isVisible()) {
                const description = await page.locator('ol > li').filter({ hasText: 'Registration Failed' }).textContent();
                console.log('Registration Failed Error:', description);
            }
            console.log('Current URL:', page.url());
            throw e;
        });

        // Ensure we are firmly on Step 2
        await expect(page.getByText('Step 2: Connect AI Client')).toBeVisible();

        // Complete
        await page.getByRole('button', { name: 'Go to Dashboard' }).click();

        // Dashboard should be visible
        await expect(page.getByRole('heading', { name: 'Dashboard' })).toBeVisible({ timeout: 10000 });

        // Verify service exists in list
        await expect(page.getByText('My Echo Service')).toBeVisible();
    });
});
